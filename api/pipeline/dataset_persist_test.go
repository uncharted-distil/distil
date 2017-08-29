package pipeline

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/model"
)

func TestDatasetHashEqual(t *testing.T) {
	filterParams0 := model.FilterParams{
		Size: 0,
		Ranged: []model.VariableRange{
			model.VariableRange{Name: "feature_a", Min: 0, Max: 100},
		},
	}
	filterParams1 := model.FilterParams{
		Size: 0,
		Ranged: []model.VariableRange{
			model.VariableRange{Name: "feature_a", Min: 0, Max: 100},
		},
	}
	hash0, err := getFilteredDatasetHash("dataset", &filterParams0)
	hash1, err := getFilteredDatasetHash("dataset", &filterParams1)
	assert.NoError(t, err)
	assert.Equal(t, hash0, hash1)
}

func TestDatasetHashNotEqual(t *testing.T) {
	filterParams0 := model.FilterParams{
		Size: 0,
		Ranged: []model.VariableRange{
			model.VariableRange{Name: "feature_a", Min: 0, Max: 100},
		},
	}
	filterParams1 := model.FilterParams{
		Size: 1,
		Ranged: []model.VariableRange{
			model.VariableRange{Name: "feature_a", Min: 0, Max: 100},
		},
	}
	hash0, err := getFilteredDatasetHash("dataset", &filterParams0)
	hash1, err := getFilteredDatasetHash("dataset", &filterParams1)
	hash2, err := getFilteredDatasetHash("dataset_X", &filterParams0)
	assert.NoError(t, err)
	assert.NotEqual(t, hash0, hash1)
	assert.NotEqual(t, hash0, hash2)
}

func fetchFilteredData(t *testing.T) FilteredDataProvider {
	return func(dataset string, filters *model.FilterParams) (*model.FilteredData, error) {
		// basic sanity to check  params are passed through and parsed
		assert.Equal(t, 2, len(filters.Ranged))
		assert.Equal(t, "int_a", filters.Ranged[0].Name)
		assert.Equal(t, "float_b", filters.Ranged[1].Name)

		return &model.FilteredData{
			Name:     "test",
			Metadata: []*model.Variable{},
			Values: [][]interface{}{
				[]interface{}{0, 1.1, false, "test_1"},
				[]interface{}{2, 3.1245678, true, "test_2"},
			},
		}, nil
	}
}

func TestPersistFilteredData(t *testing.T) {
	defer os.RemoveAll("./test_output")

	// Stubbed out params - not actually applied to stub data
	filterParams := &model.FilterParams{
		Ranged: []model.VariableRange{
			model.VariableRange{Name: "int_a", Min: 0, Max: 100},
			model.VariableRange{Name: "float_b", Min: 5.0, Max: 500.0},
		},
	}

	// Verify that a new file is created from the call
	datasetPath, err := PersistFilteredData(fetchFilteredData(t), "./test_output", "test", filterParams)
	assert.NoError(t, err)
	assert.NotEqual(t, datasetPath, "")
	_, err = os.Stat(datasetPath)
	assert.False(t, os.IsNotExist(err))

	datasetPathUnmod, err := PersistFilteredData(fetchFilteredData(t), "./test_output", "test", filterParams)
	assert.Equal(t, datasetPath, datasetPathUnmod)

	// Verify that changed params results in a new file being used
	filterParamsMod := &model.FilterParams{
		Ranged: []model.VariableRange{
			model.VariableRange{Name: "int_a", Min: 0, Max: 100},
			model.VariableRange{Name: "float_b", Min: 10.0, Max: 11.0},
		},
	}
	datasetPathMod, err := PersistFilteredData(fetchFilteredData(t), "./test_output", "test", filterParamsMod)
	assert.NotEqual(t, datasetPath, datasetPathMod)
}
