package compute

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

type ExecPipelineStatus struct {
	Progress  string
	RequestID string
	Error     error
	Timestamp time.Time
	ResultURI string
}

type ExecPipelineStatusListener func(status ExecPipelineStatus)

// ExecPipelineRequest defines a request that will execute a fully specified pipline
// on a TA2 system.
type ExecPipelineRequest struct {
	datasetURI       string
	pipelineDesc     *pipeline.PipelineDescription
	wg               *sync.WaitGroup
	mu               *sync.Mutex
	requestChannel   chan ExecPipelineStatus
	solutionChannels []chan ExecPipelineStatus
	listener         ExecPipelineStatusListener
	finished         chan error
}

// NewExecPipelineRequest creates a new request that will run the supplied dataset through
// the pipeline description.
func (e *ExecPipelineRequest) NewExecPipelineRequest(datasetURI string, pipelineDesc *pipeline.PipelineDescription) *ExecPipelineRequest {
	return &ExecPipelineRequest{
		datasetURI:     datasetURI,
		pipelineDesc:   pipelineDesc,
		wg:             &sync.WaitGroup{},
		finished:       make(chan error),
		requestChannel: make(chan ExecPipelineStatus, 1),
	}
}

// Listen listens on the solution requests for new solution statuses.  Status
// updates are buffered, so listening on a request will provide all status up
// to the point the call is made, and provide additional solution status updates
// as the call progresses.
func (e *ExecPipelineRequest) Listen(listener ExecPipelineStatusListener) error {
	e.listener = listener
	e.mu.Lock()
	// listen on main request channel
	go e.listenOnStatusChannel(e.requestChannel)
	// listen on individual solution channels
	for _, c := range e.solutionChannels {
		go e.listenOnStatusChannel(c)
	}
	e.mu.Unlock()
	return <-e.finished
}

func (e *ExecPipelineRequest) listenOnStatusChannel(statusChannel chan ExecPipelineStatus) {
	for {
		// read status from, channel
		status := <-statusChannel
		// execute callback
		e.listener(status)
	}
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

func (e *ExecPipelineRequest) dispatchRequest(client *Client, searchID string) {

	// update request status
	e.updateStatus(e.requestChannel, searchID, RequestPendingStatus)

	// search for solutions, this wont return until the search finishes or it times out
	err := client.SearchSolutions(context.Background(), searchID, func(solution *pipeline.GetSearchSolutionsResultsResponse) {
		// create a new status channel for the solution
		c := make(chan ExecPipelineStatus, 1)
		// add the solution to the request
		e.addSolution(c)
		// persist the solution
		e.updateStatus(c, searchID, SolutionPendingStatus)
		// dispatch it
		e.dispatchSolution(c, client, searchID, solution.SolutionId)
		// once done, mark as complete
		e.completeSolution()
	})

	// update request status
	if err != nil {
		e.updateError(e.requestChannel, searchID, err)
	} else {
		e.updateStatus(e.requestChannel, searchID, RequestCompletedStatus)
	}

	// wait until all are complete and the search has finished / timed out
	e.waitOnSolutions()

	// end search
	e.finished <- client.EndSearch(context.Background(), searchID)
}

func (e *ExecPipelineRequest) addSolution(c chan ExecPipelineStatus) {
	e.wg.Add(1)
	e.mu.Lock()
	e.solutionChannels = append(e.solutionChannels, c)
	if e.listener != nil {
		go e.listenOnStatusChannel(c)
	}
	e.mu.Unlock()
}

func (e *ExecPipelineRequest) completeSolution() {
	e.wg.Done()
}

func (e *ExecPipelineRequest) waitOnSolutions() {
	e.wg.Wait()
}

func (e *ExecPipelineRequest) updateStatus(statusChan chan ExecPipelineStatus, searchID string, status string) {
	// notify of update
	statusChan <- ExecPipelineStatus{
		RequestID: searchID,
		Progress:  status,
		Timestamp: time.Now(),
	}
}

func (e *ExecPipelineRequest) updateError(statusChan chan ExecPipelineStatus, searchID string, err error) {
	statusChan <- ExecPipelineStatus{
		RequestID: searchID,
		Progress:  RequestErroredStatus,
		Error:     err,
		Timestamp: time.Now(),
	}
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
			pipeline.ValueType_DATASET_URI,
		},
	}
}

func (e *ExecPipelineRequest) dispatchSolution(statusChan chan ExecPipelineStatus, client *Client, searchID string, solutionID string) {
	// generate predictions
	produceSolutionRequest := e.createProduceSolutionRequest(e.datasetURI, solutionID)

	// generate predictions
	predictionResponses, err := client.GeneratePredictions(context.Background(), produceSolutionRequest)
	if err != nil {
		e.updateError(statusChan, searchID, err)
		return
	}

	for _, response := range predictionResponses {

		if response.Progress.State != pipeline.ProgressState_COMPLETED {
			// only persist completed responses
			continue
		}

		output, ok := response.ExposedOutputs[defaultExposedOutputKey]
		if !ok {
			err := errors.Errorf("output is missing from response")
			e.updateError(statusChan, searchID, err)
			return
		}

		var uri string
		var err error
		results := output.Value
		switch res := results.(type) {
		case *pipeline.Value_DatasetUri:
			uri = res.DatasetUri
		case *pipeline.Value_CsvUri:
			uri = res.CsvUri
		default:
			err = errors.Errorf("unexpected result type '%v'", res)
			e.updateError(statusChan, searchID, err)
		}

		statusChan <- ExecPipelineStatus{
			RequestID: searchID,
			Progress:  RequestCompletedStatus,
			Timestamp: time.Now(),
			ResultURI: uri,
		}
		return
	}
}
