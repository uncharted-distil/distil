package routes

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/unchartedsoftware/distil/api/util/json"

	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/util"
	"goji.io/pattern"
)

func TestParseSearchParams(t *testing.T) {
	req, err := http.NewRequest("GET", "/distil/data/o185"+
		"?On_base_pct=integer,0,100&Position=categorical,Catcher,Pitcher&Triples&size=100", nil)
	assert.NoError(t, err)
	searchParams := parseSearchParams(req)
	assert.Equal(t, searchParams.ranged[0].Name, "On_base_pct")
	assert.Equal(t, searchParams.ranged[0].Type, "integer")
	assert.Equal(t, searchParams.ranged[0].min, 0.0)
	assert.Equal(t, searchParams.ranged[0].max, 100.0)
	assert.Equal(t, searchParams.categorical[0].Name, "Position")
	assert.Equal(t, searchParams.categorical[0].Type, "categorical")
	assert.Equal(t, searchParams.categorical[0].categories, []string{"Catcher", "Pitcher"})
	assert.Equal(t, searchParams.none[0], "Triples")
	assert.Equal(t, searchParams.size, 100)
}

func TestParseSearchParamsMalformed(t *testing.T) {
	req, err := http.NewRequest("GET", "/distil/data/o185"+
		"?On_base_pct=integer,0&Position=categorical,Catcher&Triples=integer,1,2,3&size", nil)
	assert.NoError(t, err)
	searchParams := parseSearchParams(req)
	assert.Equal(t, searchParams.categorical[0].Name, "Position")
	assert.Equal(t, searchParams.categorical[0].Type, "categorical")
	assert.Equal(t, searchParams.categorical[0].categories, []string{"Catcher"})
	assert.Equal(t, len(searchParams.ranged), 0)
	assert.Equal(t, searchParams.size, defaultSearchSize)
}

func TestFilteredDataHandler(t *testing.T) {
	// load ES result json the test server will return
	dataJSON, err := ioutil.ReadFile("./testdata/filtered_data.json")
	assert.NoError(t, err)

	// mock elasticsearch request handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(dataJSON)
	}

	// test index and dataset
	testDataset := "o_185"

	// put together a stub dataset request - need to manually account for goji's
	// parameter extraction
	req, err := http.NewRequest("GET", "/distil/filtered_data/"+testDataset+
		"?On_base_pct=0,100&Position=Catcher&Triples", nil)
	assert.NoError(t, err)

	// add params
	ctx := req.Context()
	ctx = context.WithValue(ctx, pattern.Variable("dataset"), testDataset)

	// add context to req
	req = req.WithContext(ctx)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res, err := util.TestElasticRoute(handler, req, FilteredDataHandler)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"name": "o_185",
			"metadata": [
				{"name": "On_base_pct", "type": "float"},
				{"name": "Position", "type": "categorical"},
				{"name": "Triples", "type": "integer"}
			],
			"values": [
				[0.268, "Catcher", 5],
				[0.244, "Catcher", 1]
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}
