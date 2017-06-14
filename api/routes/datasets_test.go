package routes

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/jeffail/gabs"
	"github.com/stretchr/testify/assert"

	"github.com/unchartedsoftware/distil/api/util"
)

func TestDatasetsHandler(t *testing.T) {
	// load ES result json the test server will return
	datasetJSON, err := ioutil.ReadFile("./testdata/datasets.json")
	assert.NoError(t, err)

	// mock elasticsearch request handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(datasetJSON)
	}

	// put together a stub dataset request
	request, err := http.NewRequest("GET", "/distil/datasets", nil)
	assert.NoError(t, err)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	result, err := util.TestElasticRoute(handler, request, DatasetsHandler)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected := `{
		"datasets":[
			{"name":"o_196"},
			{"name":"o_185"}
		]}`
	expectedJSON, err := gabs.ParseJSON([]byte(expected))
	assert.NoError(t, err)

	actualJSON, err := gabs.ParseJSON(result.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expectedJSON, actualJSON)
}

func TestDatasetsHandlerWithSearch(t *testing.T) {
	// load ES result json the test server will return
	datasetJSON, err := ioutil.ReadFile("./testdata/search.json")
	assert.NoError(t, err)

	// mock elasticsearch request handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(datasetJSON)
	}

	// put together a stub dataset request
	request, err := http.NewRequest("GET", "/distil/datasets?search=baseball", nil)
	assert.NoError(t, err)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	result, err := util.TestElasticRoute(handler, request, DatasetsHandler)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, result.Code)

	// compare expected and actual results - unmarshall first to ensure object
	// rather than byte equality
	expected := `{
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
	}`
	expectedJSON, err := gabs.ParseJSON([]byte(expected))
	assert.NoError(t, err)

	actualJSON, err := gabs.ParseJSON(result.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expectedJSON, actualJSON)
}
