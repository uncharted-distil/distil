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

	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
)

// OutlierDetectionHandler generates a route handler that enables outlier detection
// for either remote sensing or tabular data
func OutlierDetectionHandler(metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		// get variable name
		variable := pat.Param(r, "variable")

		// get storage clients
		metaStorage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		datasetMeta, err := metaStorage.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		outlierData, err := task.OutlierDetection(datasetMeta, variable)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, outlierData)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal outlier data into JSON"))
			return
		}
	}
}
