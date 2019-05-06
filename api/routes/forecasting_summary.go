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
	"strconv"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// ForecastingSummary contains a fetch result histogram.
type ForecastingSummary struct {
	Histogram *api.Histogram `json:"histogram"`
}

// ForecastingSummaryHandler bins forecasted result data for consumption in a downstream summary view.
func ForecastingSummaryHandler(solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		dataset := pat.Param(r, "dataset")
		xColName := pat.Param(r, "xColName")
		yColName := pat.Param(r, "yColName")
		binningInterval := pat.Param(r, "binningInterval")
		storageName := model.NormalizeDatasetID(dataset)

		interval, err := strconv.ParseInt(binningInterval, 10, 64)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		resultID, err := url.PathUnescape(pat.Param(r, "results-uuid"))
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
		res, err := solution.FetchSolutionResultByUUID(resultID)
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch forecasted histogram
		histogram, err := data.FetchForecastingSummary(dataset, storageName, xColName, yColName, int(interval), res.ResultURI, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}
		histogram.Label = "Predicted"
		histogram.Key = api.GetPredictedKey(histogram.Key, res.SolutionID)
		histogram.SolutionID = res.SolutionID

		// marshal data and sent the response back
		err = handleJSON(w, ForecastingSummary{
			Histogram: histogram,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
