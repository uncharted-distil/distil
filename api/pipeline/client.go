package pipeline

import (
	"sync"

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
	client  CoreClient
	conn    *grpc.ClientConn
	mu      *sync.Mutex
	DataDir string
}

// NewClient creates a new pipline request dispatcher instance. This will establish
// the connection to the pipeline server or return an error on fail
func NewClient(serverAddr string, dataDir string, trace bool) (*Client, error) {
	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(middleware.GenerateUnaryClientInterceptor(trace)),
		grpc.WithStreamInterceptor(middleware.GenerateStreamClientInterceptor(trace)),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to %s", serverAddr)
	}
	log.Infof("connected to %s", serverAddr)

	client := Client{}
	client.client = NewCoreClient(conn)
	client.conn = conn
	client.DataDir = dataDir
	return &client, nil
}

// Close the connection to the pipeline service
func (c *Client) Close() {
	log.Infof("client connection closed")
	c.conn.Close()
}

// StartSearch starts a pipeline search session.
func (c *Client) StartSearch(ctx context.Context, request *SearchPipelinesRequest) (string, error) {

	searchPipelineResponse, err := c.client.SearchPipelines(ctx, request)
	if err != nil {
		return "", err
	}

	return searchPipelineResponse.SearchId, nil
}

// SearchPipelines generates candidate pipel\ines.
func (c *Client) SearchPipelines(ctx context.Context, searchID string) ([]*GetSearchPipelinesResultsResponse, error) {

	searchPiplinesResultsRequest := &GetSearchPipelinesResultsRequest{
		SearchId: searchID,
	}

	searchPipelinesResultsResponse, err := c.client.GetSearchPipelinesResults(ctx, searchPiplinesResultsRequest)
	if err != nil {
		return nil, err
	}

	var pipelineResultResponses []*GetSearchPipelinesResultsResponse

	err = pullFromAPI(pullMax, pullTimeout, func() error {
		pipelineResultResponse, err := searchPipelinesResultsResponse.Recv()
		if err != nil {
			return err
		}
		pipelineResultResponses = append(pipelineResultResponses, pipelineResultResponse)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pipelineResultResponses, nil
}

// GeneratePipelineScores generates scrores for candidate pipelines.
func (c *Client) GeneratePipelineScores(ctx context.Context, pipelineID string) ([]*GetScorePipelineResultsResponse, error) {

	scorePipelineRequest := &ScorePipelineRequest{
		PipelineId: pipelineID,
	}

	scorePipelineResponse, err := c.client.ScorePipeline(ctx, scorePipelineRequest)
	if err != nil {
		return nil, err
	}

	searchPiplinesResultsRequest := &GetScorePipelineResultsRequest{
		RequestId: scorePipelineResponse.RequestId,
	}

	scorePipelineResultsResponse, err := c.client.GetScorePipelineResults(ctx, searchPiplinesResultsRequest)
	if err != nil {
		return nil, err
	}

	var pipelineResultResponses []*GetScorePipelineResultsResponse

	err = pullFromAPI(pullMax, pullTimeout, func() error {
		pipelineResultResponse, err := scorePipelineResultsResponse.Recv()
		if err != nil {
			return err
		}
		pipelineResultResponses = append(pipelineResultResponses, pipelineResultResponse)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pipelineResultResponses, nil
}

// GeneratePipelineFit generates fit for candidate pipelines.
func (c *Client) GeneratePipelineFit(ctx context.Context, pipelineID string) ([]*GetFitPipelineResultsResponse, error) {

	fitPipelineRequest := &FitPipelineRequest{
		PipelineId: pipelineID,
	}

	fitPipelineResponse, err := c.client.FitPipeline(ctx, fitPipelineRequest)
	if err != nil {
		return nil, err
	}

	fitPipelineResultsRequest := &GetFitPipelineResultsRequest{
		RequestId: fitPipelineResponse.RequestId,
	}

	fitPipelineResultsResponse, err := c.client.GetFitPipelineResults(ctx, fitPipelineResultsRequest)
	if err != nil {
		return nil, err
	}

	var pipelineResultResponses []*GetFitPipelineResultsResponse

	err = pullFromAPI(pullMax, pullTimeout, func() error {
		pipelineResultResponse, err := fitPipelineResultsResponse.Recv()
		if err != nil {
			return err
		}
		pipelineResultResponses = append(pipelineResultResponses, pipelineResultResponse)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pipelineResultResponses, nil
}

// GeneratePredictions generates predictions.
func (c *Client) GeneratePredictions(ctx context.Context, request *ProducePipelineRequest) ([]*GetProducePipelineResultsResponse, error) {

	producePipelineResponse, err := c.client.ProducePipeline(ctx, request)
	if err != nil {
		return nil, err
	}

	producePipelineResultsRequest := &GetProducePipelineResultsRequest{
		RequestId: producePipelineResponse.RequestId,
	}

	producePipelineResultsResponse, err := c.client.GetProducePipelineResults(ctx, producePipelineResultsRequest)
	if err != nil {
		return nil, err
	}

	var pipelineResultResponses []*GetProducePipelineResultsResponse

	err = pullFromAPI(pullMax, pullTimeout, func() error {
		pipelineResultResponse, err := producePipelineResultsResponse.Recv()
		if err != nil {
			return err
		}
		pipelineResultResponses = append(pipelineResultResponses, pipelineResultResponse)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pipelineResultResponses, nil
}

// EndSearch ends the pipeline search session.
func (c *Client) EndSearch(ctx context.Context, searchID string) error {

	endSearchPipelines := &EndSearchPipelinesRequest{
		SearchId: searchID,
	}

	_, err := c.client.EndSearchPipelines(ctx, endSearchPipelines)
	return err
}

// ExportPipeline exports the pipeline.
func (c *Client) ExportPipeline(ctx context.Context, pipelineID string, exportURI string) error {
	return nil
}
