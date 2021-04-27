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

	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"
)

// DeletingDatasetHandler attempts to delete mutable datasets
func DeletingDatasetHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get params
		dataset := pat.Param(r, "dataset")
		// get meta and data storage
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

		ds, err := metaStorage.FetchDataset(dataset, true, true, false)
		if err != nil {
			handleError(w, err)
			return
		}
		folder := env.ResolvePath(ds.Source, ds.Folder)

		// figure out if delete is soft
		softDelete := false
		if ds.Immutable {
			log.Infof("doing a soft delete on dataset '%s'", ds.ID)
			softDelete = true
		}

		// delete meta
		err = metaStorage.DeleteDataset(dataset, softDelete)
		if err != nil {
			handleError(w, err)
			return
		}

		// delete the query cache associated wit this dataset if it exists
		task.DeleteQueryCache(dataset)

		if !softDelete {
			// delete db tables
			err = dataStorage.DeleteDataset(ds.StorageName)
			if err != nil {
				handleError(w, err)
				return
			}
			// delete files
			err = util.RemoveContents(folder, true)
			if err != nil {
				handleError(w, err)
				return
			}
		}

		// send json
		err = handleJSON(w, map[string]interface{}{"success": true})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal clustering result into JSON"))
			return
		}
	}
}
