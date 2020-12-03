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
	"github.com/pkg/errors"
	api "github.com/uncharted-distil/distil/api/model"
	"goji.io/v3/pat"
	"net/http"
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
		invert := pat.Param(r, "invert")
		invertBool := parseBoolParam(invert)

		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}
		t, ok := params["timeseriesUris"].([]interface{})
		if !ok {
			handleError(w, errors.New("Missing timeseriesUris from query"))
			return
		}
		timeseriesURIs := []string{}
		for _, v := range t {
			s, ok := v.(string)
			if !ok {
				return
			}
			timeseriesURIs = append(timeseriesURIs, s)
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
		timeseries, err := storage.FetchTimeseries(dataset, storageName, timeseriesColName, xColName, yColName, timeseriesURIs, filterParams, invertBool)
		if err != nil {
			handleError(w, err)
			return
		}

		err = handleJSON(w, timeseries)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
