package pipeline

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// NewPipelineClient creates a new grpc client instance.
func NewPipelineClient(serverAddr string) (*PipelineComputeClient, error) {
	conn, err := grpc.Dial(serverAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to at %s", serverAddr)
	}
	client := NewPipelineComputeClient(conn)
	return &client, nil
}
