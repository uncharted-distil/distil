package pipeline

import (
	"io"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/unchartedsoftware/plog"
	context "golang.org/x/net/context"
)

// Request defines a standardized pipeline request execution function.
type Request func(ctx *context.Context, client *PipelineComputeClient) *RequestContext

// RequestContext provides information about a in-progress or completed pipeline request,
// as well as channels for handling results when the request is live.
type RequestContext struct {
	Context   *context.Context
	RequestID uuid.UUID
	Request   interface{}
	Results   chan interface{}
	Errors    chan error
}

// GeneratePipelineCreateRequest creates a PipelineCreateRequest that will initiate pipeline creation on the server and
// and handle a stream of PipelineCreateResult objects that are returned as work is completed.
func GeneratePipelineCreateRequest(createReq *PipelineCreateRequest) Request {
	return func(ctx *context.Context, client *PipelineComputeClient) *RequestContext {
		requestID := uuid.NewV1()
		requestContext := RequestContext{ctx, requestID, createReq, make(chan interface{}), make(chan error)}

		go func() {
			defer func() {
				close(requestContext.Results)
				close(requestContext.Errors)
			}()
			// iniitate pipeline creation
			stream, err := (*client).CreatePipelines(*ctx, createReq)
			if err != nil {
				log.Error(err)
				requestContext.Errors <- errors.Wrap(err, "failed to initiate create pipeline request")
				return
			}
			// handle the result stream
			for {
				result, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Error(err)
					requestContext.Errors <- errors.Wrap(err, "failed to process create pipeline result")
					break
				}
				requestContext.Results <- result
			}
			return
		}()
		return &requestContext
	}
}

// GeneratePipelineExecuteRequest creates a PipelineExecuteRequest that will execute a pipeline on the server and
// and handle a stream of PipelineExecuteResult objects that are returned as work is completed.
func GeneratePipelineExecuteRequest(executeReq *PipelineExecuteRequest) Request {
	return func(ctx *context.Context, client *PipelineComputeClient) *RequestContext {
		requestID := uuid.NewV1()
		requestContext := RequestContext{ctx, requestID, executeReq, make(chan interface{}), make(chan error)}

		go func() {
			defer func() {
				close(requestContext.Results)
				close(requestContext.Errors)
			}()
			// initiate pipeline execution
			stream, err := (*client).ExecutePipeline(*ctx, executeReq)
			if err != nil {
				log.Error(err)
				requestContext.Errors <- errors.Wrap(err, "failed to initiate execute pipeline request")
				return
			}
			// handle the result stream
			for {
				result, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Error(err)
					requestContext.Errors <- errors.Wrap(err, "failed to process execute pipeline result")
					break
				}
				requestContext.Results <- result
			}
			return
		}()

		return &requestContext
	}
}

// GenerateStartSessionRequest creates a session start request that will return a unique session ID
// to the caller.  This ID is then assigned to subsquent  pipeline calls via the session context field.
func GenerateStartSessionRequest() Request {
	return func(ctx *context.Context, client *PipelineComputeClient) *RequestContext {
		requestID := uuid.NewV1()
		requestContext := RequestContext{ctx, requestID, nil, make(chan interface{}), make(chan error)}

		go func() {
			defer func() {
				close(requestContext.Results)
				close(requestContext.Errors)
			}()
			result, err := (*client).StartSession(*ctx, &SessionRequest{})
			if err != nil {
				requestContext.Errors <- errors.Wrap(err, "failed to initiate start session request")
			}
			requestContext.Results <- result
			return
		}()
		return &requestContext
	}
}

// GenerateEndSessionRequest creates a session end request that will mark a session as closed.  The session
// is not avialable for further pipeline requests once called.
func GenerateEndSessionRequest(sessionID string) Request {
	return func(ctx *context.Context, client *PipelineComputeClient) *RequestContext {
		sessionContext := SessionContext{sessionID}
		requestID := uuid.NewV1()
		requestContext := RequestContext{ctx, requestID, sessionContext, make(chan interface{}), make(chan error)}

		go func() {
			defer func() {
				close(requestContext.Results)
				close(requestContext.Errors)
			}()
			result, err := (*client).EndSession(*ctx, &SessionContext{sessionID})
			if err != nil {
				log.Error(err)
				requestContext.Errors <- errors.Wrap(err, "failed to initiate end session request")
			}
			requestContext.Results <- result
			return
		}()
		return &requestContext
	}
}
