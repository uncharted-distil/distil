package routes

import (
	"net/http"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/model/filter"
	"github.com/unchartedsoftware/distil/api/util/json"
	"github.com/unchartedsoftware/distil/api/util/mock"
)

func TestParseSearchParams(t *testing.T) {
	params := map[string]string{
		"dataset": "o_185",
	}
	query := map[string]string{
		"On_base_pct": "numerical,0,100",
		"Position":    "categorical,Catcher,Pitcher",
		"Triples":     "",
		"size":        "100",
	}
	req := mock.HTTPRequest(t, "GET", "/distil/data", params, query)
	// parse
	filters, err := parseFilterParams(req)
	assert.NoError(t, err)
	assert.Equal(t, filters.Ranged[0].Name, "On_base_pct")
	assert.Equal(t, filters.Ranged[0].Min, 0.0)
	assert.Equal(t, filters.Ranged[0].Max, 100.0)
	assert.Equal(t, filters.Categorical[0].Name, "Position")
	assert.Equal(t, filters.Categorical[0].Categories, []string{"Catcher", "Pitcher"})
	assert.Equal(t, filters.None[0], "Triples")
	assert.Equal(t, filters.Size, 100)
}

func TestParseSearchParamsMalformed(t *testing.T) {
	params := map[string]string{
		"dataset": "o_185",
	}
	query0 := map[string]string{
		"On_base_pct": "numerical,0",
	}
	query1 := map[string]string{
		"Position": "categorical",
	}
	query2 := map[string]string{
		"Triples": "numerical,1,2,3",
	}
	query3 := map[string]string{
		"size": "",
	}
	// missing max
	req := mock.HTTPRequest(t, "GET", "/distil/data", params, query0)
	filters, err := parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// missing categories
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query1)
	filters, err = parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// extra param
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query2)
	filters, err = parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// missing size
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query3)
	filters, err = parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
}

func TestFilteredDataHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{"./testdata/filtered_data.json"})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// instantiate storage filter client constructor.
	storageCtor := filter.NewElasticFilter(ctor)

	// put together a stub dataset request
	params := map[string]string{
		"dataset": "o_185",
	}
	query := map[string]string{
		"On_base_pct": "numerical,0,100",
		"Position":    "categorical,Catcher",
		"Triples":     "",
	}
	req := mock.HTTPRequest(t, "GET", "/distil/data", params, query)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, FilteredDataHandler(storageCtor))
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

func TestFilteredPostgresHandler(t *testing.T) {
	// mock postgres client
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mock.NewMockDatabaseDriver(ctrl)

	ctor := mockContructor(mockDB)

	// instantiate storage filter client constructor.
	storageCtor := filter.NewPostgresFilter(ctor)

	// put together a stub dataset request
	params := map[string]string{
		"dataset": "o_185",
	}
	query := map[string]string{
		"On_base_pct": "numerical,0,100",
		"Position":    "categorical,Catcher",
		"Triples":     "",
	}

	// Identify the expected behaviour.
	// NOTE: It currently expects an empty set since pgx.Rows is hardly accessible.
	mockDB.EXPECT().Query("SELECT * FROM o_185 WHERE On_base_pct.value >= $1 AND On_base_pct.value <= $2 AND Position.value IN ($3);", float64(0), float64(100), "Catcher").Return(nil, nil)
	req := mock.HTTPRequest(t, "GET", "/distil/data", params, query)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, FilteredDataHandler(storageCtor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
				"name": "o_185",
				"metadata": [],
				"values": []
			}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func mockContructor(mockDB *mock.MockDatabaseDriver) filter.ClientCtor {
	return func() (filter.DatabaseDriver, error) {
		return mockDB, nil
	}
}
