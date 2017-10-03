package routes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/model/storage/elastic"
	"github.com/unchartedsoftware/distil/api/util/json"
	"github.com/unchartedsoftware/distil/api/util/mock"
)

func TestResultsHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{"./testdata/variables.json"})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// instantiate storage client constructor.
	storageCtor := elastic.NewStorage(ctor)

	// put together a stub pipeline request
	params := map[string]string{
		"dataset":      "o_185",
		"index":        "datasets",
		"results-uuid": "./testdata/results.csv",
		"inclusive":    "inclusive",
	}
	req := mock.HTTPRequest(t, "GET", "/distil/results/", params, nil)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, ResultsHandler(storageCtor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"name": "o_185",
			"columns": [
				"Games_played"
			],
			"types": [
				"integer"
			],
			"values": [
				[10],
				[20],
				[30],
				[10]
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
