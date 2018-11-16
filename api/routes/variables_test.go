package routes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/unchartedsoftware/distil/api/model/storage/elastic"
	"github.com/unchartedsoftware/distil/api/util/json"
	"github.com/unchartedsoftware/distil/api/util/mock"
)

func TestVariableHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{
		"./testdata/variables.json",
	})
	// mock elasticsearch client & storage
	ctorES := mock.ElasticClientCtor(t, handler)
	ctorESStorage := elastic.NewMetadataStorage("datasets", ctorES)

	// put together a stub dataset request
	req := mock.HTTPRequest(t, "GET", "/distil/variables", map[string]string{
		"index":   "datasets",
		"dataset": "o_185",
	}, nil)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, VariablesHandler(ctorESStorage))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"variables": [
				{"colName":"d3mIndex","colType":"integer","importance": 0,"deleted": false,"selectedRole": "index","suggestedTypes": [{ "type": "integer", "probability": 1.00, "provenance": "TEST" }], "colOriginalVariable": "","colIndex": 0, "colOriginalType":"integer", "role": ["TEST"], "colDisplayName": "d3mIndex"},
				{"colName":"Position","colType":"categorical","importance": 0,"deleted": false,"selectedRole": "attribute","suggestedTypes": [{ "type": "categorical", "probability": 1.00, "provenance": "TEST" }], "colOriginalVariable": "","colIndex": 1, "colOriginalType":"categorical", "role": ["TEST"], "colDisplayName": "Position"},
				{"colName":"Number_seasons","colType":"integer","importance": 1,"deleted": false,"selectedRole": "attribute","suggestedTypes": [ { "type": "integer", "probability": 1.00, "provenance": "TEST" }], "colOriginalVariable": "","colIndex": 2, "colOriginalType":"integer", "role": ["TEST"], "colDisplayName": "Number_seasons"},
				{"colName":"Games_played","colType":"integer","importance": 2,"deleted": false,"selectedRole": "attribute","suggestedTypes": [ { "type": "integer", "probability": 1.00, "provenance": "TEST" }], "colOriginalVariable": "","colIndex": 3, "colOriginalType":"integer", "role": ["TEST"], "colDisplayName": "Games_played"},
				{"colName":"On_base_pct","colType":"float","importance": 3,"deleted": false,"selectedRole": "attribute","suggestedTypes": [ { "type": "float", "probability": 1.00, "provenance": "TEST" }], "colOriginalVariable": "","colIndex": 4, "colOriginalType":"float", "role": ["TEST"], "colDisplayName": "On_base_pct"}
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}
