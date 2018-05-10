package pipeline

import (
	"context"
	"crypto/sha1"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/unchartedsoftware/distil/api/model"
)

const (
	defaultResourceID       = "0"
	defaultExposedOutputKey = "outputs.0"
	datasetDir              = "datasets"
	// PendingStatus represents that the solution request has been acknoledged by not yet sent to the API
	PendingStatus = "PENDING"
	// RunningStatus represents that the solution request has been sent to the API.
	RunningStatus = "RUNNING"
	// ErroredStatus represents that the solution request has terminated with an error.
	ErroredStatus = "ERRORED"
	// CompletedStatus represents that the solution request has completed successfully.
	CompletedStatus = "COMPLETED"
)

// CreateMessage represents a create model message.
type CreateMessage struct {
	Dataset       string              `json:"dataset"`
	Index         string              `json:"index"`
	TargetFeature string              `json:"target"`
	Task          string              `json:"task"`
	MaxSolutions  int32               `json:"maxSolutions"`
	Filters       *model.FilterParams `json:"filters"`
	Metrics       []string            `json:"metrics"`
}

// CreateStatus represents a solution status.
type CreateStatus struct {
	Progress   string    `json:"progress"`
	RequestID  string    `json:"requestId"`
	SolutionID string    `json:"solutionId"`
	ResultID   string    `json:"resultId"`
	Error      error     `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
}

func (m *CreateMessage) createSearchSolutionsRequest(targetIndex int) (*SearchSolutionsRequest, error) {
	return &SearchSolutionsRequest{
		Problem: &ProblemDescription{
			Problem: &Problem{
				TaskType:           convertTaskTypeFromTA3ToTA2(m.Task),
				PerformanceMetrics: convertMetricsFromTA3ToTA2(m.Metrics),
			},
			Inputs: []*ProblemInput{
				{
					DatasetId: convertDatasetTA3ToTA2(m.Dataset),
					Targets:   convertTargetFeaturesTA3ToTA2(m.TargetFeature, targetIndex),
				},
			},
		},
	}, nil
}

func (m *CreateMessage) createProduceSolutionRequest(datasetURI string, solutionID string) *ProduceSolutionRequest {
	return &ProduceSolutionRequest{
		SolutionId: solutionID,
		Inputs: []*Value{
			{
				Value: &Value_DatasetUri{
					DatasetUri: datasetURI,
				},
			},
		},
	}
}

func (m *CreateMessage) persistSolutionError(statusChan chan CreateStatus, client *Client, solutionStorage model.SolutionStorage, searchID string, solutionID string, err error) {
	// persist the updated state
	// NOTE: ignoring error
	solutionStorage.PersistSolution(searchID, solutionID, ErroredStatus, time.Now())
	// notify of error
	statusChan <- CreateStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		Progress:   ErroredStatus,
		Error:      err,
		Timestamp:  time.Now(),
	}
}

func (m *CreateMessage) persistSolutionStatus(statusChan chan CreateStatus, client *Client, solutionStorage model.SolutionStorage, searchID string, solutionID string, status string) {
	// persist the updated state
	err := solutionStorage.PersistSolution(searchID, solutionID, status, time.Now())
	if err != nil {
		// notify of error
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// notify client of update
	statusChan <- CreateStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		Progress:   status,
		Timestamp:  time.Now(),
	}
}

func (m *CreateMessage) persistSolutionResults(statusChan chan CreateStatus, client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, dataset string, solutionID string, resultID string, resultURI string) {
	// persist the completed state
	err := solutionStorage.PersistSolution(searchID, solutionID, CompletedStatus, time.Now())
	if err != nil {
		// notify of error
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// persist result metadata
	err = solutionStorage.PersistSolutionResult(solutionID, resultID, resultURI, CompletedStatus, time.Now())
	if err != nil {
		// notify of error
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// persist results
	err = dataStorage.PersistResult(dataset, resultURI)
	if err != nil {
		// notify of error
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// notify client of update
	statusChan <- CreateStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		ResultID:   resultID,
		Progress:   CompletedStatus,
		Timestamp:  time.Now(),
	}
}

func (m *CreateMessage) dispatchSolution(statusChan chan CreateStatus, client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, solutionID string, dataset string, datasetURI string) {

	// score solution
	solutionScoreResponses, err := client.GenerateSolutionScores(context.Background(), solutionID)
	if err != nil {
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}

	// persist the scores
	for _, response := range solutionScoreResponses {
		for _, score := range response.Scores {
			err := solutionStorage.PersistSolutionScore(solutionID, score.Metric.Metric.String(), score.Value.GetDouble())
			if err != nil {
				m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
				return
			}
		}
	}

	// fit solution
	_, err = client.GenerateSolutionFit(context.Background(), solutionID)
	if err != nil {
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}

	// persist solution running status
	m.persistSolutionStatus(statusChan, client, solutionStorage, searchID, solutionID, RunningStatus)

	// generate predictions
	produceSolutionRequest := m.createProduceSolutionRequest(datasetURI, solutionID)

	// generate predictions
	predictionResponses, err := client.GeneratePredictions(context.Background(), produceSolutionRequest)
	if err != nil {
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}

	for _, response := range predictionResponses {

		if response.Progress.State != ProgressState_COMPLETED {
			// only persist completed responses
			continue
		}

		output, ok := response.ExposedOutputs[defaultExposedOutputKey]
		if !ok {
			m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, errors.Errorf("output is missing from response"))
			return
		}

		datasetURI, ok := output.Value.(*Value_DatasetUri)
		if !ok {
			m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, errors.Errorf("output is not of correct format"))
			return
		}

		// remove the protocol portion if it exists. The returned value is either a
		// csv file or a directory.
		resultURI := datasetURI.DatasetUri
		resultURI = strings.Replace(resultURI, "file://", "", 1)
		if !strings.HasSuffix(resultURI, ".csv") {
			resultURI = path.Join(resultURI, D3MLearningData)
		}

		// get the result UUID. NOTE: Doing sha1 for now.
		hasher := sha1.New()
		hasher.Write([]byte(resultURI))
		bs := hasher.Sum(nil)
		resultID := fmt.Sprintf("%x", bs)

		// persist results
		m.persistSolutionResults(statusChan, client, solutionStorage, dataStorage, searchID, dataset, solutionID, resultID, resultURI)
	}
}

func (m *CreateMessage) createStatusChannels(client *Client, solutions []*GetSearchSolutionsResultsResponse, solutionStorage model.SolutionStorage, searchID string) []chan CreateStatus {

	// create channels

	// NOTE: WE BUFFER THE CHANNELS TO A SIZE OF 1 HERE SO WE CAN PERSIST BELOW
	// WITHOUT DEADLOCKING.
	var statusChannels []chan CreateStatus
	for range solutions {
		statusChannels = append(statusChannels, make(chan CreateStatus, 1))
	}

	// persist all solutions as pending

	// NOTE: we persist the solutions here so that they exist in the DB when the
	// method returns.
	// NOTE: THE CHANNELS MUST BE BUFFERED TO A SIZE OF 1 OR ELSE THIS WILL DEADLOCK.
	for i, solution := range solutions {
		m.persistSolutionStatus(statusChannels[i], client, solutionStorage, searchID, solution.SolutionId, PendingStatus)
	}

	return statusChannels
}

// DispatchSolutions dispatches all solution requests
func (m *CreateMessage) DispatchSolutions(client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, dataset string, datasetURI string) ([]chan CreateStatus, error) {

	solutions, err := client.SearchSolutions(context.Background(), searchID)
	if err != nil {
		return nil, err
	}

	// create status channels and persist solutions
	statusChannels := m.createStatusChannels(client, solutions, solutionStorage, searchID)

	// dispatch all solutions
	go func() {

		wg := &sync.WaitGroup{}

		// dispatch individual solutions
		for i, solution := range solutions {

			// increment waitgroup
			wg.Add(1)

			go func(statusChan chan CreateStatus, solutionID string) {
				m.dispatchSolution(statusChan, client, solutionStorage, dataStorage, searchID, solutionID, dataset, datasetURI)
				wg.Done()
			}(statusChannels[i], solution.SolutionId)

		}

		// wait until all are complete
		wg.Wait()

		// end search
		client.EndSearch(context.Background(), searchID)
	}()

	return statusChannels, nil
}

// PersistAndDispatch persists the solution request and dispatches it.
func (m *CreateMessage) PersistAndDispatch(client *Client, solutionStorage model.SolutionStorage, metaStorage model.MetadataStorage, dataStorage model.DataStorage) ([]chan CreateStatus, error) {

	// NOTE: D3M index field is needed in the persisted data.
	m.Filters.Variables = append(m.Filters.Variables, model.D3MIndexFieldName)

	// fetch the queried dataset
	dataset, err := model.FetchDataset(m.Dataset, m.Index, true, m.Filters, metaStorage, dataStorage)
	if err != nil {
		return nil, err
	}

	// perist the dataset and get URI
	datasetPath, targetIndex, err := PersistFilteredData(datasetDir, m.TargetFeature, dataset)
	if err != nil {
		return nil, err
	}
	// make sure the path is absolute and contains the URI prefix
	datasetPath, err = filepath.Abs(datasetPath)
	if err != nil {
		return nil, err
	}
	datasetPath = fmt.Sprintf("%s", filepath.Join(datasetPath, D3MDataSchema))

	// create search solutions request
	searchRequest, err := m.createSearchSolutionsRequest(targetIndex)
	if err != nil {
		return nil, err
	}

	// start a solution searchID
	requestID, err := client.StartSearch(context.Background(), searchRequest)
	if err != nil {
		return nil, err
	}

	// persist the request
	err = solutionStorage.PersistRequest(requestID, dataset.Metadata.Name, PendingStatus, time.Now())
	if err != nil {
		return nil, err
	}

	// store the request features
	for _, v := range m.Filters.Variables {
		var typ string
		// ignore the index field
		if v == model.D3MIndexFieldName {
			continue
		}

		if v == m.TargetFeature {
			// store target feature
			typ = model.FeatureTypeTarget
		} else {
			// store training feature
			typ = model.FeatureTypeTrain
		}
		err = solutionStorage.PersistRequestFeature(requestID, v, typ)
		if err != nil {
			return nil, err
		}
	}

	// store request filters
	err = solutionStorage.PersistRequestFilters(requestID, m.Filters)
	if err != nil {
		return nil, err
	}

	// dispatch solutions
	statusChannels, err := m.DispatchSolutions(client, solutionStorage, dataStorage, requestID, dataset.Metadata.Name, datasetPath)
	if err != nil {
		return nil, err
	}

	return statusChannels, nil
}

func convertMetricsFromTA3ToTA2(metrics []string) []*ProblemPerformanceMetric {
	var res []*ProblemPerformanceMetric
	for _, metric := range metrics {
		res = append(res, &ProblemPerformanceMetric{
			Metric: PerformanceMetric(PerformanceMetric_value[strings.ToUpper(metric)]),
		})
	}
	return res
}

func convertTaskTypeFromTA3ToTA2(taskType string) TaskType {
	return TaskType(TaskType_value[strings.ToUpper(taskType)])
}

func convertTargetFeaturesTA3ToTA2(target string, targetIndex int) []*ProblemTarget {
	return []*ProblemTarget{
		{
			ColumnName:  target,
			ResourceId:  defaultResourceID,
			TargetIndex: int32(targetIndex),
			ColumnIndex: int32(targetIndex), // TODO: is this correct?
		},
	}
}

func convertDatasetTA3ToTA2(dataset string) string {
	return dataset
}
