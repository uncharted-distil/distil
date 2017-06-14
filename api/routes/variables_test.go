package routes

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"goji.io/pattern"

	"github.com/unchartedsoftware/distil/api/util"
)

func TestVariablesHandler(t *testing.T) {
	// load ES result json the test server will return
	datasetJSON, err := ioutil.ReadFile("./testdata/variables.json")
	assert.NoError(t, err)

	// mock elasticsearch request handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(datasetJSON)
	}

	// put together a stub dataset request - need to manually account for goji's
	// parameter extraction
	request, err := http.NewRequest("GET", "/distil/variables/o_185", nil)
	request = request.WithContext(
		context.WithValue(request.Context(),
			pattern.Variable("dataset"),
			"o_185"))
	assert.NoError(t, err)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	result, err := util.TestElasticRoute(handler, request, VariablesHandler)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, result.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
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
