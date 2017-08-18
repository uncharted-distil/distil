package pipeline

import (
	"testing"

	"github.com/mitchellh/hashstructure"
	"github.com/stretchr/testify/assert"
)

func TestHashPipelineCreateHashInclude(t *testing.T) {
	request := PipelineCreateRequest{
		&SessionContext{"12345"},
		[]string{"location"},
		Task_CLASSIFICATION,
		"task description",
		Output_PROBABILITY,
		[]Metric{Metric_ACCURACY},
		[]string{"feature"},
		5,
	}
	hash, _ := hashstructure.Hash(request, nil)

	request2 := PipelineCreateRequest{
		&SessionContext{"123456"},
		[]string{"location"},
		Task_CLASSIFICATION,
		"task description",
		Output_PROBABILITY,
		[]Metric{Metric_ACCURACY},
		[]string{"feature"},
		5,
	}
	hash2, _ := hashstructure.Hash(request2, nil)

	assert.Equal(t, hash, hash2)
}
