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
	"goji.io/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// VariablesResult represents the result of a variables response.
type VariablesResult struct {
	Variables []*model.Variable `json:"variables"`
}

// VariablesHandler generates a route handler that facilitates a search of
// variable names and descriptions, returning a variable list for the specified
// dataset.
func VariablesHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		storageName := model.NormalizeDatasetID(dataset)
		// get elasticsearch client
		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		data, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		// fetch variables
		variables, err := meta.FetchVariables(dataset, false, true, false)
		if err != nil {
			handleError(w, err)
			return
		}
		for _, v := range variables {
			if model.IsNumerical(v.Type) || model.IsDateTime(v.Type) {
				extrema, err := data.FetchExtrema(storageName, v)
				if err != nil {
					handleError(w, err)
					return
				}
				v.Min = extrema.Min
				v.Max = extrema.Max
			}
		}
		// marshal data
		err = handleJSON(w, VariablesResult{
			Variables: variables,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
