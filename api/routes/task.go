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

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	apiCompute "github.com/uncharted-distil/distil/api/compute"
	api "github.com/uncharted-distil/distil/api/model"
)

// TaskHandler determines modeling task based on dataset and target variable.
func TaskHandler(dataCtor api.DataStorageCtor, esMetaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		variableName := pat.Param(r, "target")
		storageName := model.NormalizeDatasetID(dataset)

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

		// look up the target variable
		variable, err := metaStorage.FetchVariable(dataset, variableName)
		if err != nil {
			handleError(w, err)
			return
		}

		// resolve the task based on the dataset and target
		task, err := apiCompute.ResolveTask(dataStorage, storageName, variable)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data
		err = handleJSON(w, task)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
