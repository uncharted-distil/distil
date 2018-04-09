package pipeline

import (
	"strings"

	"github.com/unchartedsoftware/distil/api/model"
)

const (
	defaultResourceID = "0"
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

// ConvertTA3ToTA2 converts between our internal create model message and the
// TA2 API.
func (m *CreateMessage) ConvertTA3ToTA2() (*SearchPipelinesRequest, error) {
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
