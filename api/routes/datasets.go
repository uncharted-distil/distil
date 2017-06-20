package routes

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"goji.io/pat"
	"gopkg.in/olivere/elastic.v3"
)

// DatasetResult represents the result of a datasets response.
type DatasetResult struct {
	Datasets []model.Dataset `json:"datasets"`
}

// DatasetsHandler generates a route handler that facilitates a search of
// dataset descriptions and variable names, returning a name, description and
// variable list for any dataset that matches. The search parameter is optional
// it contains the search terms if set, and if unset, flags that a list of all
// datasets should be returned.  The full list will be contain names only,
// descriptions and variable lists will not be included.
func DatasetsHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")
		// check for search terms
		terms, err := url.QueryUnescape(r.URL.Query().Get("search"))
		if err != nil {
			handleError(w, errors.Wrap(err, "Malformed datasets query"))
			return
		}
		// if its present, forward a search, otherwise fetch all datasets
		var datasets []model.Dataset
		if terms != "" {
			datasets, err = model.SearchDatasets(client, index, terms)
		} else {
			datasets, err = model.FetchDatasets(client, index)
		}
		if err != nil {
			handleError(w, err)
			return
		}
		// marshall data
		err = handleJSON(w, DatasetResult{
			Datasets: datasets,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
