package pipeline

import (
	"hash/fnv"
	"io"

	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/unchartedsoftware/plog"
	"golang.org/x/net/context"
)

// RequestFunc defines a standardized pipeline request execution function.
type RequestFunc func(ctx *context.Context, client *PipelineComputeClient) *RequestContext

// RequestInfo provides a unique ID and hash for a request to be send, along with a function that
// can be called to initiate the request as an RPC call.
// TODO: Better name ?
// TODO: Break out ID and Hash into a separate structure that is embedded in RequestInfo and
// RequestStruct, define interfaces.
type RequestInfo struct {
	RequestID   uuid.UUID
	RequestHash uint64
	RequestFunc RequestFunc
}

// RequestContext provides information about a in-progress or completed pipeline request,
// as well as channels for handling results when the request is live.
// TODO: Better name ?
type RequestContext struct {
	Context     *context.Context
	RequestID   uuid.UUID
	RequestHash uint64
	Request     interface{}
	Results     chan interface{}
	Errors      chan error
	Done        chan struct{}
}

// NewRequestContext creates a request context with default channels.
func NewRequestContext(ctx *context.Context, requestID uuid.UUID, hash uint64, request interface{}) *RequestContext {
	return &RequestContext{
		ctx,
		requestID,
		hash,
		request,
		make(chan interface{}),
		make(chan error),
		make(chan struct{}),
	}
}

// GeneratePipelineCreateRequest creates a PipelineCreateRequest that will initiate pipeline creation on the server and
// and handle a stream of PipelineCreateResult objects that are returned as work is completed.
func GeneratePipelineCreateRequest(request *PipelineCreateRequest) *RequestInfo {
	requestID := uuid.NewV1()
	hash := structHash(request, requestID.Bytes())
	requestFunc := func(ctx *context.Context, client *PipelineComputeClient) *RequestContext {
		requestCtx := NewRequestContext(ctx, requestID, hash, request)

		go func() {
			defer func() {
				close(requestCtx.Results)
				close(requestCtx.Errors)
			}()
			// iniitate pipeline creation
			stream, err := (*client).CreatePipelines(*ctx, request)
			if err != nil {
				log.Error(err)
				requestCtx.Errors <- errors.Wrap(err, "failed to initiate create pipeline request")
				return
			}
			// handle the result stream
			for {
				result, err := stream.Recv()
				if err == io.EOF {
					requestCtx.Done <- struct{}{}
					break
				} else if err != nil {
					log.Error(err)
					requestCtx.Errors <- errors.Wrap(err, "failed to process create pipeline result")
					requestCtx.Done <- struct{}{}
					break
				}
				requestCtx.Results <- result
			}
			return
		}()
		return requestCtx
	}
	return &RequestInfo{requestID, hash, requestFunc}
}

// GeneratePipelineExecuteRequest creates a PipelineExecuteRequest that will execute a pipeline on the server and
// and handle a stream of PipelineExecuteResult objects that are returned as work is completed.
func GeneratePipelineExecuteRequest(request *PipelineExecuteRequest) *RequestInfo {
	requestID := uuid.NewV1()
	hash := structHash(request, requestID.Bytes())
	requestFunc := func(ctx *context.Context, client *PipelineComputeClient) *RequestContext {
		requestCtx := NewRequestContext(ctx, requestID, hash, request)

		go func() {
			defer func() {
				close(requestCtx.Results)
				close(requestCtx.Errors)
			}()
			// initiate pipeline execution
			stream, err := (*client).ExecutePipeline(*ctx, request)
			if err != nil {
				log.Error(err)
				requestCtx.Errors <- errors.Wrap(err, "failed to initiate execute pipeline request")
				return
			}
			// handle the result stream
			for {
				result, err := stream.Recv()
				if err == io.EOF {
					requestCtx.Done <- struct{}{}
					break
				}
				if err != nil {
					log.Error(err)
					requestCtx.Errors <- errors.Wrap(err, "failed to process execute pipeline result")
					requestCtx.Done <- struct{}{}
					break
				}
				requestCtx.Results <- result
			}
			return
		}()
		return requestCtx
	}
	return &RequestInfo{requestID, hash, requestFunc}
}

// GenerateStartSessionRequest creates a session start request that will return a unique session ID
// to the caller. This ID is then assigned to subsquent pipeline calls via the session context field.
func GenerateStartSessionRequest() *RequestInfo {
	requestID := uuid.NewV1()
	hash := hash(requestID.Bytes())
	requestFunc := func(ctx *context.Context, client *PipelineComputeClient) *RequestContext {
		requestCtx := NewRequestContext(ctx, requestID, hash, nil)

		go func() {
			defer func() {
				close(requestCtx.Results)
				close(requestCtx.Errors)
			}()
			result, err := (*client).StartSession(*ctx, &SessionRequest{})
			if err != nil {
				requestCtx.Errors <- errors.Wrap(err, "failed to initiate start session request")
				requestCtx.Done <- struct{}{}
				return
			}
			requestCtx.Results <- result
			requestCtx.Done <- struct{}{}
			return
		}()
		return requestCtx
	}
	return &RequestInfo{requestID, hash, requestFunc}
}

// GenerateEndSessionRequest creates a session end request that will mark a session as closed. The session
// is not available for further pipeline requests once called.
func GenerateEndSessionRequest(sessionID string) *RequestInfo {
	requestID := uuid.NewV1()
	hash := hash(requestID.Bytes())
	requestFunc := func(ctx *context.Context, client *PipelineComputeClient) *RequestContext {
		sessionCtx := SessionContext{sessionID}
		requestCtx := NewRequestContext(ctx, requestID, hash, sessionCtx)
		go func() {
			defer func() {
				close(requestCtx.Results)
				close(requestCtx.Errors)
			}()
			result, err := (*client).EndSession(*ctx, &SessionContext{sessionID})
			if err != nil {
				requestCtx.Errors <- errors.Wrap(err, "failed to initiate end session request")
				requestCtx.Done <- struct{}{}
				return
			}
			requestCtx.Results <- result
			requestCtx.Done <- struct{}{}
			return
		}()
		return requestCtx
	}
	return &RequestInfo{requestID, hash, requestFunc}
}

// HashInclude satisifies the Includable interface from hashstructure package, and  allows
// for the context field to be skipped when generating a hash for the PiplineCreateRequest
// struct.
func (PipelineCreateRequest) HashInclude(field string, v interface{}) (bool, error) {
	return field != "Context", nil
}

// HashInclude satisifies the Includable interface from hashstructure package, and  allows
// for the context field to be skipped when generating a hash for the PiplineExecuteRequest
// struct.
func (PipelineExecuteRequest) HashInclude(field string, v interface{}) (bool, error) {
	return field != "Context", nil
}

func hash(b []byte) uint64 {
	hash := fnv.New64a()
	hash.Write(b)
	return hash.Sum64()
}

// structHash generates a hash code from an input value, and falls back on the supplied
// default if there is a problem generating the hash.
func structHash(s interface{}, def []byte) uint64 {
	requestHash, err := hashstructure.Hash(s, nil)
	if err != nil {
		log.Errorf("using ID due to hash failure - %s", err)
		requestHash = hash(def)
	}
	return requestHash
}
