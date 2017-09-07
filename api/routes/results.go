package routes

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
)

// Results represents a results response for a variable.
type Results struct {
	Results *model.FilteredData `json:"results"`
}

// ResultsHandler fetches predicted pipeline values and returns them to the client
// in a JSON structure
func ResultsHandler(esCtor elastic.ClientCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		index := pat.Param(r, "index")
		dataset := pat.Param(r, "dataset")
		pipelineURI, err := url.PathUnescape(pat.Param(r, "result-uri"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape result uri"))
			return
		}

		esClient, err := esCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		results, err := model.FetchResults(esClient, pipelineURI, index, dataset)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall data and sent the response back
		err = handleJSON(w, Results{
			Results: results,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal pipeline result into JSON"))
			return
		}

		return
	}
}
