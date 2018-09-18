package compute

import (
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/middleware"
	"github.com/unchartedsoftware/distil/api/pipeline"
	"github.com/unchartedsoftware/plog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// ClientMock blah blah.
type ClientMock struct {
	conn *grpc.ClientConn
}

// NewClientMock blah blah.
func NewClientMock(serverAddr string, trace bool) (*ClientMock, error) {

	log.Infof("connecting to ta2 at %s", serverAddr)

	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(middleware.GenerateUnaryClientInterceptor(trace)),
		grpc.WithStreamInterceptor(middleware.GenerateStreamClientInterceptor(trace)),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to %s", serverAddr)
	}

	log.Infof("connected to %s", serverAddr)

	client := ClientMock{}
	client.conn = conn
	return &client, nil
}

// Close the connection to the solution service.
func (c *ClientMock) Close() {
	log.Infof("client connection closed")
	c.conn.Close()
}

// ExecutePipeline executes a pre-specified pipeline.
func (c *ClientMock) ExecutePipeline(ctx context.Context, pipelineDesc *pipeline.PipelineDescription) (*pipeline.PipelineExecuteResponse, error) {

	in := &pipeline.PipelineExecuteRequest{
		PipelineDescription: pipelineDesc,
	}

	out := new(pipeline.PipelineExecuteResponse)
	err := c.conn.Invoke(ctx, "/Executor/ExecutePipeline", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
