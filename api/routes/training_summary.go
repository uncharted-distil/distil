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
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// TrainingSummaryHandler generates a route handler that facilitates the
// creation and retrieval of summary information about the specified variable
// for data returned in a result set.
func TrainingSummaryHandler(metaCtor api.MetadataStorageCtor, solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variable name
		variable := pat.Param(r, "variable")
		// get variable summary mode
		mode, err := api.SummaryModeFromString(pat.Param(r, "mode"))
		if err != nil {
			handleError(w, err)
			return
		}

		// get result id
		resultID, err := url.PathUnescape(pat.Param(r, "results-uuid"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape result id"))
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

		// get solution client
		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// get result URI
		result, err := solution.FetchSolutionResultByUUID(resultID)
		if err != nil {
			handleError(w, err)
			return
		}

		// get storage client
		data, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		ds, err := meta.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// get the dataset for predictions
		if ds == nil {
			pred, err := solution.FetchPredictionResultByProduceRequestID(result.ProduceRequestID)
			if err != nil {
				handleError(w, err)
				return
			}
			dataset = pred.Dataset
			ds, err = meta.FetchDataset(dataset, false, false, false)
			if err != nil {
				handleError(w, err)
				return
			}
		}
		storageName := ds.StorageName
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
		summary, err := data.FetchSummaryByResult(dataset, storageName, variable, result.ResultURI, filterParams, nil, api.SummaryMode(mode))
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, SummaryResult{
			Summary: summary,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result variable summary into JSON"))
			return
		}
	}
}
