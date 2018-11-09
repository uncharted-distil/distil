package compute

import (
	"fmt"
	"io"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/middleware"
	"github.com/unchartedsoftware/distil-compute/pipeline"
	"github.com/unchartedsoftware/plog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const defaultTrainTestRatio = 3

// Client provides facilities for managing GPRC solution requests. Requests are
// isssued and a context object containing rx channels is returned to the caller for consumption
// of results. The context for running requests can also be fetched, along with their buffered
// results. Spawning a grpc.ClientConn per RPC call is not considered good practice - the system
// is designed such that multiple go routines make RPC calls to a single shared client, and synch
// is managed internally.
type Client struct {
	client            pipeline.CoreClient
	conn              *grpc.ClientConn
	runner            *grpc.ClientConn
	mu                *sync.Mutex
	UserAgent         string
	PullTimeout       time.Duration
	PullMax           int
	SkipPreprocessing bool
}

// SearchSolutionHandler is executed when a new search solution is returned.
type SearchSolutionHandler func(*pipeline.GetSearchSolutionsResultsResponse)

// NewClient creates a new pipline request dispatcher instance. This will establish
// the connection to the solution server or return an error on fail
func NewClient(serverAddr string, trace bool, userAgent string,
	pullTimeout time.Duration, pullMax int, skipPreprocessing bool) (*Client, error) {
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

	client := Client{
		client:            pipeline.NewCoreClient(conn),
		conn:              conn,
		UserAgent:         userAgent,
		PullTimeout:       pullTimeout,
		PullMax:           pullMax,
		SkipPreprocessing: skipPreprocessing,
	}

	// check for basic ta2 connectivity
	helloResponse, err := client.client.Hello(context.Background(), &pipeline.HelloRequest{})
	if err != nil {
		return nil, err
	}
	log.Infof("ta2 user agent: %s", helloResponse.GetUserAgent())
	log.Infof("ta2 API version: %s", helloResponse.GetVersion())
	log.Infof("ta2 Allowed value types: %+v", helloResponse.GetAllowedValueTypes())
	log.Infof("ta2 extensions: %+v", helloResponse.GetSupportedExtensions())

	if !strings.EqualFold(GetAPIVersion(), helloResponse.GetVersion()) {
		log.Warnf("ta2 API version '%s' does not match expected version '%s", helloResponse.GetVersion(), GetAPIVersion())
	}

	return &client, nil
}

// NewClientWithRunner creates a new pipline request dispatcher instance. This will establish
// the connection to the solution server or return an error on fail
func NewClientWithRunner(serverAddr string, runnerAddr string, trace bool, userAgent string, pullTimeout time.Duration, pullMax int, skipPreprocessing bool) (*Client, error) {

	client, err := NewClient(serverAddr, trace, userAgent, pullTimeout, pullMax, skipPreprocessing)
	if err != nil {
		return nil, err
	}

	log.Infof("connecting to ta2 runner at %s", runnerAddr)

	runner, err := grpc.Dial(
		runnerAddr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(middleware.GenerateUnaryClientInterceptor(trace)),
		grpc.WithStreamInterceptor(middleware.GenerateStreamClientInterceptor(trace)),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to %s", runnerAddr)
	}

	log.Infof("connected to %s", runnerAddr)

	client.runner = runner
	return client, nil
}

// Close the connection to the solution service
func (c *Client) Close() {
	log.Infof("client connection closed")
	c.conn.Close()
}

// StartSearch starts a solution search session.
func (c *Client) StartSearch(ctx context.Context, request *pipeline.SearchSolutionsRequest) (string, error) {

	searchSolutionResponse, err := c.client.SearchSolutions(ctx, request)
	if err != nil {
		return "", errors.Wrap(err, "failed to start search")
	}

	return searchSolutionResponse.SearchId, nil
}

// SearchSolutions generates candidate pipelines and executes a provided handler
// for each result. While handlers are executing asynchronously, this method
// will not return until all handlers have finished.
func (c *Client) SearchSolutions(ctx context.Context, searchID string, solutionHandler SearchSolutionHandler) error {

	searchPiplinesResultsRequest := &pipeline.GetSearchSolutionsResultsRequest{
		SearchId: searchID,
	}

	searchSolutionsResultsResponse, err := c.client.GetSearchSolutionsResults(ctx, searchPiplinesResultsRequest)
	if err != nil {
		return errors.Wrap(err, "failed to open search results stream")
	}

	// track handlers to ensure they all finish before returning
	wg := &sync.WaitGroup{}

	err = pullFromAPI(c.PullMax, c.PullTimeout, func() error {
		solutionResultResponse, err := searchSolutionsResultsResponse.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return errors.Wrap(err, "failed to get search result")
		}
		// ignore empty responses
		if solutionResultResponse.SolutionId != "" {
			wg.Add(1)
			go func() {
				solutionHandler(solutionResultResponse)
				wg.Done()
			}()
		}
		return nil
	})
	if err != nil {
		return err
	}

	// don't return until all handlers have finished executing
	wg.Wait()
	return nil
}

