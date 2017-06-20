package routes

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"goji.io/pattern"

	"github.com/unchartedsoftware/distil/api/util"
	"github.com/unchartedsoftware/distil/api/util/json"
)

func TestVariableSummariesHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := util.MockElasticResponse(t, []string{
		"./testdata/variables.json",
		"./testdata/variable_summaries_extrema.json",
		"./testdata/variable_summaries_numerical.json",
		"./testdata/variable_summaries_categorical.json",
	})

	// test index and dataset
	testIndex := "datasets"
	testDataset := "o_185"

	// put together a stub dataset request - need to manually account for goji's
	// parameter extraction
	req, err := http.NewRequest("GET", "/distil/variable_summaries/"+testIndex+"/"+testDataset, nil)
	assert.NoError(t, err)

	// add params
	ctx := req.Context()
	ctx = context.WithValue(ctx, pattern.Variable("index"), testIndex)
	ctx = context.WithValue(ctx, pattern.Variable("dataset"), testDataset)

	// add context to req
	req = req.WithContext(ctx)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res, err := util.TestElasticRoute(handler, req, VariableSummariesHandler)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"histograms":[
				{
					"name":"Number_seasons",
					"buckets":[
						{"key":"0", "count":1},
						{"key":"1", "count":0},
						{"key":"2", "count":0}
					]
				},
				{
					"name":"Games_played",
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
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}
