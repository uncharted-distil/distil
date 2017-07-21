package pipeline

import (
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/unchartedsoftware/plog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Client provides facilities for managing GPRC pipeline requests. Requests are
// isssued and a context object containing rx channels is returned to the caller for consumption
// of results. The context for running requests can also be fetched, along with their buffered
// results. Spawning a grpc.ClientConn per RPC call is not considered good practice - the system
// is designed such that multiple go routines make RPC calls to a single shared client, and synch
// is managed internally.
type Client struct {
	pendingRequests   map[uuid.UUID]*RequestContext
	completedRequests map[uuid.UUID]*RequestContext
	results           map[uuid.UUID][]*proto.Message
	mu                *sync.Mutex
	client            PipelineComputeClient
	conn              *grpc.ClientConn
	downstream        map[uuid.UUID][]*ResultProxy
}

// ResultProxy provides a channel for receiving results and another for receiving
// errors. This the main conduit for comms between the client and downstream handlers
// that are receviing request results.
type ResultProxy struct {
	Results chan *proto.Message
	Errors  chan error
	Done    chan struct{}
}

// NewClient creates a new pipline reuqest dispatcher instance. This will establish
// the connection to the pipeline server or return an error on fail
func NewClient(serverAddr string) (*Client, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to %s", serverAddr)
	}
	log.Infof("connected to %s", serverAddr)

	client := Client{}
	client.mu = &sync.Mutex{}
	client.pendingRequests = make(map[uuid.UUID]*RequestContext)
	client.completedRequests = make(map[uuid.UUID]*RequestContext)
	client.downstream = make(map[uuid.UUID][]*ResultProxy)
	client.results = make(map[uuid.UUID][]*proto.Message)
	client.client = NewPipelineComputeClient(conn)
	client.conn = conn
	return &client, nil
}

// Close the connection to the pipeline service
func (c *Client) Close() {
	log.Infof("client connection closed")
	c.conn.Close()
}

// GetExistingUUIDs will return the uuids for all pending and completed requests.
func (c *Client) GetExistingUUIDs() []uuid.UUID {
	c.mu.Lock()
	defer c.mu.Unlock()

	var uuids []uuid.UUID
	// add pending uuids
	for _, req := range c.pendingRequests {
		uuids = append(uuids, req.RequestID)
	}
	// add completed uuids
	for _, req := range c.completedRequests {
		uuids = append(uuids, req.RequestID)
	}
	return uuids
}

// Get will return a result proxy for the provided uuid.
func (c *Client) Get(requestID uuid.UUID) (*ResultProxy, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.attachToExistingRequest(requestID)
}

// GetOrDispatch will either get an existing result proxy, or dispatch a new
// request and return its result proxy.
func (c *Client) GetOrDispatch(ctx context.Context, info *RequestInfo) (*ResultProxy, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// check pending requests
	for _, req := range c.pendingRequests {
		if info.RequestHash == req.RequestHash {
			return c.attachToExistingRequest(req.RequestID)
		}
	}
	// check completed requests
	for _, req := range c.completedRequests {
		if info.RequestHash == req.RequestHash {
			return c.attachToExistingRequest(req.RequestID)
		}
	}
	// no request we could re-use, dispatch a new one and attach
	requestID := c.dispatchRequest(ctx, info.RequestFunc)
	return c.attachToExistingRequest(requestID)
}

func (c *Client) proxyError(req *RequestContext, err error) {
	c.mu.Lock()
	for _, downstream := range c.downstream[req.RequestID] {
		downstream.Errors <- err
	}
	c.mu.Unlock()
}

func (c *Client) proxyResult(req *RequestContext, result proto.Message) {
	c.mu.Lock()
	// ensure result slice exists
	_, ok := c.results[req.RequestID]
	if !ok {
		c.results[req.RequestID] = make([]*proto.Message, 0)
	}
	// append to result slice
	c.results[req.RequestID] = append(c.results[req.RequestID], &result)
	// broadcast the result downstream
	for _, downstream := range c.downstream[req.RequestID] {
		downstream.Results <- &result
	}
	c.mu.Unlock()
}

func (c *Client) proxyDone(req *RequestContext) {
	// notify downstream routines that request has finished processing
	c.mu.Lock()
	for _, downstream := range c.downstream[req.RequestID] {
		downstream.Done <- struct{}{}
		close(downstream.Results)
		close(downstream.Errors)
	}
	// request is finished so don't need to track any more
	delete(c.downstream, req.RequestID)
	delete(c.pendingRequests, req.RequestID)
	c.completedRequests[req.RequestID] = req
	c.mu.Unlock()
}

func (c *Client) dispatchRequest(ctx context.Context, request RequestFunc) uuid.UUID {
	// NOTE: this method is not thread safe and assumes locked access

	// execute the request and store the context in the pending requests map
	req := request(&ctx, &c.client)

	// store as pending
	c.pendingRequests[req.RequestID] = req

	// Store results locally and forward results and errors downstream for processing.  If
	// the source channels are closed we nil them out and close down the downstream channels.
	go func() {
		done := false
		for !done {
			select {
			case err := <-req.Errors:
				// broadcast the error downstream
				c.proxyError(req, err)

			case result := <-req.Results:
				// put the results in the buffer
				c.proxyResult(req, *result)

			case <-req.Done:
				// notify downstream routines that request has finished processing
				c.proxyDone(req)
				// flag as done
				done = true
			}
		}
	}()
	return req.RequestID
}

func (c *Client) getResultsImmutable(requestID uuid.UUID) []*proto.Message {
	// NOTE: this method is not thread safe and assumes locked access

	// make a copy of the results list so we can share - results themselves
	// are immutable
	results := c.results[requestID]
	copied := make([]*proto.Message, len(results))
	copy(copied, results)
	return copied
}

func (c *Client) attachToExistingRequest(requestID uuid.UUID) (*ResultProxy, error) {
	// NOTE: this method is not thread safe and assumes locked access

	// check if pending
	_, ok := c.pendingRequests[requestID]
	if ok {
		// get copy of results
		results := c.getResultsImmutable(requestID)

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
		_, ok := c.downstream[requestID]
		if !ok {
			c.downstream[requestID] = make([]*ResultProxy, 0)
		}
		c.downstream[requestID] = append(c.downstream[requestID], proxy)
		return proxy, nil
	}

	_, ok = c.completedRequests[requestID]
	if ok {
		// get copy of results
		results := c.getResultsImmutable(requestID)
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
