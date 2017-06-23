package routes

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/util/json"
	"github.com/unchartedsoftware/distil/api/util/mock"
	"goji.io/pattern"
)

func TestParseSearchParams(t *testing.T) {
	req, err := http.NewRequest("GET", "/distil/data/o185"+
		"?On_base_pct=integer,0,100&Position=categorical,Catcher,Pitcher&Triples&size=100", nil)
	assert.NoError(t, err)
	filterParams, err := parseFilterParams(req)
	assert.NoError(t, err)
	assert.Equal(t, filterParams.Ranged[0].Name, "On_base_pct")
	assert.Equal(t, filterParams.Ranged[0].Type, "integer")
	assert.Equal(t, filterParams.Ranged[0].Min, 0.0)
	assert.Equal(t, filterParams.Ranged[0].Max, 100.0)
	assert.Equal(t, filterParams.Categorical[0].Name, "Position")
	assert.Equal(t, filterParams.Categorical[0].Type, "categorical")
	assert.Equal(t, filterParams.Categorical[0].Categories, []string{"Catcher", "Pitcher"})
	assert.Equal(t, filterParams.None[0], "Triples")
	assert.Equal(t, filterParams.Size, 100)
}

func TestParseSearchParamsMalformed(t *testing.T) {
	req, err := http.NewRequest("GET", "/distil/data/o185"+
		"?On_base_pct=integer,0&Position=categorical,Catcher&Triples=integer,1,2,3&size", nil)
	assert.NoError(t, err)
	filterParams, err := parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filterParams)
}

func TestFilteredDataHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{"./testdata/filtered_data.json"})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// test index and dataset
	testDataset := "o_185"

	// put together a stub dataset request - need to manually account for goji's
	// parameter extraction
	req, err := http.NewRequest("GET", "/distil/filtered_data/"+testDataset+
		"?On_base_pct=integer,0,100&Position=categorical,Catcher&Triples", nil)
	assert.NoError(t, err)

	// add params
	ctx := req.Context()
	ctx = context.WithValue(ctx, pattern.Variable("dataset"), testDataset)

	// add context to req
	req = req.WithContext(ctx)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, FilteredDataHandler(ctor))
	assert.NoError(t, err)

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
