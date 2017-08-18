package pipeline

import (
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

// Session provides facilities for managing GPRC pipeline sessions.
type Session struct {
	ID                string
	client            PipelineComputeClient
	pendingRequests   map[uuid.UUID]*RequestContext
	completedRequests map[uuid.UUID]*RequestContext
	results           map[uuid.UUID][]*proto.Message
	downstream        map[uuid.UUID][]*ResultProxy
	mu                *sync.Mutex
}

// NewSession instantiates and returns a new session.
func NewSession(id string, client PipelineComputeClient) *Session {
	return &Session{
		ID:                id,
		client:            client,
		mu:                &sync.Mutex{},
		pendingRequests:   make(map[uuid.UUID]*RequestContext),
		completedRequests: make(map[uuid.UUID]*RequestContext),
		downstream:        make(map[uuid.UUID][]*ResultProxy),
		results:           make(map[uuid.UUID][]*proto.Message),
	}
}

// ResultProxy provides a channel for receiving results and another for receiving
// errors. This the main conduit for comms between the client and downstream handlers
// that are receviing request results.
type ResultProxy struct {
	Results chan *proto.Message
	Errors  chan error
	Done    chan struct{}
}

// GetExistingUUIDs will return the uuids for all pending and completed requests.
func (s *Session) GetExistingUUIDs() []uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()

	var uuids []uuid.UUID
	// add pending uuids
	for _, req := range s.pendingRequests {
		uuids = append(uuids, req.RequestID)
	}
	// add completed uuids
	for _, req := range s.completedRequests {
		uuids = append(uuids, req.RequestID)
	}
	return uuids
}

// Get will return a result proxy for the provided uuid.
func (s *Session) Get(requestID uuid.UUID) (*ResultProxy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.attachToExistingRequest(requestID)
}

// GetOrDispatch will either get an existing result proxy, or dispatch a new
// request and return its result proxy.
func (s *Session) GetOrDispatch(ctx context.Context, info *RequestInfo) (*ResultProxy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// check pending requests
	for _, req := range s.pendingRequests {
		if info.RequestHash == req.RequestHash {
			return s.attachToExistingRequest(req.RequestID)
		}
	}
	// check completed requests
	for _, req := range s.completedRequests {
		if info.RequestHash == req.RequestHash {
			return s.attachToExistingRequest(req.RequestID)
		}
	}
	// no request we could re-use, dispatch a new one and attach
	requestID := s.dispatchRequest(ctx, info.RequestFunc)
	return s.attachToExistingRequest(requestID)
}

func (s *Session) proxyError(req *RequestContext, err error) {
	s.mu.Lock()
	for _, downstream := range s.downstream[req.RequestID] {
		downstream.Errors <- err
	}
	s.mu.Unlock()
}

func (s *Session) proxyResult(req *RequestContext, result proto.Message) {
	s.mu.Lock()
	// ensure result slice exists
	_, ok := s.results[req.RequestID]
	if !ok {
		s.results[req.RequestID] = make([]*proto.Message, 0)
	}
	// append to result slice
	s.results[req.RequestID] = append(s.results[req.RequestID], &result)
	// broadcast the result downstream
	for _, downstream := range s.downstream[req.RequestID] {
		downstream.Results <- &result
	}
	s.mu.Unlock()
}

func (s *Session) proxyDone(req *RequestContext) {
	// notify downstream routines that request has finished processing
	s.mu.Lock()
	for _, downstream := range s.downstream[req.RequestID] {
		downstream.Done <- struct{}{}
		close(downstream.Results)
		close(downstream.Errors)
	}
	// request is finished so don't need to track any more
	delete(s.downstream, req.RequestID)
	delete(s.pendingRequests, req.RequestID)
	s.completedRequests[req.RequestID] = req
	s.mu.Unlock()
}

func (s *Session) dispatchRequest(ctx context.Context, request RequestFunc) uuid.UUID {
	// NOTE: this method is not thread safe and assumes locked access

	// execute the request and store the context in the pending requests map
	req := request(&ctx, &s.client)

	// store as pending
	s.pendingRequests[req.RequestID] = req

	// Store results locally and forward results and errors downstream for processing.  If
	// the source channels are closed we nil them out and close down the downstream channels.
	go func() {
		done := false
		for !done {
			select {
			case err := <-req.Errors:
				// broadcast the error downstream
				s.proxyError(req, err)

			case result := <-req.Results:
				// put the results in the buffer
				s.proxyResult(req, *result)

			case <-req.Done:
				// notify downstream routines that request has finished processing
				s.proxyDone(req)
				// flag as done
				done = true
			}
		}
	}()
	return req.RequestID
}

func (s *Session) getResultsImmutable(requestID uuid.UUID) []*proto.Message {
	// NOTE: this method is not thread safe and assumes locked access

	// make a copy of the results list so we can share - results themselves
	// are immutable
	results := s.results[requestID]
	copied := make([]*proto.Message, len(results))
	copy(copied, results)
	return copied
}

func (s *Session) attachToExistingRequest(requestID uuid.UUID) (*ResultProxy, error) {
	// NOTE: this method is not thread safe and assumes locked access

	// check if pending
	_, ok := s.pendingRequests[requestID]
	if ok {
		// get copy of results
		results := s.getResultsImmutable(requestID)

		// create a result proxy object for communicating result and request
		// state to downstream consumer
		proxy := &ResultProxy{
			Results: make(chan *proto.Message, len(results)),
			Errors:  make(chan error),
			Done:    make(chan struct{}),
		}

		// write to buffered results
		for _, result := range results {
			proxy.Results <- result
		}

		// add to downstream
		_, ok := s.downstream[requestID]
		if !ok {
			s.downstream[requestID] = make([]*ResultProxy, 0)
		}
		s.downstream[requestID] = append(s.downstream[requestID], proxy)
		return proxy, nil
	}

	_, ok = s.completedRequests[requestID]
	if ok {
		// get copy of results
		results := s.getResultsImmutable(requestID)
		// create a result proxy object for communicating result and request
		// state to downstream consumer
		proxy := &ResultProxy{
			Results: make(chan *proto.Message),
			Errors:  make(chan error),
			Done:    make(chan struct{}),
		}
		// write to result channel, block so that done channel always comes
		// last
		go func() {
			// write results
			for _, result := range results {
				proxy.Results <- result
			}
			// write to done
			proxy.Done <- struct{}{}
		}()

		return proxy, nil
	}

	return nil, errors.Errorf("can't attach - no running request with id %s", requestID)
}
