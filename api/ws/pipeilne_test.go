package ws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/model"
)

func TestParseFilterParameters(t *testing.T) {
	expected := model.FilterParams{
		Size: 123,
		Filters: []*model.VariableFilter{
			model.NewNumericalFilter("feature_a", 0, 100),
			model.NewNumericalFilter("feature_b", 5, 500),
			model.NewCategoricalFilter("feature_c", []string{"alpha", "bravo", "charlie"}),
			model.NewEmptyFilter("feature_d"),
		},
	}
	rawMsg := []byte(
		`{
			"size": 123,
			"filters": [
				{ "name": "feature_a", "type": "numerical", "min": 0, "max": 100 },
				{ "name": "feature_b", "type": "numerical", "min": 5, "max": 500 },
				{ "name": "feature_c", "type": "categorical", "categories": ["alpha", "bravo", "charlie"] },
				{ "name": "feature_d", "type": "empty" }
			]
		}`)
	filterParams, err := model.ParseFilterParamsJSON(json.RawMessage(rawMsg))

	assert.NoError(t, err)
	assert.Equal(t, expected, *filterParams)
}
