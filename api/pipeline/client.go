package pipeline

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/unchartedsoftware/plog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Client provides facilities for managing GPRC pipeline requests.  Requests are
// isssued and a context object containing rx channels is returned to the caller for consumption
// of results.  The context for running requests can also be fetched, along with their buffered
// results.  Spawning a grpc.ClientConn per RPC call is not considered good practice - the system
// is designed such that multiple go routines make RPC calls to a single shared client, and synch
// is managed internally.
type Client struct {
	pendingRequests   map[uuid.UUID]*RequestContext
	completedRequests map[uuid.UUID]*RequestContext
	results           map[uuid.UUID][]interface{}
	reqMutex          sync.Mutex
	client            PipelineComputeClient
	conn              *grpc.ClientConn
	downstreamMutex   sync.Mutex
	downstream        map[uuid.UUID][]RequestResult
}

// RequestResult provides a channel for receiving results and another for receiving
// errors.
type RequestResult struct {
	Results chan interface{}
	Errors  chan error
}

// NewClient creates a new pipline reuqest dispatcher instance.  This will establish
// the connection to the pipeline server or return an error on fail
func NewClient(serverAddr string) (*Client, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to %s", serverAddr)
	}
	log.Infof("connected to %s", serverAddr)

	client := Client{}

	client.pendingRequests = make(map[uuid.UUID]*RequestContext)
	client.completedRequests = make(map[uuid.UUID]*RequestContext)
	client.results = make(map[uuid.UUID][]interface{})

	client.client = NewPipelineComputeClient(conn)
	client.conn = conn

	client.downstream = make(map[uuid.UUID][]RequestResult)

	return &client, nil
}

// Close the connection to the pipeline service
func (r *Client) Close() {
	log.Infof("client connection closed")
	r.conn.Close()
}

// IsRequestAttachable determines if there is a running request similar to that supplied by the
// user that we can attach to.  Internal logic will identify requests types that can't be re-used
// by their nature, so the caller can pass any request type through.
func (r *Client) IsRequestAttachable(info *RequestInfo) (uuid.UUID, bool) {
	r.reqMutex.Lock()
	for _, v := range r.pendingRequests {
		if info.RequestHash == v.RequestHash {
			r.reqMutex.Unlock()
			return v.RequestID, true
		}
	}
	r.reqMutex.Unlock()
	return uuid.Nil, false
}

// Dispatch sends a request to the compute client and returns the request ID to the caller
func (r *Client) Dispatch(ctx context.Context, request RequestFunc) uuid.UUID {
	// execute the request and store the context in the pending requests map
	requestCtx := request(&ctx, &r.client)

	r.reqMutex.Lock()
	r.pendingRequests[requestCtx.RequestID] = requestCtx
	r.reqMutex.Unlock()

	// Store results locally and forward results and errors downstream for processing.  If
	// the source channels are closed, closed the downstream channels.
	go func() {
		for {
			select {
			case err, ok := <-requestCtx.Errors:
				if !ok {
					requestCtx.Errors = nil
				} else {
					// broadcast the error downstream
					log.Error(err)
					r.downstreamMutex.Lock()
					for _, downstream := range r.downstream[requestCtx.RequestID] {
						downstream.Errors <- err
					}
					r.downstreamMutex.Unlock()
				}
			case result, ok := <-requestCtx.Results:
				if !ok {
					requestCtx.Results = nil
				} else {
					// put the results in the buffer
					r.reqMutex.Lock()
					if _, ok := r.results[requestCtx.RequestID]; !ok {
						r.results[requestCtx.RequestID] = make([]interface{}, 0)
					}
					r.results[requestCtx.RequestID] = append(r.results[requestCtx.RequestID], result)
					r.reqMutex.Unlock()

					// broadcast the result downstream
					r.downstreamMutex.Lock()
					for _, downstream := range r.downstream[requestCtx.RequestID] {
						downstream.Results <- result
					}
					r.downstreamMutex.Unlock()
				}
			}
			if requestCtx.Errors == nil && requestCtx.Results == nil {
				break
			}
		}
	}()
	return requestCtx.RequestID
}

// Attach to an already running request.  This provides the caller with channels to handle
// request data and errors.
func (r *Client) Attach(requestID uuid.UUID) (*RequestResult, []interface{}) {
	r.reqMutex.Lock()
	if _, ok := r.pendingRequests[requestID]; ok {
		results := r.results[requestID]
		resultsCopy := make([]interface{}, len(results))
		copy(resultsCopy, results)
		r.reqMutex.Unlock()

		requestResult := RequestResult{make(chan interface{}), make(chan error)}
		r.downstreamMutex.Lock()
		r.downstream[requestID] = append(r.downstream[requestID], requestResult)
		r.downstreamMutex.Unlock()

		return &requestResult, resultsCopy
	}
	log.Warnf("can't attach - no running request with id %s", requestID)
	r.reqMutex.Unlock()
	return nil, nil
}
