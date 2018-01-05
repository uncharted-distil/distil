package pipeline

import (
	"testing"

	"github.com/mitchellh/hashstructure"
	"github.com/stretchr/testify/assert"
)

func TestHashPipelineCreateHashInclude(t *testing.T) {
	request := PipelineCreateRequest{
		Context:         &SessionContext{"12345"},
		DatasetUri:      "location",
		Task:            TaskType_CLASSIFICATION,
		TaskSubtype:     TaskSubtype_NONE,
		TaskDescription: "task description",
		Metrics:         []PerformanceMetric{PerformanceMetric_ACCURACY},
		PredictFeatures: []*Feature{{"feature_alpha", "location"}},
		TargetFeatures:  []*Feature{{"target_alpha", "location"}},
		MaxPipelines:    5,
	}
	hash, _ := hashstructure.Hash(request, nil)

	request2 := PipelineCreateRequest{
		Context:         &SessionContext{"12345678910"},
		DatasetUri:      "location",
		Task:            TaskType_CLASSIFICATION,
		TaskSubtype:     TaskSubtype_NONE,
		TaskDescription: "task description",
		Metrics:         []PerformanceMetric{PerformanceMetric_ACCURACY},
		PredictFeatures: []*Feature{{"feature_alpha", "location"}},
		TargetFeatures:  []*Feature{{"target_alpha", "location"}},
		MaxPipelines:    5,
	}
	hash2, _ := hashstructure.Hash(request2, nil)

	assert.Equal(t, hash, hash2)
}
