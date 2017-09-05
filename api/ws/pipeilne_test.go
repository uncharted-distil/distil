package ws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/model"
)

func TestParseFilterParameters(t *testing.T) {
	expected := model.FilterParams{
		Size: datasetSizeLimit,
		Ranged: []model.VariableRange{
			{Name: "feature_a", Min: 0, Max: 100},
			{Name: "feature_b", Min: 5, Max: 500},
		},
		Categorical: []model.VariableCategories{
			{Name: "feature_c", Categories: []string{"alpha", "bravo", "charlie"}},
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
