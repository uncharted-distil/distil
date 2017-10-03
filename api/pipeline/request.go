package pipeline

import (
	"bytes"
	"compress/gzip"
	"hash/fnv"
	"io"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	protobuf "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/unchartedsoftware/plog"
	"golang.org/x/net/context"
)

const (
	userAgent    = "uncharted-distil" // TODO: get version embed into string
	versionUnset = "client-version-unset"
)

var (
	version = versionUnset
)

// APIVersion provides the ta3-ta2 api version as defined in the protobuf def
func APIVersion() string {
	if version == versionUnset {
		version = getAPIVersion()
	}
	return version
}

// RequestFunc defines a standardized pipeline request execution function.
type RequestFunc func(ctx *context.Context, client *CoreClient) *RequestContext

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
	Request     *proto.Message

	Results chan *proto.Message
	Errors  chan error
	Done    chan struct{}
}

// NewRequestContext creates a request context with default channels.
func NewRequestContext(ctx *context.Context, requestID uuid.UUID, hash uint64, request proto.Message) *RequestContext {
	return &RequestContext{
		ctx,
		requestID,
		hash,
		&request,
		make(chan *proto.Message),
		make(chan error),
		make(chan struct{}),
	}
}

type streamReceiveFunc func() (proto.Message, error)
type streamRequestFunc func(*CoreClient, *context.Context, proto.Message) (streamReceiveFunc, error)
type msgHashFunc func(proto.Message, uuid.UUID) uint64

func generateStreamRequest(request proto.Message, streamReq streamRequestFunc, hashFunc msgHashFunc) *RequestInfo {
	// generate a timestamp/mac addr uuid for the request
	requestID := uuid.NewV1()
	// hash the request reflectively - see HashInclude structs for per-message field inlcude/excludes
	hashVal := hashFunc(request, requestID)

	// create the function that the caller will use to execute the request
	requestFunc := func(ctx *context.Context, client *CoreClient) *RequestContext {
		requestCtx := NewRequestContext(ctx, requestID, hashVal, request)

		// start a go routine that will send the grpc request, and return streamed results and errors through
		// the context object's channels
		go func() {
			defer func() {
				close(requestCtx.Results)
				close(requestCtx.Errors)
			}()
			// execute the grpc call
			receive, err := streamReq(client, ctx, request)
			if err != nil {
				log.Error(err)
				requestCtx.Errors <- errors.Wrap(err, "failed to initiate create pipeline request")
				return
			}
			// handle the result stream
			for {
				result, err := receive()
				if err == io.EOF {
					// EOF signifies the server has finished sending and has terminated the stream successfully
					requestCtx.Done <- struct{}{}
					break
				} else if err != nil {
					// Other errors returned are problems that interrupted the stream and couldn't be recovered from
					log.Error(err)
					requestCtx.Errors <- errors.Wrap(err, "failed to process create pipeline result")
					requestCtx.Done <- struct{}{}
					break
				}
				requestCtx.Results <- &result
			}
			return
		}()
		return requestCtx
	}
	return &RequestInfo{requestID, hashVal, requestFunc}
}

// GeneratePipelineCreateRequest creates a PipelineCreateRequest that will initiate pipeline creation on the server and
// and handle a stream of PipelineCreateResult objects that are returned as work is completed.
func GeneratePipelineCreateRequest(request *PipelineCreateRequest) *RequestInfo {
	grpcFunc := func(client *CoreClient, ctx *context.Context, request proto.Message) (streamReceiveFunc, error) {
		// execute the grpc create pipeline request
		req := request.(*PipelineCreateRequest)
		stream, err := (*client).CreatePipelines(*ctx, req)
		// return a function to receive stream updates
		return func() (proto.Message, error) { return stream.Recv() }, err
	}
	return generateStreamRequest(request, grpcFunc, msgHash)
}

// GeneratePipelineExecuteRequest creates a PipelineExecuteRequest that will execute a pipeline on the server and
// and handle a stream of PipelineExecuteResult objects that are returned as work is completed.
func GeneratePipelineExecuteRequest(request *PipelineExecuteRequest) *RequestInfo {
	grpcFunc := func(client *CoreClient, ctx *context.Context, request proto.Message) (streamReceiveFunc, error) {
		// execute the grpc execute pipeline request
		req := request.(*PipelineExecuteRequest)
		stream, err := (*client).ExecutePipeline(*ctx, req)
		// return a function to receive stream updates
		return func() (proto.Message, error) { return stream.Recv() }, err
	}
	return generateStreamRequest(request, grpcFunc, msgHash)
}

type grpcRequestFunc func(*CoreClient, *context.Context, proto.Message) (proto.Message, error)

