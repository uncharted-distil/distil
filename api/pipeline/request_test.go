package pipeline

import (
	"testing"

	"github.com/mitchellh/hashstructure"
	"github.com/stretchr/testify/assert"
)

func TestHashPipelineCreateHashInclude(t *testing.T) {
	request := PipelineCreateRequest{
		&SessionContext{"12345"},
		[]*Feature{{"feature_alpha", "location"}},
		TaskType_CLASSIFICATION,
		TaskSubtype_NONE,
		"task description",
		OutputType_PROBABILITY,
		[]Metric{Metric_ACCURACY},
		[]*Feature{{"target_alpha", "location"}},
		5,
	}
	hash, _ := hashstructure.Hash(request, nil)

	request2 := PipelineCreateRequest{
		&SessionContext{"12345678910"},
		[]*Feature{{"feature_alpha", "location"}},
		TaskType_CLASSIFICATION,
		TaskSubtype_NONE,
		"task description",
		OutputType_PROBABILITY,
		[]Metric{Metric_ACCURACY},
		[]*Feature{{"target_alpha", "location"}},
		5,
	}
	hash2, _ := hashstructure.Hash(request2, nil)

	assert.Equal(t, hash, hash2)
}
