package routes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/util/json"
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
	filters, err := parseFilterParams(req)
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
	filters, err := parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// missing categories
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query1)
	filters, err = parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// extra param
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query2)
	filters, err = parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// missing size
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query3)
	filters, err = parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
}

func TestFilteredDataHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{"./testdata/filtered_data.json"})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// instantiate storage filter client constructor.
	storageCtor := filter.NewElasticFilter(esClientCtor)

	// put together a stub dataset request
	params := map[string]string{
		"dataset": "o_185",
	}
	query := map[string]string{
		"On_base_pct": "numerical,0,100",
		"Position":    "categorical,Catcher",
		"Triples":     "",
	}
	req := mock.HTTPRequest(t, "GET", "/distil/data", params, query)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, FilteredDataHandler(storageCtor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"name": "o_185",
			"metadata": [
				{"name": "On_base_pct", "type": "float"},
				{"name": "Position", "type": "categorical"},
				{"name": "Triples", "type": "integer"}
			],
			"values": [
				[0.268, "Catcher", 5],
				[0.244, "Catcher", 1]
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
