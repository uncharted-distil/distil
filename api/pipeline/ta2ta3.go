package pipeline

import (
	"context"
	"fmt"
	"strings"
	"path/filepath"

	"github.com/unchartedsoftware/distil/api/model"
)

const (
	defaultResourceID = "0"
	datasetDir        = "datasets"
)

// CreateMessage represents a create model message.
type CreateMessage struct {
	Dataset      string          `json:"dataset"`
	Index        string          `json:"index"`
	TargetFeature string    `json:"target"`
	Task         string          `json:"task"`
	MaxPipelines int32           `json:"maxPipelines"`
	Filters      *model.FilterParams    `json:"filters"`
	Metrics      []string        `json:"metric"`
}

// PipelineStatus represents a pipeline status.
type PipelineStatus struct {
	Progress Progress
	RequestID string
	PipelineID string
	Error error
}

func (m *CreateMessage) createSearchPipelinesRequest() (*SearchPipelinesRequest, error) {
	return &SearchPipelinesRequest{
		Problem: &ProblemDescription{
			Problem: &Problem{
				TaskType: convertTaskTypeFromTA3ToTA2(m.Task),
				PerformanceMetrics: convertMetricsFromTA3ToTA2(m.Metrics),
			},
			Inputs: []*ProblemInput{
				&ProblemInput{
					DatasetId: convertDatasetTA3ToTA2(m.Dataset),
					Targets: convertTargetFeaturseTA3ToTA2(m.TargetFeature),
				},
			},
		},
	}, nil
}

// func (m *CreateMessage) createProducePipelineRequest(datasetURI string) (*ProducePipelineRequest, error) {
// 	return &ProducePipelineRequest{
// 		PipelineId: pipeline.PipelineId,
// 		Inputs: []*Value{
// 			{
// 				Value: &Value_DatasetUri{
// 					DatasetUri: datasetURI,
// 				},
// 			},
// 		},
// 	}, nil
// }

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
	for _, f := range createMsg.TargetFeatures {
		err = pipelineStorage.PersistRequestFeature(requestID, m.TargetFeature, model.FeatureTypeTarget)
		if err != nil {
			return nil, err
		}
	}

	// store request filters
	err = pipelineStorage.PersistRequestFilters(requestID, filters)
	if err != nil {
		return nil, err
	}

	// dispatch pipelines
	statusChannels, err := client.DispatchPipelines(requestID, datasetPath)
	if err != nil {
		return nil, err
	}

	// TODO: finish this
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
			ColumnName: target,
			ResourceId:  defaultResourceID,
			ColumnIndex: 0, // TODO: fix this
			TargetIndex: 0, // TODO: what is this?
		},
	}
}

func convertDatasetTA3ToTA2(dataset string) string {
	return dataset
}
