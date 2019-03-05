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
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
)

// GroupingResult represents a summary response for a grouping.
type GroupingResult struct {
	Histogram *api.Histogram `json:"histogram"`
}

// GroupingSummaryHandler generates a route handler that facilitates the
// creation and retrieval of summary information about the specified groupings.
func GroupingSummaryHandler(ctorStorage api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		storageName := model.NormalizeDatasetID(dataset)

		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		groupType, ok := json.String(params, "type")
		if !ok {
			handleError(w, fmt.Errorf("no `type` argument"))
		}

		// groupID, ok := json.String(params, "idCol")
		// if !ok {
		// 	handleError(w, fmt.Errorf("no `idCol` argument"))
		// }

		// TODO: only support timeseries groupings so far
		properties, ok := json.Get(params, "properties")
		if !ok {
			handleError(w, fmt.Errorf("no `properties` argument"))
		}

		clusterCol, ok := json.String(properties, "clusterCol")
		if !ok {
			handleError(w, fmt.Errorf("no `clusterCol` argument"))
		}

		// get storage client
		storage, err := ctorStorage()
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch summary histogram
		histogram, err := storage.FetchGroup(dataset, storageName, clusterCol, groupType)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, SummaryResult{
			Histogram: histogram,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
	}
}
