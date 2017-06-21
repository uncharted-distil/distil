package routes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/unchartedsoftware/distil/api/util/mock"
)

func TestVariableSummariesHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{
		"./testdata/variables.json",
		"./testdata/variable_summaries_extrema.json",
		"./testdata/variable_summaries_numerical.json",
		"./testdata/variable_summaries_categorical.json",
	})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// mock http request
	req := mock.HTTPRequest(t, "GET", "/distil/variable_summaries/", map[string]string{
		"index":   "datasets",
		"dataset": "o_185",
	})

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, VariableSummariesHandler(ctor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected :=
		`{
			"histograms":[
				{
					"name":"Number_seasons",
					"extrema": {
						"min": 0,
						"max": 4
					},
					"buckets":[
						{"key":"0", "count":1},
						{"key":"1", "count":0},
						{"key":"2", "count":0}
					]
				},
				{
					"name":"Games_played",
					"extrema": {
						"min": 1,
						"max": 5
					},
					"buckets":[
						{"key":"1","count":1},
						{"key":"2","count":0},
						{"key":"3","count":3}
					]
				},
				{
					"name":"Player",
					"buckets":[
						{"key":"a","count":0},
						{"key":"b","count":0},
						{"key":"c","count":0}
					]
				}
			]
		}`
	actual := string(res.Body.Bytes())
	assert.JSONEq(t, expected, actual)
}
