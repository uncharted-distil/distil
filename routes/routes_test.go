package routes

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeffail/gabs"
	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/plog"
	"goji.io/pattern"
	"gopkg.in/olivere/elastic.v2"
)

func testElasticRoute(
	elasticResponseHandler func(w http.ResponseWriter, r *http.Request),
	testRequest *http.Request,
	testRoute func(*elastic.Client) func(http.ResponseWriter, *http.Request),
) (*httptest.ResponseRecorder, error) {
	// create a new test server to handle elastic search rest requests
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		elasticResponseHandler(w, r)
	}))

	// create an elastic search rest client that will route requests to our test server
	client, err := elastic.NewSimpleClient(elastic.SetURL(testServer.URL))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// wrap the route under test as an http service
	routeHandler := http.HandlerFunc(testRoute(client))

	// forward the caller supplied test request to the route http service and return the recorded
	// results
	responseRec := httptest.NewRecorder()
	routeHandler.ServeHTTP(responseRec, testRequest)
	return responseRec, nil
}

func TestDatasetsHandler(t *testing.T) {
	// load ES result json the test server will return
	datasetJSON, err := ioutil.ReadFile("./testdata/datasets.json")
	assert.NoError(t, err)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(datasetJSON)
	}

	// put together a stub dataset request
	request, err := http.NewRequest("GET", "/distil/datasets", nil)
	assert.NoError(t, err)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	result, err := testElasticRoute(handler, request, DatasetsHandler)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, result.Code)

	// compare expected and acutal results - unmarshall first to ensure object rather than byte equality
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

func TestVariablesHandler(t *testing.T) {
	// load ES result json the test server will return
	datasetJSON, err := ioutil.ReadFile("./testdata/variables.json")
	assert.NoError(t, err)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(datasetJSON)
	}

	// put together a stub dataset request - need to manually account for goji's parameter extraction
	request, err := http.NewRequest("GET", "/distil/variables/o_185", nil)
	request = request.WithContext(context.WithValue(request.Context(), pattern.Variable("dataset"), "o_185"))
	assert.NoError(t, err)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	result, err := testElasticRoute(handler, request, VariablesHandler)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, result.Code)

	// compare expected and acutal results - unmarshall first to ensure object rather than byte equality
	expected := `{
		"variables":[
			{"name":"d3mIndex","type":"integer"},
			{"name":"Player","type":"categorical"},
			{"name":"Number_seasons","type":"integer"},
			{"name":"Games_played","type":"integer"}
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

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(datasetJSON)
	}

	// put together a stub dataset request
	request, err := http.NewRequest("GET", "/distil/datasets?search=baseball", nil)
	assert.NoError(t, err)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	result, err := testElasticRoute(handler, request, DatasetsHandler)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, result.Code)

	// compare expected and acutal results - unmarshall first to ensure object rather than byte equality
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
