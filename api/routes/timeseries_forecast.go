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

	"github.com/uncharted-distil/distil/api/compute"
	api "github.com/uncharted-distil/distil/api/model"
)

// TimeseriesForecastResult represents the result of a timeseries request.
type TimeseriesForecastResult struct {
	Timeseries        []*api.TimeseriesObservation `json:"timeseries"`
	Forecast          []*api.TimeseriesObservation `json:"forecast"`
	ForecastTestRange []float64                    `json:"forecastTestRange"`
	IsDateTime        bool                         `json:"isDateTime"`
	Min               api.NullableFloat64          `json:"min"`
	Max               api.NullableFloat64          `json:"max"`
	Mean              api.NullableFloat64          `json:"mean"`
}

// TimeseriesForecastHandler returns timeseries data.
func TimeseriesForecastHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, solutionCtor api.SolutionStorageCtor, trainTestSplitTimeSeries float64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		truthDataset := pat.Param(r, "truthDataset")
		forecastDataset := pat.Param(r, "forecastDataset")
		timeseriesColName := pat.Param(r, "timeseriesColName")
		xColName := pat.Param(r, "xColName")
		yColName := pat.Param(r, "yColName")
		resultUUID := pat.Param(r, "result-uuid")
		timeseriesURI := pat.Param(r, "timeseriesURI")

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

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		dst, err := meta.FetchDataset(truthDataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		truthStorageName := dst.StorageName

		dsf, err := meta.FetchDataset(forecastDataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		predictedStorageName := dsf.StorageName

		// fetch the ground truth timeseries
		timeseries, err := data.FetchTimeseries(truthDataset, truthStorageName, timeseriesColName, xColName, yColName, timeseriesURI, filterParams, false)
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

		// fetch the predicted timeseries
		forecast, err := data.FetchTimeseriesForecast(forecastDataset, predictedStorageName, timeseriesColName, xColName, yColName, timeseriesURI, res.ResultURI, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		// Recompute train/test split info for visualization purposes
		split := compute.SplitTimeSeries(timeseries.Timeseries, trainTestSplitTimeSeries)

		err = handleJSON(w, TimeseriesForecastResult{
			Timeseries:        timeseries.Timeseries,
			Forecast:          forecast.Timeseries,
			ForecastTestRange: []float64{split.SplitValue, split.EndValue},
			IsDateTime:        timeseries.IsDateTime,
			Min:               forecast.Min,
			Max:               forecast.Max,
			Mean:              forecast.Mean,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
