package pipeline

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/model"
)

func TestDatasetHashEqual(t *testing.T) {
	filterParams0 := model.FilterParams{
		Size: 0,
		Ranged: []model.VariableRange{
			{Name: "feature_a", Min: 0, Max: 100},
		},
	}
	filterParams1 := model.FilterParams{
		Size: 0,
		Ranged: []model.VariableRange{
			{Name: "feature_a", Min: 0, Max: 100},
		},
	}
	hash0, err := getFilteredDatasetHash("dataset", "target", &filterParams0)
	hash1, err := getFilteredDatasetHash("dataset", "target", &filterParams1)
	assert.NoError(t, err)
	assert.Equal(t, hash0, hash1)
}

func TestDatasetHashNotEqual(t *testing.T) {
	filterParams0 := model.FilterParams{
		Size: 0,
		Ranged: []model.VariableRange{
			{Name: "feature_a", Min: 0, Max: 100},
		},
	}
	filterParams1 := model.FilterParams{
		Size: 1,
		Ranged: []model.VariableRange{
			{Name: "feature_a", Min: 0, Max: 100},
		},
	}
	hash0, err := getFilteredDatasetHash("dataset", "target", &filterParams0)
	hash1, err := getFilteredDatasetHash("dataset", "target", &filterParams1)
	hash2, err := getFilteredDatasetHash("dataset_X", "target", &filterParams0)
	hash3, err := getFilteredDatasetHash("dataset", "target_X", &filterParams1)
	assert.NoError(t, err)
	assert.NotEqual(t, hash0, hash1)
	assert.NotEqual(t, hash0, hash2)
	assert.NotEqual(t, hash0, hash3)
}

func fetchFilteredData(t *testing.T) FilteredDataProvider {
	return func(dataset string, index string, filters *model.FilterParams, inclusive bool) (*model.FilteredData, error) {
		// basic sanity to check  params are passed through and parsed
		assert.Equal(t, 2, len(filters.Ranged))
		assert.Equal(t, "int_a", filters.Ranged[0].Name)
		assert.Equal(t, "float_b", filters.Ranged[1].Name)

		return &model.FilteredData{
			Name: "test",
			Columns: []string{
				"feature0",
				"feature1",
				"feature2",
				"feature3",
			},
			Types: []string{
				"integer",
				"float",
				"boolean",
				"string",
			},
			Values: [][]interface{}{
				{0, 1.1, false, "test_1"},
				{2, 3.1245678, true, "test_2"},
				{4, 3.1245678, true, "test_3"},
			},
		}, nil
	}
}

func TestPersistFilteredData(t *testing.T) {
	defer os.RemoveAll("./test_output")

	// Stubbed out params - not actually applied to stub data
	filterParams := &model.FilterParams{
		Ranged: []model.VariableRange{
			{Name: "int_a", Min: 0, Max: 100},
			{Name: "float_b", Min: 5.0, Max: 500.0},
		},
	}

	// Verify that a new file is created from the call
	datasetPath, err := PersistFilteredData(fetchFilteredData(t), "./test_output", "test", "test", "feature1", filterParams, true)
	assert.NoError(t, err)
	assert.NotEqual(t, datasetPath, "")
	_, err = os.Stat(path.Join(datasetPath, D3MTrainData))
	assert.False(t, os.IsNotExist(err))

	_, err = os.Stat(path.Join(datasetPath, D3MTrainTargets))
	assert.False(t, os.IsNotExist(err))

	datasetPathUnmod, err := PersistFilteredData(fetchFilteredData(t), "./test_output", "test", "test", "feature1", filterParams, true)
	assert.Equal(t, datasetPath, datasetPathUnmod)

	// Verify that changed params results in a new file being used
	filterParamsMod := &model.FilterParams{
		Ranged: []model.VariableRange{
			{Name: "int_a", Min: 0, Max: 100},
			{Name: "float_b", Min: 10.0, Max: 11.0},
		},
	}
	datasetPathMod, err := PersistFilteredData(fetchFilteredData(t), "./test_output", "test", "test", "feature1", filterParamsMod, true)
	assert.NotEqual(t, datasetPath, datasetPathMod)
}
