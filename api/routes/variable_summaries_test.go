package routes

import (
	"net/http"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/unchartedsoftware/distil/api/model/storage/elastic"
	"github.com/unchartedsoftware/distil/api/model/storage/postgres"
	"github.com/unchartedsoftware/distil/api/util/json"
	"github.com/unchartedsoftware/distil/api/util/mock"
)

func TestVariableSummaryHandlerNumerical(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{
		"./testdata/variables.json",
		"./testdata/variable_summaries_extrema.json",
		"./testdata/variable_summaries_numerical.json",
	})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// instantiate storage client constructor.
	storageCtor := elastic.NewStorage(ctor)

	// put together a stub dataset request
	req := mock.HTTPRequest(t, "GET", "/distil/variable_summaries", map[string]string{
		"index":    "datasets",
		"dataset":  "o_185",
		"variable": "Number_seasons",
	}, nil)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, VariableSummaryHandler(storageCtor, ctor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"histogram": {
				"name":"Number_seasons",
				"type":"numerical",
				"extrema": {
					"min": 0,
					"max": 4
				},
				"buckets":[
					{"key":"0", "count":1},
					{"key":"1", "count":0},
					{"key":"2", "count":0}
				]
			}
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestVariableSummaryHandlerCategorical(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{
		"./testdata/variables.json",
		"./testdata/variable_summaries_categorical.json",
	})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// instantiate storage client constructor.
	storageCtor := elastic.NewStorage(ctor)

	// put together a stub dataset request
	req := mock.HTTPRequest(t, "GET", "/distil/variable_summaries", map[string]string{
		"index":    "datasets",
		"dataset":  "o_185",
		"variable": "Player",
	}, nil)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, VariableSummaryHandler(storageCtor, ctor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"histogram": {
				"name":"Player",
				"type":"categorical",
				"buckets":[
					{"key":"a","count":0},
					{"key":"b","count":0},
					{"key":"c","count":0}
				]
			}
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestVariableSummaryHandlerCategoricalPostgres(t *testing.T) {
	// mock postgres client
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mock.NewDatabaseDriver(ctrl)

	ctor := mockContructor(mockDB)

	// instantiate storage filter client constructor.
	storageCtor := postgres.NewStorage(ctor)

	// mock elasticsearch request handler
	handlerES := mock.ElasticHandler(t, []string{
		"./testdata/variables.json",
		"./testdata/variable_summaries_categorical.json",
	})
	// mock elasticsearch client
	ctorES := mock.ElasticClientCtor(t, handlerES)

	// put together a stub dataset request
	req := mock.HTTPRequest(t, "GET", "/distil/variable_summaries", map[string]string{
		"index":    "datasets",
		"dataset":  "o_185",
		"variable": "Player",
	}, nil)

	// setup the expected query
	mockDB.EXPECT().Query(
		"SELECT \"Player\", COUNT(*) AS count FROM o_185 GROUP BY \"Player\";").Return(nil, nil)

	// execute the test request
	res := mock.HTTPResponse(t, req, VariableSummaryHandler(storageCtor, ctorES))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"histogram": {
				"name":"Player",
				"type":"categorical",
				"buckets":[]
			}
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}
