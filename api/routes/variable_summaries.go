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
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variabloe name
		variable := pat.Param(r, "variable")

		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		// get variable names and ranges out of the params
		filterParams, err := model.ParseFilterParamsFromJSON(params)
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
		histogram, err := storage.FetchSummary(dataset, variable, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, SummaryResult{
			Histogram: histogram,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
	}
}
