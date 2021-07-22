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
	"net/http"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/compute"
	"github.com/uncharted-distil/distil/api/model"
	api "github.com/uncharted-distil/distil/api/model"
	"goji.io/v3/pat"
)

// TimeseriesForecastResult represents the result of a timeseries request.
type TimeseriesForecastResult struct {
	VarKey            string                       `json:"variableKey"`
	SeriesID          string                       `json:"seriesID"`
	Timeseries        []*api.TimeseriesObservation `json:"timeseries"`
	Forecast          []*api.TimeseriesObservation `json:"forecast"`
	ForecastTestRange []float64                    `json:"forecastTestRange"`
	IsDateTime        bool                         `json:"isDateTime"`
	Min               float64                      `json:"min"`
	Max               float64                      `json:"max"`
	Mean              float64                      `json:"mean"`
}

// TimeseriesForecastHandler returns timeseries data.
func TimeseriesForecastHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, solutionCtor api.SolutionStorageCtor, trainTestSplitTimeSeries float64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		truthDataset := pat.Param(r, "truthDataset")
		forecastDataset := pat.Param(r, "forecastDataset")
		xColName := pat.Param(r, "xColName")
		yColName := pat.Param(r, "yColName")
		resultUUID := pat.Param(r, "result-uuid")

		// parse POST params
		params, err := parsePostParms(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		// validate the bucket operation
		operation := api.TimeseriesOp(params.DuplicateOperation)
		if operation == "" {
			operation = model.TimeseriesDefaultOp //default
		}

		// get variable names and ranges out of the params
		var filterParams *model.FilterParams
		if params.FilterParams != nil {
			filterParams, err = api.ParseFilterParamsFromJSONRaw(params.FilterParams)
			if err != nil {
				handleError(w, err)
				return
			}
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

		dst, err := meta.FetchDataset(truthDataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		truthStorageName := dst.StorageName

		dsf, err := meta.FetchDataset(forecastDataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		forecastStorageName := dsf.StorageName

		if err != nil {
			handleError(w, err)
			return
		}

		// get the result UUID
		res, err := solution.FetchSolutionResultByUUID(resultUUID)
		if err != nil {
			handleError(w, err)
			return
		}

		// CDB TODO: - need to optimize query for multiple series, mutliple variables
		timeseries := []*model.TimeseriesData{}
		forecasts := []*model.TimeseriesData{}
		for _, t := range params.TimeseriesIDs {
			// fetch the timeseries variable and find the grouping col
			variable, err := meta.FetchVariable(forecastDataset, t.VarKey)
			if err != nil {
				handleError(w, err)
				return
			}

			// fetch timeseries and forecast
			timeseriesData, err := data.FetchTimeseries(truthDataset, truthStorageName, t.VarKey, variable.Grouping.GetIDCol(),
				xColName, yColName, []string{t.SeriesID}, operation, filterParams)
			if err != nil {
				handleError(w, err)
				return
			}
			timeseries = append(timeseries, timeseriesData...)

			// fetch the predicted timeseries
			forecastData, err := data.FetchTimeseriesForecast(forecastDataset, forecastStorageName, t.VarKey, variable.Grouping.GetIDCol(),
				xColName, yColName, []string{t.SeriesID}, operation, res.ResultURI, filterParams)
			if err != nil {
				handleError(w, err)
				return
			}
			forecasts = append(forecasts, forecastData...)
		}

		result := []*TimeseriesForecastResult{}
		for idx, t := range timeseries {
			// Recompute train/test split info for visualization purposes
			split := compute.SplitTimeSeries(t.Timeseries, trainTestSplitTimeSeries)
			forecast := forecasts[idx]
			result = append(result, &TimeseriesForecastResult{
				VarKey:            t.VarKey,
				SeriesID:          t.SeriesID,
				Timeseries:        t.Timeseries,
				Forecast:          forecast.Timeseries,
				ForecastTestRange: []float64{split.SplitValue, split.EndValue},
				IsDateTime:        true,
				Min:               t.Min,
				Max:               t.Max,
				Mean:              t.Mean,
			})
		}
		err = handleJSON(w, result)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
