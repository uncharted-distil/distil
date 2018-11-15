package routes

import (
	"net/http"
	"net/url"

	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"github.com/russross/blackfriday"

	"github.com/unchartedsoftware/distil/api/model"
)

// DatasetResult represents the result of a datasets response.
type DatasetResult struct {
	Dataset []*model.Dataset `json:"datasets"`
}

// DatasetsHandler generates a route handler that facilitates a search of
// dataset descriptions and variable names, returning a name, description and
// variable list for any dataset that matches. The search parameter is optional
// it contains the search terms if set, and if unset, flags that a list of all
// datasets should be returned.  The full list will be contain names only,
// descriptions and variable lists will not be included.
func DatasetsHandler(metaCtor model.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for search terms
		terms, err := url.QueryUnescape(r.URL.Query().Get("search"))
		if err != nil {
			handleError(w, errors.Wrap(err, "Malformed datasets query"))
			return
		}
		// get elasticsearch client
		storage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		// if its present, forward a search, otherwise fetch all datasets
		var datasets []*model.Dataset
		if terms != "" {
			datasets, err = storage.SearchDatasets(terms, false, false)
		} else {
			datasets, err = storage.FetchDatasets(false, false)
		}
		if err != nil {
			handleError(w, err)
			return
		}
		// render dataset description as HTML
		for _, dataset := range datasets {
			dataset.Description = renderMarkdown(dataset.Description)
		}
		// marshall data
		err = handleJSON(w, DatasetResult{
			Dataset: datasets,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}

func renderMarkdown(markdown string) string {
	// process the markdown into HTML
	unsafe := blackfriday.Run([]byte(markdown))
	// just to be safe, sanatize the HTML
	return string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
}
