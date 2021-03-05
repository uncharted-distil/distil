//
//   Copyright © 2021 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package routes

import (
	"math"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

func fetchPredictionResultExtrema(meta api.MetadataStorage, data api.DataStorage, solution api.SolutionStorage,
	dataset string, storageName string, target string, fittedSolutionID string, resultURI string) (*api.Extrema, error) {
	// check target var type
	variable, err := meta.FetchVariable(dataset, target)
	if err != nil {
		return nil, err
	}

	if !model.IsNumerical(variable.Type) && variable.Type != model.TimeSeriesType {
		return nil, nil
	}

	min := math.MaxFloat64
	max := -math.MaxFloat64
	// predicted extrema
	predictedExtrema, err := data.FetchResultsExtremaByURI(dataset, storageName, resultURI)
	if err != nil {
		return nil, err
	}
	max = math.Max(max, predictedExtrema.Max)
	min = math.Min(min, predictedExtrema.Min)

	return api.NewExtrema(min, max)
}

// PredictionResultSummaryHandler bins predicted result data for consumption in a downstream summary view.
func PredictionResultSummaryHandler(metaCtor api.MetadataStorageCtor, solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get variable summary mode
		mode, err := api.SummaryModeFromString(pat.Param(r, "mode"))
		if err != nil {
			handleError(w, err)
			return
		}

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

		// get the prediction result metadata.
		res, err := solution.FetchPredictionResultByUUID(resultUUID)
		if err != nil {
			handleError(w, err)
			return
		}

		// get the associated prediction request.
		prediction, err := solution.FetchPrediction(res.ProduceRequestID)
		if err != nil {
			handleError(w, err)
			return
		}

		ds, err := meta.FetchDataset(prediction.Dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName

		// extract extrema for solution
		extrema, err := fetchPredictionResultExtrema(meta, data, solution, prediction.Dataset, storageName, prediction.Target, prediction.FittedSolutionID, res.ResultURI)
		if err != nil {
			handleError(w, err)
			return
		}
		// if the variable is a geobounds and there is a band column, add a filter
		// to only consider the first band.
		hasBand := false
		isGeobounds := false
		for _, v := range ds.Variables {
			if v.DisplayName == "band" {
				hasBand = true
			} else if model.IsGeoBounds(v.Type) {
				isGeobounds = true
			}
		}
		if hasBand && isGeobounds {
			boundsFilter := model.NewCategoricalFilter("band", model.IncludeFilter, []string{"01"})
			boundsFilter.IsBaselineFilter = true
			filterParams.Filters = append(filterParams.Filters, boundsFilter)
		}
		// fetch summary histogram
		summary, err := data.FetchPredictedSummary(prediction.Dataset, storageName, res.ResultURI, filterParams, extrema, api.SummaryMode(mode))
		if err != nil {
			handleError(w, err)
			return
		}
		summary.Key = api.GetPredictedKey(res.ProduceRequestID)

		// marshal data and sent the response back
		err = handleJSON(w, PredictedSummary{
			PredictedSummary: summary,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
