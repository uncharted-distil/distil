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

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// TimeseriesForecastResult represents the result of a timeseries request.
type TimeseriesForecastResult struct {
	Timeseries [][]float64 `json:"timeseries"`
	Forecast   [][]float64 `json:"forecast"`
}

// TimeseriesForecastHandler returns timeseries data.
func TimeseriesForecastHandler(dataCtor api.DataStorageCtor, solutionCtor api.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		dataset := pat.Param(r, "dataset")
		timeseriesColName := pat.Param(r, "timeseriesColName")
		xColName := pat.Param(r, "xColName")
		yColName := pat.Param(r, "yColName")
		resultUUID := pat.Param(r, "result-uuid")
		timeseriesURI := pat.Param(r, "timeseriesURI")
		storageName := model.NormalizeDatasetID(dataset)

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

		// get storage client
		data, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch timeseries
		timeseries, err := data.FetchTimeseries(dataset, storageName, timeseriesColName, xColName, yColName, timeseriesURI, filterParams, false)
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

		forecast, err := data.FetchTimeseriesForecast(dataset, storageName, timeseriesColName, xColName, yColName, timeseriesURI, res.ResultURI, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		err = handleJSON(w, TimeseriesForecastResult{
			Timeseries: timeseries,
			Forecast:   forecast,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
