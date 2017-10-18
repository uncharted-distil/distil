package routes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/util/mock"
)

func TestParseSearchParams(t *testing.T) {
	params := map[string]string{
		"dataset": "o_185",
	}
	query := map[string]string{
		"On_base_pct": "numerical,0,100",
		"Position":    "categorical,Catcher,Pitcher",
		"Triples":     "",
		"size":        "100",
	}
	req := mock.HTTPRequest(t, "GET", "/distil/data", params, query)
	// parse
	filters, err := ParseFilterParams(req)
	assert.NoError(t, err)
	assert.Equal(t, filters.Ranged[0].Name, "On_base_pct")
	assert.Equal(t, filters.Ranged[0].Min, 0.0)
	assert.Equal(t, filters.Ranged[0].Max, 100.0)
	assert.Equal(t, filters.Categorical[0].Name, "Position")
	assert.Equal(t, filters.Categorical[0].Categories, []string{"Catcher", "Pitcher"})
	assert.Equal(t, filters.None[0], "Triples")
	assert.Equal(t, filters.Size, 100)
}

func TestParseSearchParamsMalformed(t *testing.T) {
	params := map[string]string{
		"dataset": "o_185",
	}
	query0 := map[string]string{
		"On_base_pct": "numerical,0",
	}
	query1 := map[string]string{
		"Position": "categorical",
	}
	query2 := map[string]string{
		"Triples": "numerical,1,2,3",
	}
	query3 := map[string]string{
		"size": "",
	}
	// missing max
	req := mock.HTTPRequest(t, "GET", "/distil/data", params, query0)
	filters, err := ParseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// missing categories
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query1)
	filters, err = ParseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// extra param
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query2)
	filters, err = ParseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// missing size
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query3)
	filters, err = ParseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
}
