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

	"goji.io/v3/pat"

	api "github.com/uncharted-distil/distil/api/model"
)

// DeleteHandler inserts a new field based on existing fields.
func DeleteHandler(dataCtor api.DataStorageCtor, esMetaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		variable := pat.Param(r, "variable")

		// initialize the storage
		metaStorage, err := esMetaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		dataStorage, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		ds, err := metaStorage.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName

		// delete the var
		err = metaStorage.DeleteVariable(dataset, variable)
		if err != nil {
			handleError(w, err)
			return
		}
		err = dataStorage.DeleteVariable(dataset, storageName, variable)
		if err != nil {
			handleError(w, err)
			return
		}
	}
}
