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
				{"varName":"d3mIndex","varType":"integer","importance": 0,"varRole": "index","suggestedTypes": null},
				{"varName":"Position","varType":"categorical","importance": 0,"varRole": "attribute","suggestedTypes": null},
				{"varName":"Number_seasons","varType":"integer","importance": 1,"varRole": "attribute","suggestedTypes": null},
				{"varName":"Games_played","varType":"integer","importance": 2,"varRole": "attribute","suggestedTypes": null},
				{"varName":"On_base_pct","varType":"float","importance": 3,"varRole": "attribute","suggestedTypes": null}
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}
