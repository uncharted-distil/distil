package routes

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil-compute/model"
	api "github.com/unchartedsoftware/distil/api/model"
)

// TargetSummaryHandler generates a route handler that facilitates the
// creation and retrieval of summary information about the specified variable
// for data returned in a result set.
func TargetSummaryHandler(metaCtor api.MetadataStorageCtor, solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		storageName := model.NormalizeDatasetID(dataset)
		// get variable name
		target := pat.Param(r, "target")

		// get result id
		resultID, err := url.PathUnescape(pat.Param(r, "results-uuid"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape result id"))
			return
		}

		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		// get variable names and ranges out of the params
		filterParams, err := api.ParseFilterParamsFromJSON(params)
		if err != nil {
			handleError(w, err)
			return
		}

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		data, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// get result URI
		result, err := solution.FetchSolutionResultByUUID(resultID)
		if err != nil {
			handleError(w, err)
			return
		}

		// extract extrema for solution
		extrema, err := fetchSolutionPredictedExtrema(meta, data, solution, dataset, storageName, target, "")
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch summary histogram
		histogram, err := data.FetchSummaryByResult(dataset, storageName, target, result.ResultURI, filterParams, extrema)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, SummaryResult{
			Histogram: histogram,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result variable summary into JSON"))
			return
		}
	}
}
