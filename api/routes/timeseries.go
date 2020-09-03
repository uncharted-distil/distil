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

	api "github.com/uncharted-distil/distil/api/model"
)

// TimeseriesResult represents the result of a timeseries request.
type TimeseriesResult struct {
	Timeseries []*api.TimeseriesObservation `json:"timeseries"`
	IsDateTime bool                         `json:"isDateTime"`
	Min        api.NullableFloat64          `json:"min"`
	Max        api.NullableFloat64          `json:"max"`
	Mean       api.NullableFloat64          `json:"mean"`
}

// TimeseriesHandler returns timeseries data.
func TimeseriesHandler(metaCtor api.MetadataStorageCtor, ctorStorage api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		dataset := pat.Param(r, "dataset")
		timeseriesColName := pat.Param(r, "timeseriesColName")
		xColName := pat.Param(r, "xColName")
		yColName := pat.Param(r, "yColName")
		timeseriesURI := pat.Param(r, "timeseriesURI")
		invert := pat.Param(r, "invert")
		invertBool := parseBoolParam(invert)

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
		storage, err := ctorStorage()
		if err != nil {
			handleError(w, err)
			return
		}

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		ds, err := meta.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName

		// fetch timeseries
		timeseries, err := storage.FetchTimeseries(dataset, storageName, timeseriesColName, xColName, yColName, timeseriesURI, filterParams, invertBool)
		if err != nil {
			handleError(w, err)
			return
		}

		err = handleJSON(w, TimeseriesResult{
			Timeseries: timeseries.Timeseries,
			IsDateTime: timeseries.IsDateTime,
			Min:        api.NullableFloat64(timeseries.Min),
			Max:        api.NullableFloat64(timeseries.Max),
			Mean:       api.NullableFloat64(timeseries.Mean),
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