// GenerateSolutionScores generates scrores for candidate solutions.
func (c *Client) GenerateSolutionScores(ctx context.Context, solutionID string, datasetURI string, metrics []string) ([]*pipeline.GetScoreSolutionResultsResponse, error) {

	scoreSolutionRequest := &pipeline.ScoreSolutionRequest{
		SolutionId: solutionID,
		Inputs: []*pipeline.Value{
			{
				Value: &pipeline.Value_DatasetUri{
					DatasetUri: datasetURI,
				},
			},
		},
		PerformanceMetrics: convertMetricsFromTA3ToTA2(metrics),
		Configuration: &pipeline.ScoringConfiguration{
			Method:         pipeline.EvaluationMethod_HOLDOUT,
			TrainTestRatio: defaultTrainTestRatio,
		},
	}

	scoreSolutionResponse, err := c.client.ScoreSolution(ctx, scoreSolutionRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start solution scoring")
	}

	searchPiplinesResultsRequest := &pipeline.GetScoreSolutionResultsRequest{
		RequestId: scoreSolutionResponse.RequestId,
	}

	scoreSolutionResultsResponse, err := c.client.GetScoreSolutionResults(ctx, searchPiplinesResultsRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open solution scoring results stream")
	}

	var solutionResultResponses []*pipeline.GetScoreSolutionResultsResponse

	err = pullFromAPI(c.PullMax, c.PullTimeout, func() error {
		solutionResultResponse, err := scoreSolutionResultsResponse.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return errors.Wrap(err, "failed to receive solution scoring result")
		}
		solutionResultResponses = append(solutionResultResponses, solutionResultResponse)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return solutionResultResponses, nil
}

// GenerateSolutionFit generates fit for candidate solutions.
func (c *Client) GenerateSolutionFit(ctx context.Context, solutionID string, datasetURI string) ([]*pipeline.GetFitSolutionResultsResponse, error) {

	fitSolutionRequest := &pipeline.FitSolutionRequest{
		SolutionId: solutionID,
		Inputs: []*pipeline.Value{
			{
				Value: &pipeline.Value_DatasetUri{
					DatasetUri: datasetURI,
				},
			},
		},
	}

	fitSolutionResponse, err := c.client.FitSolution(ctx, fitSolutionRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start solution fitting")
	}

	fitSolutionResultsRequest := &pipeline.GetFitSolutionResultsRequest{
		RequestId: fitSolutionResponse.RequestId,
	}

	fitSolutionResultsResponse, err := c.client.GetFitSolutionResults(ctx, fitSolutionResultsRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open solution fitting result stream")
	}

	var solutionResultResponses []*pipeline.GetFitSolutionResultsResponse

	err = pullFromAPI(c.PullMax, c.PullTimeout, func() error {
		solutionResultResponse, err := fitSolutionResultsResponse.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return errors.Wrap(err, "failed to receving solution fitting result")
		}
		solutionResultResponses = append(solutionResultResponses, solutionResultResponse)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return solutionResultResponses, nil
}

// GeneratePredictions generates predictions.
func (c *Client) GeneratePredictions(ctx context.Context, request *pipeline.ProduceSolutionRequest) ([]*pipeline.GetProduceSolutionResultsResponse, error) {

	produceSolutionResponse, err := c.client.ProduceSolution(ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start solution produce")
	}

	produceSolutionResultsRequest := &pipeline.GetProduceSolutionResultsRequest{
		RequestId: produceSolutionResponse.RequestId,
	}

	produceSolutionResultsResponse, err := c.client.GetProduceSolutionResults(ctx, produceSolutionResultsRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open solution produce result stream")
	}

	var solutionResultResponses []*pipeline.GetProduceSolutionResultsResponse

	err = pullFromAPI(c.PullMax, c.PullTimeout, func() error {
		solutionResultResponse, err := produceSolutionResultsResponse.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return errors.Wrap(err, "failed to receive solution produce result")
		}
		solutionResultResponses = append(solutionResultResponses, solutionResultResponse)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return solutionResultResponses, nil
}

// StopSearch stop the solution search session.
func (c *Client) StopSearch(ctx context.Context, searchID string) error {

	stopSearchSolutions := &pipeline.StopSearchSolutionsRequest{
		SearchId: searchID,
	}

	_, err := c.client.StopSearchSolutions(ctx, stopSearchSolutions)
	return errors.Wrap(err, "failed to stop solution search")
}

// EndSearch ends the solution search session.
func (c *Client) EndSearch(ctx context.Context, searchID string) error {

	endSearchSolutions := &pipeline.EndSearchSolutionsRequest{
		SearchId: searchID,
	}

	_, err := c.client.EndSearchSolutions(ctx, endSearchSolutions)
	return errors.Wrap(err, "failed to end solution search")
}

// ExportSolution exports the solution.
func (c *Client) ExportSolution(ctx context.Context, fittedSolutionID string) error {
	exportSolution := &pipeline.SolutionExportRequest{
		Rank:             1,
		FittedSolutionId: fittedSolutionID,
	}
	_, err := c.client.SolutionExport(ctx, exportSolution)
	return errors.Wrap(err, "failed to export solution")
}

// ExecutePipeline executes a pre-specified pipeline.
func (c *Client) ExecutePipeline(ctx context.Context, datasetURI string, pipelineDesc *pipeline.PipelineDescription) (*pipeline.PipelineExecuteResponse, error) {

	datasetURI = fmt.Sprintf("file://%s", path.Join(datasetURI, D3MDataSchema))

	in := &pipeline.PipelineExecuteRequest{
		PipelineDescription: pipelineDesc,
		Inputs: []*pipeline.Value{
			{
				Value: &pipeline.Value_DatasetUri{
					DatasetUri: datasetURI,
				},
			},
		},
	}
	out := new(pipeline.PipelineExecuteResponse)
	err := c.runner.Invoke(ctx, "/Executor/ExecutePipeline", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
