package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveDuplicates(t *testing.T) {
	data := [][]float64{
		{1, 10},
		{2, 20},
		{2, 30},
		{3, 40},
		{4, 50},
		{4, 60},
		{4, 70},
		{5, 80},
	}

	expected := [][]float64{
		{1, 10},
		{2, 50},
		{3, 40},
		{4, 180},
		{5, 80},
	}

	result := removeDuplicates(data)
	assert.Equal(t, expected, result)
}

func TestRemoveDuplicatesNoDuplicates(t *testing.T) {

	data := [][]float64{
		{1, 10},
		{2, 20},
		{3, 30},
		{4, 40},
	}
	result := removeDuplicates(data)
	assert.Equal(t, data, result)
}

func TestRemoveDuplicatesAllDuplicates(t *testing.T) {
	data := [][]float64{
		{1, 10},
		{1, 20},
		{1, 30},
		{1, 40},
	}

	expected := [][]float64{
		{1, 100},
	}

	result := removeDuplicates(data)
	assert.Equal(t, expected, result)
}
