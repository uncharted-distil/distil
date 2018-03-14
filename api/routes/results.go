package routes

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// Results represents a results response for a variable.
type Results struct {
	Results *model.FilteredData `json:"results"`
}

// ResultsHandler fetches predicted pipeline values and returns them to the client
// in a JSON structure
func ResultsHandler(pipelineCtor model.PipelineStorageCtor, dataCtor model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

		dataset := pat.Param(r, "dataset")
		esIndex := pat.Param(r, "index")

		pipelineID, err := url.PathUnescape(pat.Param(r, "pipeline-id"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape pipeline id"))
			return
		}

		pipeline, err := pipelineCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		data, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// get the result URI
		res, err := pipeline.FetchResultMetadataByPipelineID(pipelineID)
		if err != nil {
			handleError(w, err)
			return
		}

		if res == nil {
			handleError(w, errors.Errorf("pipeline id `%s` cannot be mapped to result URI", pipelineID))
			return
		}

		results, err := data.FetchFilteredResults(dataset, esIndex, res.ResultURI, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall data and sent the response back
		err = handleJSON(w, results)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal pipeline result into JSON"))
			return
		}

		return
	}
}
