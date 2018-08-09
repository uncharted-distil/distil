package routes

import (
	"math"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// PredictedSummary contains a fetch result histogram.
type PredictedSummary struct {
	PredictedSummary *model.Histogram `json:"histogram"`
}

func fetchSolutionPredictedExtrema(meta model.MetadataStorage, data model.DataStorage, solution model.SolutionStorage, dataset string, target string, solutionID string) (*model.Extrema, error) {
	// check target var type
	variable, err := meta.FetchVariable(dataset, target)
	if err != nil {
		return nil, err
	}

	if !model.IsNumerical(variable.Type) {
		return nil, nil
	}

	// we need to get extrema min and max for all solutions sharing dataset and target
	requests, err := solution.FetchRequestByDatasetTarget(dataset, target, solutionID)
	if err != nil {
		return nil, err
	}

	// get extrema
	min := math.MaxFloat64
	max := -math.MaxFloat64
	for _, req := range requests {
		for _, sol := range req.Solutions {
			if sol.Result != nil && !sol.IsBad {
				// result uri
				resultURI := sol.Result.ResultURI
				// predicted extrema
				predictedExtrema, err := data.FetchResultsExtremaByURI(dataset, resultURI)
				if err != nil {
					return nil, err
				}
				max = math.Max(max, predictedExtrema.Max)
				min = math.Min(min, predictedExtrema.Min)
				// result extrema
				resultExtrema, err := data.FetchExtremaByURI(dataset, resultURI, target)
				if err != nil {
					return nil, err
				}
				max = math.Max(max, resultExtrema.Max)
				min = math.Min(min, resultExtrema.Min)
			}
		}
	}
	return model.NewExtrema(min, max)
}

// PredictedSummaryHandler bins predicted result data for consumption in a downstream summary view.
func PredictedSummaryHandler(metaCtor model.MetadataStorageCtor, solutionCtor model.SolutionStorageCtor, dataCtor model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		dataset := pat.Param(r, "dataset")
		target := pat.Param(r, "target")

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

		// extract extrema for solution
		extrema, err := fetchSolutionPredictedExtrema(meta, data, solution, dataset, target, "")
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
		histogram, err := data.FetchPredictedSummary(dataset, res.ResultURI, filterParams, extrema)
		if err != nil {
			handleError(w, err)
			return
		}
		histogram.Key = model.GetPredictedKey(histogram.Key, res.SolutionID)
		histogram.Label = "Predicted"
		histogram.SolutionID = res.SolutionID

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
