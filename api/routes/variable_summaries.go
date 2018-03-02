package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// SummaryResult represents a summary response for a variable.
type SummaryResult struct {
	Histogram *model.Histogram `json:"histogram"`
}

// VariableSummaryHandler generates a route handler that facilitates the
// creation and retrieval of summary information about the specified variable.
func VariableSummaryHandler(ctorStorage model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variabloe name
		variable := pat.Param(r, "variable")

		// get variable names and ranges out of the params
		filterParams, err := model.ParseFilterParamsURL(r.URL.Query())
		if err != nil {
			handleError(w, err)
			return
		}

		// get storage client
		storage, err := ctorStorage()
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch summary histogram
		histogram, err := storage.FetchSummary(dataset, index, variable, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall output into JSON
		err = handleJSON(w, SummaryResult{
			Histogram: histogram,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
	}
}
