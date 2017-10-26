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

func TestVariableSummaryHandlerCategorical(t *testing.T) {
	// mock elasticsearch request handler
	handlerES := mock.ElasticHandler(t, []string{
		"./testdata/variables.json",
	})
	// mock elasticsearch client & storage
	ctorES := mock.ElasticClientCtor(t, handlerES)
	ctorESStorage := elastic.NewMetadataStorage(ctorES)

	// mock postgres client
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mock.NewDatabaseDriver(ctrl)

	ctor := mockContructor(mockDB)

	// instantiate storage filter client constructor.
	storageCtor := postgres.NewDataStorage(ctor, ctorESStorage)

	// put together a stub dataset request
	req := mock.HTTPRequest(t, "GET", "/distil/variable_summaries", map[string]string{
		"index":    "datasets",
		"dataset":  "o_185",
		"variable": "Position",
	}, nil)

	// setup the expected query
	mockDB.EXPECT().Query(
		"SELECT \"Position\", COUNT(*) AS count FROM o_185 GROUP BY \"Position\" ORDER BY count desc, \"Position\" LIMIT 10;").Return(nil, nil)

	// execute the test request
	res := mock.HTTPResponse(t, req, VariableSummaryHandler(storageCtor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"histogram": {
				"name":"Position",
				"type":"categorical",
				"buckets":[]
			}
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}
