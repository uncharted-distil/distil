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
	api "github.com/uncharted-distil/distil/api/model"
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"
)

// DeletingModelHandler attempts to delete an exported model.
func DeletingModelHandler(modelCtor api.ExportedModelStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get params
		fittedSolutionID := pat.Param(r, "model")
		// get meta and data storage
		modelStorage, err := modelCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// delete meta
		log.Infof("deleting model %s", fittedSolutionID)
		err = modelStorage.DeleteModel(fittedSolutionID)
		if err != nil {
			handleError(w, err)
			return
		}

		// send json
		err = handleJSON(w, map[string]interface{}{"success": true})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal clustering result into JSON"))
			return
		}
	}
}
