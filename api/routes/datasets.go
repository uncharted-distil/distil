package routes

import (
	"net/http"
	"net/url"

	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"github.com/russross/blackfriday"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// DatasetResult represents the result of a dataset response.
type DatasetResult struct {
	Dataset *model.Dataset `json:"dataset"`
}

// DatasetsResult represents the result of a datasets response.
type DatasetsResult struct {
	Datasets []*model.Dataset `json:"datasets"`
}

// DatasetHandler generates a route handler that returns a specified dataset
// summary.
func DatasetHandler(ctor model.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// get dataset name
		dataset := pat.Param(r, "dataset")

		// get metadata client
		storage, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}

		// get dataset summary
		res, err := storage.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data
		err = handleJSON(w, DatasetResult{
			Dataset: res,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}

// DatasetsHandler generates a route handler that facilitates a search of
// dataset descriptions and variable names, returning a name, description and
// variable list for any dataset that matches. The search parameter is optional
// it contains the search terms if set, and if unset, flags that a list of all
// datasets should be returned.  The full list will be contain names only,
// descriptions and variable lists will not be included.
func DatasetsHandler(metaCtors []model.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var datasets []*model.Dataset
		// check for search terms
		terms, err := url.QueryUnescape(r.URL.Query().Get("search"))
		if err != nil {
			handleError(w, errors.Wrap(err, "Malformed datasets query"))
			return
		}
		for _, ctor := range metaCtors {
			// get metadata client
			storage, err := ctor()
			if err != nil {
				handleError(w, err)
				return
			}
			// if its present, forward a search, otherwise fetch all datasets
			var datasetsPart []*model.Dataset
			if terms != "" {
				datasetsPart, err = storage.SearchDatasets(terms, false, false)
			} else {
				datasetsPart, err = storage.FetchDatasets(false, false)
			}
			if err != nil {
				//handleError(w, err)
				continue
			}
			// render dataset description as HTML
			for _, dataset := range datasetsPart {
				dataset.Description = renderMarkdown(dataset.Description)
			}

			datasets = append(datasets, datasetsPart...)
		}

		// marshal data
		err = handleJSON(w, DatasetsResult{
			Datasets: datasets,
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
