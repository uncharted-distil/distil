package routes

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// ResultsSummary contains a fetch result histogram.
type ResultsSummary struct {
	ResultsSummary *model.Histogram `json:"histogram"`
}

// ResultsSummaryHandler bins predicted result data for consumption in a downstream summary view.
func ResultsSummaryHandler(pipelineCtor model.PipelineStorageCtor, dataCtor model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		index := pat.Param(r, "index")
		dataset := pat.Param(r, "dataset")

		resultUUID, err := url.PathUnescape(pat.Param(r, "results-uuid"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape results uuid"))
			return
		}
		extremaMin, err := strconv.ParseFloat(pat.Param(r, "min"), 64)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to parse extrema min"))
			return
		}
		extremaMax, err := strconv.ParseFloat(pat.Param(r, "max"), 64)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to parse extrema max"))
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

		// get the result URI. Error ignored to make it ES compatible.
		res, err := pipeline.FetchPipelineResultByUUID(resultUUID)
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch summary histogram
		histogram, err := data.FetchResultsSummary(dataset, res.ResultURI, index, filterParams, &model.Extrema{
			Min: extremaMin,
			Max: extremaMax,
		})
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall data and sent the response back
		err = handleJSON(w, ResultsSummary{
			ResultsSummary: histogram,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
