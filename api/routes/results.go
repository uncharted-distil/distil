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
func ResultsHandler(storageCtor model.PipelineStorageCtor, storageDataCtor model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		index := pat.Param(r, "index")
		dataset := pat.Param(r, "dataset")
		inclusive := pat.Param(r, "inclusive")
		inclusiveBool := inclusive == "inclusive"

		pipelineID, err := url.PathUnescape(pat.Param(r, "pipeline-id"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape result uuid"))
			return
		}

		// get variable names and ranges out of the params
		filterParams, err := model.ParseFilterParamsURL(r.URL.Query())
		if err != nil {
			handleError(w, err)
			return
		}

		client, err := storageCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		clientData, err := storageDataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// get the result URI
		res, err := client.FetchResultMetadataByPipelineID(pipelineID)
		if err != nil {
			handleError(w, err)
			return
		}

		if res == nil {
			handleError(w, errors.Errorf("pipeline id `%s` cannot be mapped to result URI", pipelineID))
			return
		}

		results, err := clientData.FetchFilteredResults(dataset, index, res.ResultURI, filterParams, inclusiveBool)
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
