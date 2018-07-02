package routes

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// CorrectnessSummary contains a fetch result histogram.
type CorrectnessSummary struct {
	CorrectnessSummary *model.Histogram `json:"histogram"`
}

// CorrectnessSummaryHandler bins predicted result data for consumption in a downstream summary view.
func CorrectnessSummaryHandler(solutionCtor model.SolutionStorageCtor, dataCtor model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		dataset := pat.Param(r, "dataset")

		resultUUID, err := url.PathUnescape(pat.Param(r, "results-uuid"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape results uuid"))
			return
		}

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

		// get the result URI. Error ignored to make it ES compatible.
		res, err := solution.FetchSolutionResultByUUID(resultUUID)
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch summary histogram
		histogram, err := data.FetchCorrectnessSummary(dataset, res.ResultURI, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}
		histogram.Key = model.GetErrorKey(histogram.Key, res.SolutionID)
		histogram.Label = "Error"
		histogram.SolutionID = res.SolutionID

		// marshall data and sent the response back
		err = handleJSON(w, CorrectnessSummary{
			CorrectnessSummary: histogram,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
