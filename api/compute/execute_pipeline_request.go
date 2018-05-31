package compute

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

// ExecPipelineStatus contains status / result information for a pipeline status
// request.
type ExecPipelineStatus struct {
	Progress  string
	RequestID string
	Error     error
	Timestamp time.Time
	ResultURI string
}

// ExecPipelineStatusListener defines a funtction type for handling pipeline
// execution result updates.
type ExecPipelineStatusListener func(status ExecPipelineStatus)

// ExecPipelineRequest defines a request that will execute a fully specified pipline
// on a TA2 system.
type ExecPipelineRequest struct {
	datasetURI    string
	pipelineDesc  *pipeline.PipelineDescription
	wg            *sync.WaitGroup
	statusChannel chan ExecPipelineStatus
	finished      chan error
}

// NewExecPipelineRequest creates a new request that will run the supplied dataset through
// the pipeline description.
func (e *ExecPipelineRequest) NewExecPipelineRequest(datasetURI string, pipelineDesc *pipeline.PipelineDescription) *ExecPipelineRequest {
	return &ExecPipelineRequest{
		datasetURI:    datasetURI,
		pipelineDesc:  pipelineDesc,
		wg:            &sync.WaitGroup{},
		finished:      make(chan error),
		statusChannel: make(chan ExecPipelineStatus, 1),
	}
}

// Listen listens for new solution statuses and invokes the caller supplied function
// when a status update is received.  The call will block until the request completes.
func (e *ExecPipelineRequest) Listen(listener ExecPipelineStatusListener) error {
	go func() {
		for {
			listener(<-e.statusChannel)
		}
	}()
	return <-e.finished
}

// Dispatch dispatches a pipeline exeucute request for processing by TA2
func (e *ExecPipelineRequest) Dispatch(client *Client) error {
	requestID, err := client.StartSearch(context.Background(), &pipeline.SearchSolutionsRequest{
		Version:   GetAPIVersion(),
		UserAgent: client.UserAgent,
		Template:  e.pipelineDesc,
	})
	if err != nil {
		return err
	}

	// dispatch search request
	go e.dispatchRequest(client, requestID)

	return nil
}

func (e *ExecPipelineRequest) dispatchRequest(client *Client, requestID string) {

	// Update request status
	e.notifyStatus(e.statusChannel, requestID, RequestPendingStatus)

	var firstSolution string
	var produceCalled bool
	// Search for solutions, this wont return until the produce finishes or it times out.
	err := client.SearchSolutions(context.Background(), requestID, func(solution *pipeline.GetSearchSolutionsResultsResponse) {
		defer e.wg.Done()

		// A complete pipeline specification should result in a single solution being generated.  Consider it an
		// error condition when that is not the case.
		if firstSolution == "" {
			firstSolution = solution.GetSolutionId()
		} else if firstSolution != solution.GetSolutionId() {
			err := errors.Errorf("multiple solutions found for request %s, expected 1", requestID)
			e.notifyError(e.statusChannel, requestID, err)
			return
		}

		// handle solution search update - status codes pertain to the search itself, and not a particular
		// solution
		switch solution.GetProgress().GetState() {
		case pipeline.ProgressState_ERRORED:
			// search errored - most likely case is that the supplied pipeline had a problem in its specification
			err := errors.Errorf("could not generate solution for request - %s", solution.GetProgress().GetStatus())
			e.notifyError(e.statusChannel, requestID, err)
			return
		case pipeline.ProgressState_RUNNING:
			fallthrough
		case pipeline.ProgressState_COMPLETED:
			// sarch is actively running or has completed - safe to call produce at this point, but we should
			// only do so once.  A status update with no actual solution ID is valid.
			e.notifyStatus(e.statusChannel, requestID, RequestRunningStatus)
			if solution.GetSolutionId() != "" && !produceCalled {
				produceCalled = true
				e.dispatchProduce(e.statusChannel, client, requestID, solution.GetSolutionId())
			}
		default:
			e.notifyStatus(e.statusChannel, requestID, RequestRunningStatus)
		}
	})

	if err != nil {
		e.notifyError(e.statusChannel, requestID, err)
	}

	// wait until all are complete and the search has finished / timed out
	e.wg.Wait()

	// end search
	e.finished <- client.EndSearch(context.Background(), requestID)
}

func (e *ExecPipelineRequest) createProduceSolutionRequest(datsetURI string, solutionID string) *pipeline.ProduceSolutionRequest {
	return &pipeline.ProduceSolutionRequest{
		SolutionId: solutionID,
		Inputs: []*pipeline.Value{
			{
				Value: &pipeline.Value_DatasetUri{
					DatasetUri: e.datasetURI,
				},
			},
		},
		ExposeOutputs: []string{defaultExposedOutputKey},
		ExposeValueTypes: []pipeline.ValueType{
			pipeline.ValueType_CSV_URI,
		},
	}
}

func (e *ExecPipelineRequest) dispatchProduce(statusChan chan ExecPipelineStatus, client *Client, requestID string, solutionID string) {
	// generate predictions
	produceRequest := e.createProduceSolutionRequest(e.datasetURI, solutionID)

	// run produce - this blocks until all responses are returned
	responses, err := client.GeneratePredictions(context.Background(), produceRequest)
	if err != nil {
		e.notifyError(statusChan, requestID, err)
		return
	}

	// find the completed response
	var completed *pipeline.GetProduceSolutionResultsResponse
	for _, response := range responses {
		if response.Progress.State == pipeline.ProgressState_COMPLETED {
			completed = response
			break
		}
	}
	if completed == nil {
		err := errors.Errorf("no completed response found")
		e.notifyError(statusChan, requestID, err)
		return
	}

	// make sure the exposed output is what was asked for
	output, ok := completed.ExposedOutputs[defaultExposedOutputKey]
	if !ok {
		err := errors.Errorf("output is missing from response")
		e.notifyError(statusChan, requestID, err)
		return
	}

	var uri string
	results := output.Value
	switch res := results.(type) {
	case *pipeline.Value_DatasetUri:
		uri = res.DatasetUri
	case *pipeline.Value_CsvUri:
		uri = res.CsvUri
	default:
		err = errors.Errorf("unexpected result type '%v'", res)
		e.notifyError(statusChan, requestID, err)
	}
	e.notifyResult(statusChan, requestID, uri)
}

func (e *ExecPipelineRequest) notifyStatus(statusChan chan ExecPipelineStatus, requestID string, status string) {
	// notify of update
	statusChan <- ExecPipelineStatus{
		RequestID: requestID,
		Progress:  status,
		Timestamp: time.Now(),
	}
}

func (e *ExecPipelineRequest) notifyError(statusChan chan ExecPipelineStatus, requestID string, err error) {
	statusChan <- ExecPipelineStatus{
		RequestID: requestID,
		Progress:  RequestErroredStatus,
		Error:     err,
		Timestamp: time.Now(),
	}
}

func (e *ExecPipelineRequest) notifyResult(statusChan chan ExecPipelineStatus, requestID string, resultURI string) {
	statusChan <- ExecPipelineStatus{
		RequestID: requestID,
		Progress:  RequestCompletedStatus,
		ResultURI: resultURI,
		Timestamp: time.Now(),
	}
}
