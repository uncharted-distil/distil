package routes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/unchartedsoftware/distil/api/util/json"
	"github.com/unchartedsoftware/distil/api/util/mock"
)

func TestVariableHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{
		"./testdata/variables.json",
	})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// put together a stub dataset request
	req := mock.HTTPRequest(t, "GET", "/distil/variables", map[string]string{
		"index":   "datasets",
		"dataset": "o_185",
	}, nil)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, VariablesHandler(ctor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"variables": [
				{"name":"d3mIndex","type":"integer","importance": 0,"role": "index","suggestedTypes": [{ "type": "integer", "probability": 1.00 }]},
				{"name":"Position","type":"categorical","importance": 0,"role": "attribute","suggestedTypes": [{ "type": "categorical", "probability": 1.00 }]},
				{"name":"Number_seasons","type":"integer","importance": 1,"role": "attribute","suggestedTypes": [ { "type": "integer", "probability": 1.00 }]},
				{"name":"Games_played","type":"integer","importance": 2,"role": "attribute","suggestedTypes": [ { "type": "integer", "probability": 1.00 }]},
				{"name":"On_base_pct","type":"float","importance": 3,"role": "attribute","suggestedTypes": [ { "type": "float", "probability": 1.00 }]}
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}
