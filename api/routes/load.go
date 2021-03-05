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

	"goji.io/v3/pat"

	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	log "github.com/unchartedsoftware/plog"
)

// LoadHandler requests that the downstream auto ml system load a model for further use.
func LoadHandler(modelStorageCtor api.ExportedModelStorageCtor, solutionStorageCtor api.SolutionStorageCtor,
	metadataStorageCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		solutionID := pat.Param(r, "solution-id")
		fitted := pat.Param(r, "fitted")
		fittedBool := parseBoolParam(fitted)

		modelStorage, err := modelStorageCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		metadataStorage, err := metadataStorageCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		solutionStorage, err := solutionStorageCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		if fittedBool {
			// Fetch the path to saved solution
			fittedSolution, err := modelStorage.FetchModelByID(solutionID)
			if err != nil {
				handleError(w, err)
				return
			}
			// Request that the ta2 load the solution
			loadedModel, err := task.LoadFittedSolution(fittedSolution.FilePath, solutionStorage, metadataStorage)
			if err != nil {
				handleError(w, err)
				return
			}
			log.Infof("loaded model - %s", loadedModel)
		} else {
			_, err = task.LoadSolution(solutionID)
		}

		if err != nil {
			handleError(w, err)
			return
		}
		log.Infof("Completed export request for %s", solutionID)

		err = handleJSON(w, map[string]interface{}{"solution-id": solutionID, "result": "loaded"})
		if err != nil {
			handleError(w, err)
			return
		}
	}
}
