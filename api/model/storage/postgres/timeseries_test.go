package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	api "github.com/uncharted-distil/distil/api/model"
)

func TestRemoveDuplicates(t *testing.T) {
	data := []*api.TimeseriesObservation{
		{Time: 1, Value: 10},
		{Time: 2, Value: 20},
		{Time: 2, Value: 30},
		{Time: 3, Value: 40},
		{Time: 4, Value: 50},
		{Time: 4, Value: 60},
		{Time: 4, Value: 70},
		{Time: 5, Value: 80},
	}

	expected := []*api.TimeseriesObservation{
		{Time: 1, Value: 10},
		{Time: 2, Value: 50},
		{Time: 3, Value: 40},
		{Time: 4, Value: 180},
		{Time: 5, Value: 80},
	}

	result := removeDuplicates(data)
	assert.Equal(t, expected, result)
}

func TestRemoveDuplicatesNoDuplicates(t *testing.T) {

	data := []*api.TimeseriesObservation{
		{Time: 1, Value: 10},
		{Time: 2, Value: 20},
		{Time: 3, Value: 30},
		{Time: 4, Value: 40},
	}
	result := removeDuplicates(data)
	assert.Equal(t, data, result)
}

func TestRemoveDuplicatesAllDuplicates(t *testing.T) {

	data := []*api.TimeseriesObservation{
		{Time: 1, Value: 10},
		{Time: 1, Value: 20},
		{Time: 1, Value: 30},
		{Time: 1, Value: 40},
	}

	expected := []*api.TimeseriesObservation{
		{Time: 1, Value: 100},
	}

	result := removeDuplicates(data)
	assert.Equal(t, expected, result)
}
