package routes

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// PredictedSummary contains a fetch result histogram.
type PredictedSummary struct {
	PredictedSummary *model.Histogram `json:"histogram"`
}

// PredictedSummaryHandler bins predicted result data for consumption in a downstream summary view.
func PredictedSummaryHandler(solutionCtor model.SolutionStorageCtor, dataCtor model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		dataset := pat.Param(r, "dataset")

		resultUUID, err := url.PathUnescape(pat.Param(r, "results-uuid"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape results uuid"))
			return
		}

		minStr := pat.Param(r, "min")
		maxStr := pat.Param(r, "max")
		var extrema *model.Extrema
		if minStr != "null" && maxStr != "null" {
			extremaMin, err := strconv.ParseFloat(minStr, 64)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to parse extrema min"))
				return
			}
			extremaMax, err := strconv.ParseFloat(maxStr, 64)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to parse extrema max"))
				return
			}
			extrema, err = model.NewExtrema(extremaMin, extremaMax)
			if err != nil {
				handleError(w, err)
				return
			}
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
		histogram, err := data.FetchResultsSummary(dataset, res.ResultURI, filterParams, extrema)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall data and sent the response back
		err = handleJSON(w, PredictedSummary{
			PredictedSummary: histogram,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
