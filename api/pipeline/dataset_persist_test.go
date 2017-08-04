package pipeline

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/model"
)

func TestParseFilterParameters(t *testing.T) {
	expected := model.FilterParams{
		Size: datasetSizeLimit,
		Ranged: []model.VariableRange{
			model.VariableRange{Name: "feature_a", Min: 0, Max: 100},
			model.VariableRange{Name: "feature_b", Min: 5, Max: 500},
		},
		Categorical: []model.VariableCategories{
			model.VariableCategories{Name: "feature_c", Categories: []string{"alpha", "bravo", "charlie"}},
		},
		None: []string{"feature_d"},
	}

	rawMsg := []byte(
		`{
			"feature_a": { "name": "feature_a", "type": "numerical", "min": 0, "max": 100, "enabled": true},
			"feature_b": { "name": "feature_b", "type": "numerical", "min": 5, "max": 500.0, "enabled": true},
			"feature_c": { "name": "feature_c", "type": "categorical", "categories": ["alpha", "bravo", "charlie"], "enabled": true},
			"feature_d": { "name": "feature_d", "enabled": false}
		}`)
	filterParams, err := parseDatasetFilters(json.RawMessage(rawMsg))

	assert.NoError(t, err)
	assert.Equal(t, expected, *filterParams)
}

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
	rawFilterParams := json.RawMessage([]byte(
		`{
			"int_a": { "name": "int_a", "type": "numerical", "min": 0, "max": 100, "enabled": true},
			"float_b": { "name": "float_b", "type": "numerical", "min": 5, "max": 500.0, "enabled": true}
		}`))

	// Verify that a new file is created from the call
	datasetPath, err := PersistFilteredData(fetchFilteredData(t), "./test_output", "test", rawFilterParams)
	assert.NoError(t, err)
	assert.NotEqual(t, datasetPath, "")
	_, err = os.Stat(datasetPath)
	assert.False(t, os.IsNotExist(err))

	datasetPathUnmod, err := PersistFilteredData(fetchFilteredData(t), "./test_output", "test", rawFilterParams)
	assert.Equal(t, datasetPath, datasetPathUnmod)

	// Verify that changed params results in a new file being used
	rawFilterParamsMod := json.RawMessage([]byte(
		`{
			"int_a": { "name": "int_a", "type": "numerical", "min": 0, "max": 100, "enabled": true},
			"float_b": { "name": "float_b", "type": "numerical", "min": 10, "max": 11, "enabled": true}
		}`))
	datasetPathMod, err := PersistFilteredData(fetchFilteredData(t), "./test_output", "test", rawFilterParamsMod)
	assert.NotEqual(t, datasetPath, datasetPathMod)
}
