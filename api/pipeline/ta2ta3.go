package pipeline

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/unchartedsoftware/distil/api/model"
)

const (
	defaultResourceID = "0"
	datasetDir        = "datasets"
)

// CreateMessage represents a create model message.
type CreateMessage struct {
	Dataset       string              `json:"dataset"`
	Index         string              `json:"index"`
	TargetFeature string              `json:"target"`
	Task          string              `json:"task"`
	MaxPipelines  int32               `json:"maxPipelines"`
	Filters       *model.FilterParams `json:"filters"`
	Metrics       []string            `json:"metric"`
}

// PipelineStatus represents a pipeline status.
type PipelineStatus struct {
	Progress   Progress
	RequestID  string
	PipelineID string
	Error      error
	Timestamp time.Time
}

func (m *CreateMessage) createSearchPipelinesRequest() (*SearchPipelinesRequest, error) {
	return &SearchPipelinesRequest{
		Problem: &ProblemDescription{
			Problem: &Problem{
				TaskType:           convertTaskTypeFromTA3ToTA2(m.Task),
				PerformanceMetrics: convertMetricsFromTA3ToTA2(m.Metrics),
			},
			Inputs: []*ProblemInput{
				&ProblemInput{
					DatasetId: convertDatasetTA3ToTA2(m.Dataset),
					Targets:   convertTargetFeaturseTA3ToTA2(m.TargetFeature),
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

func (m *CreateMessage) dispatchPipeline(statusChan chan PipelineStatus, client *Client, pipelineStorage model.PipelineStorage, searchID string, pipelineID string, datasetURI string) {

	// notify that the pipeline is pending
	statusChan <- PipelineStatus{
		RequestID:  searchID,
		PipelineID: pipelineID,
		Progress:   Progress_PENDING,
		Timestamp: time.Now(),
	}

	// score pipeline
	pipelineScoreResponses, err := client.GeneratePipelineScores(context.Background(), pipelineID)
	if err != nil {
		statusChan <- PipelineStatus{
			RequestID:  searchID,
			PipelineID: pipelineID,
			Progress:   Progress_ERRORED,
			Error:      err,
			Timestamp: time.Now(),
		}
		return
	}

	// persist the scores
	for _, response := range pipelineScoreResponses {
		for _, score := range response.Scores {
			err := pipelineStorage.PersistPipelineScore(pipelineID, score.Metric.Metric.String(), score.Value.GetDouble())
			if err != nil {
				statusChan <- PipelineStatus{
					RequestID:  searchID,
					PipelineID: pipelineID,
					Progress:   Progress_ERRORED,
					Error:      err,
					Timestamp: time.Now(),
				}
				return
			}
		}
	}

	// fit pipeline
	_, err = client.GeneratePipelineFit(context.Background(), pipelineID)
	if err != nil {
		statusChan <- PipelineStatus{
			RequestID:  searchID,
			PipelineID: pipelineID,
			Progress:   Progress_ERRORED,
			Error:      err,
			Timestamp: time.Now(),
		}
		return
	}

	// notify that the pipeline is running
	statusChan <- PipelineStatus{
		RequestID:  searchID,
		PipelineID: pipelineID,
		Progress:   Progress_RUNNING,
		Timestamp: time.Now(),
	}

	// generate predictions
	producePipelineRequest := m.createProducePipelineRequest(datasetURI, pipelineID)

	// generate predictions
	predictionResponses, err := client.GeneratePredictions(context.Background(), producePipelineRequest)
	if err != nil {
		statusChan <- PipelineStatus{
			RequestID:  searchID,
			PipelineID: pipelineID,
			Progress:   Progress_ERRORED,
			Error:      err,
			Timestamp: time.Now(),
		}
		return
	}

	for _, response := range predictionResponses {
		/*
		// remove the protocol portion if it exists. The returned value is either a
		// csv file or a directory.
		resultURI = strings.Replace(resultURI, "file://", "", 1)
		if !strings.HasSuffix(resultURI, ".csv") {
			resultURI = path.Join(resultURI, D3MLearningData)
		}

		// get the result UUID. NOTE: Doing sha1 for now.
		hasher := sha1.New()
		hasher.Write([]byte(resultURI))
		bs := hasher.Sum(nil)
		resultID = fmt.Sprintf("%x", bs)

		// persist predictions
		err := pipelineStorage.PersistPipelineResult(pipelineID, resultID, resultURI, Progress_COMPLETED, time.Now())
		if err != nil {
			statusChan <- PipelineStatus{
				RequestID:  searchID,
				PipelineID: pipelineID,
				Progress:   Progress_ERRORED,
				Error:      err,
				Timestamp: time.Now(),
			}
			return
		}
		*/
	}

	statusChan <- PipelineStatus{
		RequestID:  searchID,
		PipelineID: pipelineID,
		Progress:   Progress_COMPLETED,
		Timestamp: time.Now(),
	}
}

// DispatchPipelines dispatches all pipeline requests.
func (m *CreateMessage) DispatchPipelines(client *Client, pipelineStorage model.PipelineStorage, searchID string, datasetURI string) ([]chan PipelineStatus, error) {

	pipelines, err := client.SearchPipelines(context.Background(), searchID)
	if err != nil {
		return nil, err
	}

	// create status channels
	var statusChannels []chan PipelineStatus
	for range pipelines {
		statusChannels = append(statusChannels, make(chan PipelineStatus))
	}

	wg := &sync.WaitGroup{}

	// dispatch all pipelines
	go func() {
		for i, pipeline := range pipelines {
			statusChan := statusChannels[i]
			wg.Add(1)
			go func(pipelineID string) {
				m.dispatchPipeline(statusChan, client, pipelineStorage, searchID, pipelineID, datasetURI)
				wg.Done()
			}(pipeline.PipelineId)

		}
		// end search
		client.EndSearch(context.Background(), searchID)
	}()

	return statusChannels, nil
}

// PersistAndDispatch persists the pipeline request and dispatches it.
func (m *CreateMessage) PersistAndDispatch(client *Client, pipelineStorage model.PipelineStorage, metaStorage model.MetadataStorage, dataStorage model.DataStorage) ([]chan PipelineStatus, error) {

	// create search pipelines request
	searchRequest, err := m.createSearchPipelinesRequest()
	if err != nil {
		return nil, err
	}

	// start a pipeline searchID
	requestID, err := client.StartSearch(context.Background(), searchRequest)
	if err != nil {
		return nil, err
	}

	// fetch the queried dataset
	dataset, err := model.FetchDataset(m.Dataset, m.Index, true, m.Filters, metaStorage, dataStorage)
	if err != nil {
		return nil, err
	}

	// perist the dataset and get URI
	datasetPath, err := PersistFilteredData(datasetDir, m.TargetFeature, dataset)
	if err != nil {
		return nil, err
	}
	// make sure the path is absolute and contains the URI prefix
	datasetPath, err = filepath.Abs(datasetPath)
	if err != nil {
		return nil, err
	}
	datasetPath = fmt.Sprintf("%s", filepath.Join(datasetPath, D3MDataSchema))

	// store the request features
	for _, v := range m.Filters.Variables {
		err = pipelineStorage.PersistRequestFeature(requestID, v, model.FeatureTypeTrain)
		if err != nil {
			return nil, err
		}
	}

	// store target feature
	err = pipelineStorage.PersistRequestFeature(requestID, m.TargetFeature, model.FeatureTypeTarget)
	if err != nil {
		return nil, err
	}

	// store request filters
	err = pipelineStorage.PersistRequestFilters(requestID, m.Filters)
	if err != nil {
		return nil, err
	}

	// dispatch pipelines
	statusChannels, err := m.DispatchPipelines(client, pipelineStorage, requestID, datasetPath)
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

func convertTargetFeaturseTA3ToTA2(target string) []*ProblemTarget {
	return []*ProblemTarget{
		&ProblemTarget{
			ColumnName:  target,
			ResourceId:  defaultResourceID,
			ColumnIndex: 0, // TODO: fix this
			TargetIndex: 0, // TODO: what is this?
		},
	}
}

func convertDatasetTA3ToTA2(dataset string) string {
	return dataset
}
