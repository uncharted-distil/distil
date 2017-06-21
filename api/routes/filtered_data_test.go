package routes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/unchartedsoftware/distil/api/filter"
	"github.com/unchartedsoftware/distil/api/util/mock"
)

func TestParseSearchParams(t *testing.T) {
	url := "/distil/filtered_data/o_185"
	params := "On_base_pct=integer,0,100&Position=categorical,Catcher,Pitcher&Triples&size=100"
	req := mock.HTTPRequest(t, "GET", url+"?"+params, nil)
	// parse filters
	set, err := parseFilterSet(req)
	assert.NoError(t, err)
	// check range filter
	assert.Contains(t, set.Filters, "On_base_pct")
	assert.IsType(t, set.Filters["On_base_pct"], &filter.Range{})
	assert.Equal(t, set.Filters["On_base_pct"].(*filter.Range).Min, 0.0)
	assert.Equal(t, set.Filters["On_base_pct"].(*filter.Range).Max, 100.0)
	// check category filter
	assert.Contains(t, set.Filters, "Position")
	assert.IsType(t, set.Filters["Position"], &filter.Category{})
	assert.Equal(t, set.Filters["Position"].(*filter.Category).Categories, []string{"Catcher", "Pitcher"})
	// check empty filter
	assert.Contains(t, set.Filters, "Triples")
	assert.IsType(t, set.Filters["Triples"], &filter.Empty{})
	// check size filter
	assert.Equal(t, set.Size, 100)
}

func TestParseSearchParamsMalformed(t *testing.T) {
	url := "/distil/filtered_data/o_185"
	missingSizeParams := "On_base_pct=integer,0,10&size"
	missingNumericTypeParams := "On_base_pct=0,10"
	missingNumericRangeParams := "On_base_pct=integer,0"
	missingCategoricalTypeParams := "Position=Catcher"
	missingCategoricalCategoriesParams := "Position=categorical"

	// parse filters
	req := mock.HTTPRequest(t, "GET", url+"?"+missingSizeParams, nil)
	_, err := parseFilterSet(req)
	// check for error
	assert.Error(t, err)

	// parse filters
	req = mock.HTTPRequest(t, "GET", url+"?"+missingNumericTypeParams, nil)
	_, err = parseFilterSet(req)
	// check for error
	assert.Error(t, err)

	// parse filters
	req = mock.HTTPRequest(t, "GET", url+"?"+missingNumericRangeParams, nil)
	_, err = parseFilterSet(req)
	// check for error
	assert.Error(t, err)

	// parse filters
	req = mock.HTTPRequest(t, "GET", url+"?"+missingCategoricalTypeParams, nil)
	_, err = parseFilterSet(req)
	// check for error
	assert.Error(t, err)

	// parse filters
	req = mock.HTTPRequest(t, "GET", url+"?"+missingCategoricalCategoriesParams, nil)
	_, err = parseFilterSet(req)
	// check for error
	assert.Error(t, err)
}

func TestFilteredDataHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{"./testdata/filtered_data.json"})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// mock http request
	url := "/distil/filtered_data/o_185"
	params := "On_base_pct=float,0,100&Position=categorical,Catcher&Triples"
	req := mock.HTTPRequest(t, "GET", url+"?"+params, map[string]string{
		"dataset": "o_185",
	})

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, FilteredDataHandler(ctor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected := `{
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
		}`
	actual := string(res.Body.Bytes())
	assert.JSONEq(t, expected, actual)
}
