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
//

package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"goji.io/v3/pat"
)

// AddFieldHandler generates a route handler that adds columns to datasets
// expects at least two parameters "name" of the field and "fieldType" type of the field. Optional parameter is "defaultValue"
func AddFieldHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")

		// parse update list
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		if params["name"] == nil || params["fieldType"] == nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
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
		errorMsg := "Error casting param "
		storageName := ds.StorageName
		name, ok := params["name"].(string)
		if !ok {
			handleError(w, errors.New(errorMsg+"name"))
			return
		}
		fieldType, ok := params["fieldType"].(string)
		if !ok {
			handleError(w, errors.New(errorMsg+"fieldType"))
			return
		}
		defaultValue := ""
		// update postgres
		if params["defaultValue"] != nil {
			defaultValue, ok = params["defaultValue"].(string)
			if !ok {
				handleError(w, errors.New(errorMsg+"defaultValue"))
				return
			}
		}
		err = dataStorage.AddVariable(dataset, storageName, name, fieldType, defaultValue)
		if err != nil {
			handleError(w, err)
			return
		}
		displayName := name
		if params["displayName"] != nil {
			displayName, ok = params["displayName"].(string)
			if !ok {
				displayName = name
			}
		}
		// update elasticsearch
		err = metaStorage.AddVariable(dataset, name, displayName, fieldType, model.VarDistilRoleData)
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
