package routes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/unchartedsoftware/distil/api/util/mock"
)

func TestDatasetsHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{
		"./testdata/datasets.json",
	})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// mock http request
	req := mock.HTTPRequest(t, "GET", "/distil/datasets/", map[string]string{
		"index": "datasets",
	})

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, DatasetsHandler(ctor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected :=
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
		}`
	actual := string(res.Body.Bytes())
	assert.JSONEq(t, expected, actual)
}

func TestDatasetsHandlerWithSearch(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{
		"./testdata/search.json",
	})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// mock http request
	req := mock.HTTPRequest(t, "GET", "/distil/datasets?search=baseball", map[string]string{
		"index": "datasets",
	})

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, DatasetsHandler(ctor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and actual results - unmarshall first to ensure object
	// rather than byte equality
	expected :=
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
		}`
	actual := string(res.Body.Bytes())
	assert.JSONEq(t, expected, actual)
}
