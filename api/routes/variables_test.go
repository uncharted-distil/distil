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

func TestVariablesHandler(t *testing.T) {
	// load ES result json the test server will return
	datasetJSON, err := ioutil.ReadFile("./testdata/variables.json")
	assert.NoError(t, err)

	// mock elasticsearch request handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(datasetJSON)
	}

	// test index and dataset
	testIndex := "datasets"
	testDataset := "o_185"

	// put together a stub dataset request - need to manually account for goji's
	// parameter extraction
	req, err := http.NewRequest("GET", "/distil/variables/"+testIndex+"/"+testDataset, nil)
	assert.NoError(t, err)

	// add params
	ctx := req.Context()
	ctx = context.WithValue(ctx, pattern.Variable("index"), testIndex)
	ctx = context.WithValue(ctx, pattern.Variable("dataset"), testDataset)

	// add context to req
	req = req.WithContext(ctx)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res, err := util.TestElasticRoute(handler, req, VariablesHandler)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"variables":[
				{"name":"d3mIndex","type":"integer"},
				{"name":"Player","type":"categorical"},
				{"name":"Number_seasons","type":"integer"},
				{"name":"Games_played","type":"integer"}
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}
