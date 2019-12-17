//
//   Copyright © 2019 Uncharted Software Inc.
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

// PredictedSummary contains a fetch result histogram.
type PredictedSummary struct {
	PredictedSummary *api.VariableSummary `json:"summary"`
}

func fetchSolutionPredictedExtrema(meta api.MetadataStorage, data api.DataStorage, solution api.SolutionStorage, dataset string, storageName string, target string, solutionID string) (*api.Extrema, error) {
	// check target var type
	variable, err := meta.FetchVariable(dataset, target)
	if err != nil {
		return nil, err
	}

	if !model.IsNumerical(variable.Type) && variable.Type != model.TimeSeriesType {
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
			if sol.Results != nil && !sol.IsBad {
				// result uri
				resultURI := sol.Results[0].ResultURI
				// predicted extrema
				predictedExtrema, err := data.FetchResultsExtremaByURI(dataset, storageName, resultURI)
				if err != nil {
					return nil, err
				}
				max = math.Max(max, predictedExtrema.Max)
				min = math.Min(min, predictedExtrema.Min)
				// result extrema
				resultExtrema, err := data.FetchExtremaByURI(dataset, storageName, resultURI, target)
				if err != nil {
					return nil, err
				}
				max = math.Max(max, resultExtrema.Max)
				min = math.Min(min, resultExtrema.Min)
			}
		}
	}
	return api.NewExtrema(min, max)
}

// PredictedSummaryHandler bins predicted result data for consumption in a downstream summary view.
func PredictedSummaryHandler(metaCtor api.MetadataStorageCtor, solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		dataset := pat.Param(r, "dataset")
		target := pat.Param(r, "target")
		storageName := model.NormalizeDatasetID(dataset)

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

		// extract extrema for solution
		extrema, err := fetchSolutionPredictedExtrema(meta, data, solution, dataset, storageName, target, "")
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
		summary, err := data.FetchPredictedSummary(dataset, storageName, res.ResultURI, filterParams, extrema)
		if err != nil {
			handleError(w, err)
			return
		}
		summary.Key = api.GetPredictedKey(res.SolutionID)
		summary.Label = "Predicted"
		summary.SolutionID = res.SolutionID

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
