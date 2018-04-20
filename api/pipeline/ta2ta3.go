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
	// PendingStatus represents that the pipeline request has been acknoledged by not yet sent to the API
	PendingStatus = "PENDING"
	// RunningStatus represents that the pipeline request has been sent to the API.
	RunningStatus = "RUNNING"
	// ErroredStatus represents that the pipeline request has terminated with an error.
	ErroredStatus = "ERRORED"
	// CompletedStatus represents that the pipeline request has completed successfully.
	CompletedStatus = "COMPLETED"
)

// CreateMessage represents a create model message.
type CreateMessage struct {
	Dataset       string              `json:"dataset"`
	Index         string              `json:"index"`
	TargetFeature string              `json:"target"`
	Task          string              `json:"task"`
	MaxPipelines  int32               `json:"maxPipelines"`
	Filters       *model.FilterParams `json:"filters"`
	Metrics       []string            `json:"metrics"`
}

// CreateStatus represents a pipeline status.
type CreateStatus struct {
	Progress   string    `json:"progress"`
	RequestID  string    `json:"requestId"`
	PipelineID string    `json:"pipelineId"`
	ResultID   string    `json:"resultId"`
	Error      error     `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
}

func (m *CreateMessage) createSearchPipelinesRequest(targetIndex int) (*SearchPipelinesRequest, error) {
	return &SearchPipelinesRequest{
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

func (m *CreateMessage) createProducePipelineRequest(datasetURI string, pipelineID string) *ProducePipelineRequest {
	return &ProducePipelineRequest{
		PipelineId: pipelineID,
		Inputs: []*Value{
			{
				Value: &Value_DatasetUri{
					DatasetUri: datasetURI,
				},
			},
		},
	}
}

func (m *CreateMessage) persistPipelineError(statusChan chan CreateStatus, client *Client, pipelineStorage model.PipelineStorage, searchID string, pipelineID string, err error) {
	// persist the updated state
	// NOTE: ignoring error
	pipelineStorage.PersistPipeline(searchID, pipelineID, ErroredStatus, time.Now())
	// notify of error
	statusChan <- CreateStatus{
		RequestID:  searchID,
		PipelineID: pipelineID,
		Progress:   ErroredStatus,
		Error:      err,
		Timestamp:  time.Now(),
	}
}

func (m *CreateMessage) persistPipelineStatus(statusChan chan CreateStatus, client *Client, pipelineStorage model.PipelineStorage, searchID string, pipelineID string, status string) {
	// persist the updated state
	err := pipelineStorage.PersistPipeline(searchID, pipelineID, status, time.Now())
	if err != nil {
		// notify of error
		m.persistPipelineError(statusChan, client, pipelineStorage, searchID, pipelineID, err)
		return
	}
	// notify client of update
	statusChan <- CreateStatus{
		RequestID:  searchID,
		PipelineID: pipelineID,
		Progress:   status,
		Timestamp:  time.Now(),
	}
}

func (m *CreateMessage) persistPipelineResults(statusChan chan CreateStatus, client *Client, pipelineStorage model.PipelineStorage, dataStorage model.DataStorage, searchID string, dataset string, pipelineID string, resultID string, resultURI string) {
	// persist the completed state
	err := pipelineStorage.PersistPipeline(searchID, pipelineID, CompletedStatus, time.Now())
	if err != nil {
		// notify of error
		m.persistPipelineError(statusChan, client, pipelineStorage, searchID, pipelineID, err)
		return
	}
	// persist result metadata
	err = pipelineStorage.PersistPipelineResult(pipelineID, resultID, resultURI, CompletedStatus, time.Now())
	if err != nil {
		// notify of error
		m.persistPipelineError(statusChan, client, pipelineStorage, searchID, pipelineID, err)
		return
	}
	// persist results
	err = dataStorage.PersistResult(dataset, resultURI)
	if err != nil {
		// notify of error
		m.persistPipelineError(statusChan, client, pipelineStorage, searchID, pipelineID, err)
		return
	}
	// notify client of update
	statusChan <- CreateStatus{
		RequestID:  searchID,
		PipelineID: pipelineID,
		ResultID:   resultID,
		Progress:   CompletedStatus,
		Timestamp:  time.Now(),
	}
}

func (m *CreateMessage) dispatchPipeline(statusChan chan CreateStatus, client *Client, pipelineStorage model.PipelineStorage, dataStorage model.DataStorage, searchID string, pipelineID string, dataset string, datasetURI string) {

	// score pipeline
	pipelineScoreResponses, err := client.GeneratePipelineScores(context.Background(), pipelineID)
	if err != nil {
		m.persistPipelineError(statusChan, client, pipelineStorage, searchID, pipelineID, err)
		return
	}

	// persist the scores
	for _, response := range pipelineScoreResponses {
		for _, score := range response.Scores {
			err := pipelineStorage.PersistPipelineScore(pipelineID, score.Metric.Metric.String(), score.Value.GetDouble())
			if err != nil {
				m.persistPipelineError(statusChan, client, pipelineStorage, searchID, pipelineID, err)
				return
			}
		}
	}

	// fit pipeline
	_, err = client.GeneratePipelineFit(context.Background(), pipelineID)
	if err != nil {
		m.persistPipelineError(statusChan, client, pipelineStorage, searchID, pipelineID, err)
		return
	}

	// persist pipeline running status
	m.persistPipelineStatus(statusChan, client, pipelineStorage, searchID, pipelineID, RunningStatus)

	// generate predictions
	producePipelineRequest := m.createProducePipelineRequest(datasetURI, pipelineID)

	// generate predictions
	predictionResponses, err := client.GeneratePredictions(context.Background(), producePipelineRequest)
	if err != nil {
		m.persistPipelineError(statusChan, client, pipelineStorage, searchID, pipelineID, err)
		return
	}

	for _, response := range predictionResponses {

		if response.Progress != Progress_COMPLETED {
			// only persist completed responses
			continue
		}

		output, ok := response.ExposedOutputs[defaultExposedOutputKey]
		if !ok {
			m.persistPipelineError(statusChan, client, pipelineStorage, searchID, pipelineID, errors.Errorf("output is missing from response"))
			return
		}

		datasetURI, ok := output.Value.(*Value_DatasetUri)
		if !ok {
			m.persistPipelineError(statusChan, client, pipelineStorage, searchID, pipelineID, errors.Errorf("output is not of correct format"))
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
		m.persistPipelineResults(statusChan, client, pipelineStorage, dataStorage, searchID, dataset, pipelineID, resultID, resultURI)
	}
}

func (m *CreateMessage) createStatusChannels(client *Client, pipelines []*GetSearchPipelinesResultsResponse, pipelineStorage model.PipelineStorage, searchID string) []chan CreateStatus {

	// create channels

	// NOTE: WE BUFFER THE CHANNELS TO A SIZE OF 1 HERE SO WE CAN PERSIST BELOW
	// WITHOUT DEADLOCKING.
	var statusChannels []chan CreateStatus
	for range pipelines {
		statusChannels = append(statusChannels, make(chan CreateStatus, 1))
	}

	// persist all pipelines as pending

	// NOTE: we persist the pipelines here so that they exist in the DB when the
	// method returns.
	// NOTE: THE CHANNELS MUST BE BUFFERED TO A SIZE OF 1 OR ELSE THIS WILL DEADLOCK.
	for i, pipeline := range pipelines {
		m.persistPipelineStatus(statusChannels[i], client, pipelineStorage, searchID, pipeline.PipelineId, PendingStatus)
	}

	return statusChannels
}

// DispatchPipelines dispatches all pipeline requests
func (m *CreateMessage) DispatchPipelines(client *Client, pipelineStorage model.PipelineStorage, dataStorage model.DataStorage, searchID string, dataset string, datasetURI string) ([]chan CreateStatus, error) {

	pipelines, err := client.SearchPipelines(context.Background(), searchID)
	if err != nil {
		return nil, err
	}

	// create status channels and persist pipelines
	statusChannels := m.createStatusChannels(client, pipelines, pipelineStorage, searchID)

	// dispatch all pipelines
	go func() {

		wg := &sync.WaitGroup{}

		// dispatch individual pipelines
		for i, pipeline := range pipelines {

			// increment waitgroup
			wg.Add(1)

			go func(statusChan chan CreateStatus, pipelineID string) {
				m.dispatchPipeline(statusChan, client, pipelineStorage, dataStorage, searchID, pipelineID, dataset, datasetURI)
				wg.Done()
			}(statusChannels[i], pipeline.PipelineId)

		}

		// wait until all are complete
		wg.Wait()

		// end search
		client.EndSearch(context.Background(), searchID)
	}()

	return statusChannels, nil
}

// PersistAndDispatch persists the pipeline request and dispatches it.
func (m *CreateMessage) PersistAndDispatch(client *Client, pipelineStorage model.PipelineStorage, metaStorage model.MetadataStorage, dataStorage model.DataStorage) ([]chan CreateStatus, error) {

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

	// create search pipelines request
	searchRequest, err := m.createSearchPipelinesRequest(targetIndex)
	if err != nil {
		return nil, err
	}

	// start a pipeline searchID
	requestID, err := client.StartSearch(context.Background(), searchRequest)
	if err != nil {
		return nil, err
	}

	// persist the request
	err = pipelineStorage.PersistRequest(requestID, dataset.Metadata.Name, PendingStatus, time.Now())
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
		err = pipelineStorage.PersistRequestFeature(requestID, v, typ)
		if err != nil {
			return nil, err
		}
	}

	// store request filters
	err = pipelineStorage.PersistRequestFilters(requestID, m.Filters)
	if err != nil {
		return nil, err
	}

	// dispatch pipelines
	statusChannels, err := m.DispatchPipelines(client, pipelineStorage, dataStorage, requestID, dataset.Metadata.Name, datasetPath)
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
