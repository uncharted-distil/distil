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
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/model"
	api "github.com/uncharted-distil/distil/api/model"
	"goji.io/v3/pat"
)

// TimeseriesHandler returns timeseries data.
func TimeseriesHandler(metaCtor api.MetadataStorageCtor, ctorStorage api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		dataset := pat.Param(r, "dataset")
		xColName := pat.Param(r, "xColName")
		yColName := pat.Param(r, "yColName")

		// parse POST params
		params, err := parsePostParms(r)
		if err != nil {
			handleError(w, err)
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

		ds, err := meta.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName

		// CDB TODO: - need to optimize query for multiple series, mutliple variables
		timeseries := []*model.TimeseriesData{}
		for _, t := range params.TimeseriesIDs {
			// fetch the timeseries variable and find the grouping col
			variable, err := meta.FetchVariable(dataset, t.VarKey)
			if err != nil {
				handleError(w, err)
				return
			}

			// fetch timeseries
			timeseriesData, err := storage.FetchTimeseries(dataset, storageName, t.VarKey, variable.Grouping.GetIDCol(),
				xColName, yColName, []string{t.SeriesID}, operation, filterParams)
			if err != nil {
				handleError(w, err)
				return
			}
			timeseries = append(timeseries, timeseriesData...)
		}

		err = handleJSON(w, timeseries)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}

type timeseriesParams struct {
	TimeseriesIDs []struct {
		SeriesID string `json:"seriesID"`
		VarKey   string `json:"varKey"`
	} `json:"timeseries"`
	DuplicateOperation string          `json:"duplicateOperation"`
	FilterParams       json.RawMessage `json:"filterParams"`
}

// parse post parameters into a structure
func parsePostParms(r *http.Request) (*timeseriesParams, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse POST request")
	}

	var params timeseriesParams
	if err = json.Unmarshal(body, &params); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshall post request params")
	}

	return &params, nil
}
