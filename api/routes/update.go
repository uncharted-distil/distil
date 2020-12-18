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

	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
)

// UpdateHandler generates a route handler that enables clustering
// of a variable and the creation of the new column to hold the cluster label.
func UpdateHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, config env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")

		// parse update list
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		if params["updates"] == nil {
			missingParamErr(w, "updates")
			return
		}

		updates, err := api.ParseVariableUpdateList(params)
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

		// need to update one variable at a time
		mappedUpdates := map[string]map[string]string{}
		for _, u := range updates {
			if mappedUpdates[u.Name] == nil {
				mappedUpdates[u.Name] = map[string]string{}
			}

			mappedUpdates[u.Name][u.Index] = u.Value
		}

		// need the storage name
		ds, err := metaStorage.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName

		updateCount := 0
		for varName, mapped := range mappedUpdates {
			err = dataStorage.UpdateVariableBatch(storageName, varName, mapped)
			if err != nil {
				handleError(w, err)
				return
			}
			updateCount = updateCount + len(mapped)
		}

		// marshal output into JSON
		err = handleJSON(w, map[string]interface{}{
			"result": "success",
			"count":  updateCount,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal clustering result into JSON"))
			return
		}
	}
}
