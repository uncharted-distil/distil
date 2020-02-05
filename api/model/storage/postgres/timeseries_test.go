package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveDuplicates(t *testing.T) {
	data := [][]float64{
		[]float64{1, 10},
		[]float64{2, 20},
		[]float64{2, 30},
		[]float64{3, 40},
		[]float64{4, 50},
		[]float64{4, 60},
		[]float64{4, 70},
		[]float64{5, 80},
	}

	expected := [][]float64{
		[]float64{1, 10},
		[]float64{2, 50},
		[]float64{3, 40},
		[]float64{4, 180},
		[]float64{5, 80},
	}

	result := removeDuplicates(data)
	assert.Equal(t, expected, result)
}

func TestRemoveDuplicatesNoDuplicates(t *testing.T) {

	data := [][]float64{
		[]float64{1, 10},
		[]float64{2, 20},
		[]float64{3, 30},
		[]float64{4, 40},
	}
	result := removeDuplicates(data)
	assert.Equal(t, data, result)
}

func TestRemoveDuplicatesAllDuplicates(t *testing.T) {
	data := [][]float64{
		[]float64{1, 10},
		[]float64{1, 20},
		[]float64{1, 30},
		[]float64{1, 40},
	}

	expected := [][]float64{
		[]float64{1, 100},
	}

	result := removeDuplicates(data)
	assert.Equal(t, expected, result)
}
