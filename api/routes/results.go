package routes

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// Results represents a results response for a variable.
type Results struct {
	Results *api.FilteredData `json:"results"`
}

// ResultsHandler fetches predicted solution values and returns them to the client
// in a JSON structure
func ResultsHandler(solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

		dataset := pat.Param(r, "dataset")
		storageName := model.NormalizeDatasetID(dataset)

		solutionID, err := url.PathUnescape(pat.Param(r, "solution-id"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape solution id"))
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

		// get the filters
		req, err := solution.FetchRequestBySolutionID(solutionID)
		if err != nil {
			handleError(w, err)
			return
		}
		if req == nil {
			handleError(w, errors.Errorf("solution id `%s` cannot be mapped to result URI", solutionID))
			return
		}

		// merge provided filterParams with those of the request
		filterParams.Merge(req.Filters)

		// get the result URI
		res, err := solution.FetchSolutionResult(solutionID)
		if err != nil {
			handleError(w, err)
			return
		}

		// if no result, return an empty map
		if res == nil {
			handleJSON(w, make(map[string]interface{}))
			return
		}

		results, err := data.FetchResults(dataset, storageName, res.ResultURI, solutionID, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data and sent the response back
		err = handleJSON(w, results)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal solution result into JSON"))
			return
		}

		return
	}
}
