package routes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/util/json"
	"github.com/unchartedsoftware/distil/api/util/mock"
)

func TestResultsSummaryHandlerInt(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{"./testdata/variables.json"})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// put together a stub pipeline request
	params := map[string]string{
		"dataset":    "o_185",
		"index":      "datasets",
		"result-uri": "./testdata/results.csv",
	}
	req := mock.HTTPRequest(t, "GET", "/distil/results-summary/", params, nil)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, ResultsSummaryHandler(ctor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"name": "Games_played",
			"type": "numerical",
			"extrema": {
				"min": 10,
				"max": 30
			},
			"buckets": [
				{"key": "10", "count": 2},
				{"key": "20", "count": 1},
				{"key": "30", "count": 1}
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestResultsSummaryHandlerFloat(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{"./testdata/variables.json"})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// put together a stub pipeline request
	params := map[string]string{
		"dataset":    "o_185",
		"index":      "datasets",
		"result-uri": "./testdata/results_float.csv",
	}
	req := mock.HTTPRequest(t, "GET", "/distil/results-summary/", params, nil)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, ResultsSummaryHandler(ctor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"name": "On_base_pct",
			"type": "numerical",
			"extrema": {
				"min": 0.1,
				"max": 0.3
			},
			"buckets": [
				{"key": "0.1", "count": 2},
				{"key": "0.2", "count": 1},
				{"key": "0.3", "count": 1}
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestResultsSummaryHandlerCategorical(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{"./testdata/variables.json"})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// put together a stub pipeline request
	params := map[string]string{
		"dataset":    "o_185",
		"index":      "datasets",
		"result-uri": "./testdata/results_categorical.csv",
	}
	req := mock.HTTPRequest(t, "GET", "/distil/results-summary/", params, nil)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, ResultsSummaryHandler(ctor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"name": "Position",
			"type": "categorical",
			"buckets": [
				{"key": "Pitcher", "count": 1},
				{"key": "Catcher", "count": 2},
				{"key": "First_base", "count": 1}
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
