package task

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-compute/pipeline"
	"github.com/unchartedsoftware/distil/api/env"
	apiModel "github.com/unchartedsoftware/distil/api/model"
)

type testSubmitter struct{}

func (testSubmitter) submit(datasetURIs []string, pipelineDesc *pipeline.PipelineDescription) (string, error) {
	return "file://test_data/result.csv", nil
}

func TestJoin(t *testing.T) {

	varsLeft := []*model.Variable{
		{
			Name:        "d3mIndex",
			DisplayName: "D3M Index",
			Type:        model.IntegerType,
		},
		{
			Name:        "alpha",
			DisplayName: "Alpha",
			Type:        model.FloatType,
		},
		{
			Name:        "bravo",
			DisplayName: "Bravo",
			Type:        model.IntegerType,
		},
	}

	varsRight := []*model.Variable{
		{
			Name:        "d3mIndex",
			DisplayName: "D3M Index",
			Type:        model.IntegerType,
		},
		{
			Name:        "charlie",
			DisplayName: "Charlie",
			Type:        model.CategoricalType,
		},
		{
			Name:        "delta",
			DisplayName: "Delta",
			Type:        model.IntegerType,
		},
	}

	cfg, err := env.LoadConfig()
	assert.NoError(t, err)

	cfg.TmpDataPath = "test_data"
	cfg.D3MInputDir = "test_data"

	result, err := joinPrimitive("file://test_data/test_1/datasetDoc.json", "file://test_data/test_2/datasetDoc.json",
		"alpha", "bravo", varsLeft, varsRight, testSubmitter{}, &cfg)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.ElementsMatch(t, result.Columns, []apiModel.Column{
		{
			Label: "D3M Index",
			Key:   "d3mIndex",
			Type:  model.IntegerType,
		},
		{
			Label: "Alpha",
			Key:   "alpha",
			Type:  model.FloatType,
		},
		{
			Label: "Charlie",
			Key:   "charlie",
			Type:  model.CategoricalType,
		},
	})

	expected := [][]interface{}{
		{int64(0), 1.0, "a"},
		{int64(1), 2.0, "b"},
		{int64(2), 3.0, "c"},
		{int64(3), 4.0, "d"},
	}
	assert.Equal(t, result.NumRows, 4)
	for i, row := range result.Values {
		assert.ElementsMatch(t, row, expected[i])
	}
}
