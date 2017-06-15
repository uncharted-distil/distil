package routes

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"goji.io/pattern"

	"github.com/unchartedsoftware/distil/api/util"
	"github.com/unchartedsoftware/distil/api/util/json"
)

func TestDatasetsHandler(t *testing.T) {
	// load ES result json the test server will return
	datasetJSON, err := ioutil.ReadFile("./testdata/datasets.json")
	assert.NoError(t, err)

	// mock elasticsearch request handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(datasetJSON)
	}

	// test index
	testIndex := "datasets"

	// put together a stub dataset request
	req, err := http.NewRequest("GET", "/distil/datasets/"+testIndex, nil)
	assert.NoError(t, err)

	// add params
	ctx := req.Context()
	ctx = context.WithValue(ctx, pattern.Variable("index"), testIndex)

	// add context to req
	req = req.WithContext(ctx)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res, err := util.TestElasticRoute(handler, req, DatasetsHandler)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"datasets":[
				{
					"name": "o_185",
					"description": "**Author**: Jeffrey S. Simonoff",
					"variables": [
						{"name":"d3mIndex","type":"integer"},
						{"name":"Player","type":"categorical"},
						{"name":"Number_seasons","type":"integer"},
						{"name":"Games_played","type":"integer"}
					]
				},
				{
					"name": "o_196",
					"description": "**Author**:",
					"variables": [
						{"name":"d3mIndex","type":"integer"},
						{"name":"cylinders","type":"categorical"},
						{"name":"displacement","type":"categorical"}
					]
				}
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestDatasetsHandlerWithSearch(t *testing.T) {
	// load ES result json the test server will return
	datasetJSON, err := ioutil.ReadFile("./testdata/search.json")
	assert.NoError(t, err)

	// mock elasticsearch request handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(datasetJSON)
	}

	// test index
	testIndex := "datasets"

	// put together a stub dataset request
	req, err := http.NewRequest("GET", "/distil/datasets/"+testIndex+"?search=baseball", nil)
	assert.NoError(t, err)

	// add params
	ctx := req.Context()
	ctx = context.WithValue(ctx, pattern.Variable("index"), testIndex)

	// add context to req
	req = req.WithContext(ctx)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res, err := util.TestElasticRoute(handler, req, DatasetsHandler)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and actual results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"datasets":[
				{
					"name": "o_185",
					"description": "**Author**: Jeffrey S. Simonoff",
					"variables": [
						{"name":"d3mIndex","type":"integer"},
						{"name":"Player","type":"categorical"},
						{"name":"Number_seasons","type":"integer"},
						{"name":"Games_played","type":"integer"}
					]
				},
				{
					"name": "o_196",
					"description": "**Author**:",
					"variables": [
						{"name":"d3mIndex","type":"integer"},
						{"name":"cylinders","type":"categorical"},
						{"name":"displacement","type":"categorical"}
					]
				}
			]
			}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}
