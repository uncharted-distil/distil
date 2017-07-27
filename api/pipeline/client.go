package pipeline

import (
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/middleware"
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
	sessions map[string]*Session
	client   PipelineComputeClient
	conn     *grpc.ClientConn
	mu       *sync.Mutex
}

// NewClient creates a new pipline reuqest dispatcher instance. This will establish
// the connection to the pipeline server or return an error on fail
func NewClient(serverAddr string) (*Client, error) {
	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(middleware.GenerateUnaryClientInterceptor()),
		grpc.WithStreamInterceptor(middleware.GenerateStreamClientInterceptor()),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to %s", serverAddr)
	}
	log.Infof("connected to %s", serverAddr)

	client := Client{}
	client.mu = &sync.Mutex{}
	client.sessions = make(map[string]*Session)
	client.client = NewPipelineComputeClient(conn)
	client.conn = conn
	return &client, nil
}

// Close the connection to the pipeline service
func (c *Client) Close() {
	log.Infof("client connection closed")
	c.conn.Close()
}

// GetSession returns an existing session struct.
func (c *Client) GetSession(id string) (*Session, bool) {
	// check for session
	c.mu.Lock()
	session, ok := c.sessions[id]
	c.mu.Unlock()
	if !ok {
		return nil, false
	}
	return session, true
}

// StartSession starts a new session.
func (c *Client) StartSession(ctx context.Context) (*Session, error) {
	// create start session request
	req := GenerateStartSessionRequest()
	// execute the request
	results, err := c.dispatchRequestSync(ctx, req.RequestFunc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start pipeline session")
	}
	// create session
	result, ok := (*results[0]).(*Response)
	if !ok {
		return nil, errors.Errorf("unable to start session")
	}
	// create session
	session := NewSession(result.GetContext().GetSessionId(), c.client)
	// store session
	c.mu.Lock()
	c.sessions[session.ID] = session
	c.mu.Unlock()
	return session, nil
}

// EndSession ends a session.
func (c *Client) EndSession(ctx context.Context, id string) error {
	// check for session
	c.mu.Lock()
	_, ok := c.sessions[id]
	c.mu.Unlock()
	if !ok {
		return errors.Errorf("session id `%s` is not recognized", id)
	}
	// create start session request
	req := GenerateEndSessionRequest(id)
	// execute the request
	_, err := c.dispatchRequestSync(ctx, req.RequestFunc)
	if err != nil {
		return errors.Wrap(err, "failed to end pipeline session")
	}
	c.mu.Lock()
	delete(c.sessions, id)
	c.mu.Unlock()
	return nil
}

func (c *Client) dispatchRequestSync(ctx context.Context, request RequestFunc) ([]*proto.Message, error) {
	// execute the start session request
	req := request(&ctx, &c.client)

	// wait until session id is available
	var results []*proto.Message

	done := false
	for !done {
		select {
		case err := <-req.Errors:
			// return err
			return results, err

		case result := <-req.Results:
			// create session
			results = append(results, result)

		case <-req.Done:
			// flag as done
			done = true
		}
	}

	return results, nil
}