func generateRequest(request proto.Message, grpcRequest grpcRequestFunc, hashFunc msgHashFunc) *RequestInfo {
	// generate a timestamp/mac addr uuid for the request
	requestID := uuid.NewV1()

	// hash the request reflectively - see HashInclude structs for per-message field inlcude/excludes
	hashVal := hashFunc(request, requestID)

	// create the function that the caller will use to execute the request
	requestFunc := func(ctx *context.Context, client *CoreClient) *RequestContext {
		requestCtx := NewRequestContext(ctx, requestID, hashVal, request)
		go func() {
			defer func() {
				close(requestCtx.Results)
				close(requestCtx.Errors)
			}()
			// send the grpc reuqest
			result, err := grpcRequest(client, ctx, request)
			if err != nil {
				requestCtx.Errors <- errors.Wrap(err, "failed to initiate end session request")
				requestCtx.Done <- struct{}{}
				return
			}
			requestCtx.Results <- &result
			requestCtx.Done <- struct{}{}
			return
		}()
		return requestCtx
	}
	return &RequestInfo{requestID, hashVal, requestFunc}
}

// GenerateStartSessionRequest creates a session start request that will return a unique session ID
// to the caller. This ID is then assigned to subsquent pipeline calls via the session context field.
func GenerateStartSessionRequest() *RequestInfo {
	sessionRequest := SessionRequest{
		UserAgent: userAgent,
		Version:   APIVersion(),
	}
	grpcFunc := func(client *CoreClient, ctx *context.Context, request proto.Message) (proto.Message, error) {
		// execute the start session request
		return (*client).StartSession(*ctx, &sessionRequest)
	}
	hashFunc := func(msg proto.Message, id uuid.UUID) uint64 { return hash(id.Bytes()) }
	return generateRequest(&sessionRequest, grpcFunc, hashFunc)
}

// GenerateEndSessionRequest creates a session end request that will mark a session as closed. The session
// is not available for further pipeline requests once called.
func GenerateEndSessionRequest(sessionID string) *RequestInfo {
	sessionCtx := SessionContext{sessionID}
	grpcFunc := func(client *CoreClient, ctx *context.Context, request proto.Message) (proto.Message, error) {
		// execute the end session request
		return (*client).EndSession(*ctx, &sessionCtx)
	}
	return generateRequest(&sessionCtx, grpcFunc, msgHash)
}

// GenerateExportPipelineRequest creates a request that signals the pipeline compute server to export the
// pipeline indicated by caller supplied pipeline ID.
func GenerateExportPipelineRequest(sessionID string, pipelineID string, pipelineURI string) *RequestInfo {
	exportRequest := PipelineExportRequest{
		Context:         &SessionContext{sessionID},
		PipelineId:      pipelineID,
		PipelineExecUri: pipelineURI,
	}
	grpcFunc := func(client *CoreClient, ctx *context.Context, request proto.Message) (proto.Message, error) {
		// execute the export pipeline request
		return (*client).ExportPipeline(*ctx, &exportRequest)
	}
	hashFunc := func(msg proto.Message, id uuid.UUID) uint64 { return hash(id.Bytes()) }
	return generateRequest(&exportRequest, grpcFunc, hashFunc)
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

func msgHash(m proto.Message, id uuid.UUID) uint64 {
	hashVal, err := hashstructure.Hash(m, nil)
	if err != nil {
		log.Error("hash fail on message contents - using hash of id")
		hashVal = hash(id.Bytes())
	}
	return hashVal
}

func hash(b []byte) uint64 {
	hash := fnv.New64a()
	hash.Write(b)
	return hash.Sum64()
}

// Gets API version from protobuf file.  Computes once and caches the value since it is immutable.
// Note that protobuf init must be complete before the version can be extracted so we initialize
// lazily.
func getAPIVersion() string {
	// Get the raw file descriptor bytes
	fileDesc := proto.FileDescriptor(E_ProtocolVersion.Filename)
	if fileDesc == nil {
		log.Errorf("failed to find file descriptor for %v", E_ProtocolVersion.Filename)
		return versionUnset
	}

	// Open a gzip reader and decompress
	r, err := gzip.NewReader(bytes.NewReader(fileDesc))
	if err != nil {
		log.Errorf("failed to open gzip reader: %v", err)
		return versionUnset
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Errorf("failed to decompress descriptor: %v", err)
		return versionUnset
	}

	// Unmarshall the bytes from the proto format
	fd := &protobuf.FileDescriptorProto{}
	if err := proto.Unmarshal(b, fd); err != nil {
		log.Errorf("malformed FileDescriptorProto: %v", err)
		return versionUnset
	}

	// Fetch the extension from the FileDescriptorOptions message
	ex, err := proto.GetExtension(fd.GetOptions(), E_ProtocolVersion)
	if err != nil {
		log.Errorf("failed to fetch extension: %v", err)
		return versionUnset
	}
	return *ex.(*string)
}
