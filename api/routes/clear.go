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
	"goji.io/v3/pat"

	api "github.com/uncharted-distil/distil/api/model"
)

// ClearHandler generates a route handler that enables the clearing of variable,
// optionally being able to specify filter params.
func ClearHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variable key
		variableKey := pat.Param(r, "variable")

		// parse update list
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

		// get storage clients
		metaStorage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		dataStorage, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// need the storage name
		ds, err := metaStorage.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		if ds.Immutable {
			handleError(w, errors.New("can not mutate an immutable dataset"))
			return
		}

		// replace any grouped variables in filter params with the group's
		expandedFilterParams, err := api.ExpandFilterParams(dataset, filterParams, false, metaStorage)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to expand filter params"))
			return
		}

		// need to change unlabeled to empty string
		err = dataStorage.SetVariableValue(dataset, ds.StorageName, variableKey, "", expandedFilterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, map[string]interface{}{
			"result": "success",
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal clustering result into JSON"))
			return
		}
	}
}
