package routes

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// ResultVariableSummaryHandler generates a route handler that facilitates the
// creation and retrieval of summary information about the specified variable
// for data returned in a result set.
func ResultVariableSummaryHandler(ctorPipeline model.PipelineStorageCtor, ctorStorage model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variable name
		variable := pat.Param(r, "variable")
		// get result id
		resultID, err := url.PathUnescape(pat.Param(r, "results-uuid"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape result id"))
			return
		}
		min := pat.Param(r, "min")
		max := pat.Param(r, "max")
		var extrema *model.Extrema
		if min != "null" && max != "null" {
			extremaMin, err := strconv.ParseFloat(min, 64)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to parse extrema min"))
				return
			}
			extremaMax, err := strconv.ParseFloat(max, 64)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to parse extrema max"))
				return
			}
			extrema = &model.Extrema{
				Min: extremaMin,
				Max: extremaMax,
			}
		}

		// get pipeline client
		pipelineData, err := ctorPipeline()
		if err != nil {
			handleError(w, err)
			return
		}

		// get result URI
		result, err := pipelineData.FetchResultMetadataByUUID(resultID)
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
		histogram, err := storage.FetchSummaryByResult(dataset, index, variable, result.ResultURI, extrema)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall output into JSON
		err = handleJSON(w, SummaryResult{
			Histogram: histogram,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result variable summary into JSON"))
			return
		}
	}
}
