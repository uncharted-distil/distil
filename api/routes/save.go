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
	"strconv"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	log "github.com/unchartedsoftware/plog"
)

func parseBoolParam(value string) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		parsed = false
	}

	return parsed
}

// SaveHandler exports the caller supplied solution by calling through to the compute
// server export functionality.
func SaveHandler(modelStorageCtor api.ExportedModelStorageCtor, solutionStorageCtor api.SolutionStorageCtor,
	metadataStorageCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		solutionID := pat.Param(r, "solution-id")
		fitted := pat.Param(r, "fitted")
		fittedBool := parseBoolParam(fitted)

		modelStorage, err := modelStorageCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to create model storage client"))
			return
		}

		metadataStorage, err := metadataStorageCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to create metadata storage client"))
			return
		}

		solutionStorage, err := solutionStorageCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to create solution storage client"))
			return
		}

		if fittedBool {
			var exported *api.ExportedModel
			exported, err = task.SaveFittedSolution(solutionID, solutionStorage, metadataStorage)
			if err != nil {
				handleError(w, errors.Wrap(err, "failed saving fitted solution"))
				return
			}

			err = modelStorage.PersistExportedModel(exported)
		} else {
			_, err = task.SaveSolution(solutionID)
		}

		if err != nil {
			handleError(w, errors.Wrap(err, "failed solution export request"))
			return
		}
		log.Infof("Completed export request for %s", solutionID)

		err = handleJSON(w, map[string]interface{}{"solution-id": solutionID, "result": "saved"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal save result into JSON"))
			return
		}
	}
}
