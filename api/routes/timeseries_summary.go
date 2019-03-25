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
	"goji.io/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// TimeseriesSummaryResult represents a summary response for a variable.
type TimeseriesSummaryResult struct {
	Timeseries [][]float64 `json:"timeseries"`
}

// TimeseriesSummaryHandler generates a route handler that facilitates the
// creation and retrieval of summary information about the specified variable.
func TimeseriesSummaryHandler(ctorStorage api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		dataset := pat.Param(r, "dataset")
		xColName := pat.Param(r, "xColName")
		yColName := pat.Param(r, "yColName")
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
		storage, err := ctorStorage()
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch summary histogram
		timeseries, err := storage.FetchTimeseriesSummary(dataset, storageName, xColName, yColName, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, TimeseriesSummaryResult{
			Timeseries: timeseries,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
	}
}
