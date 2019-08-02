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
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-ingest/metadata"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util/json"
)

func missingParamErr(w http.ResponseWriter, paramName string) {
	handleError(w, errors.Errorf(paramName+" needed for joined dataset import"))
}

// JoinHandler generates a route handler that joins two datasets using caller supplied
// columns.  The joined data is returned to the caller, but is NOT added to storage.
func JoinHandler(metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	fmt.Printf("testing\n")
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("testing 2\n")

		//replace with pulling info out of the post json here

		// parse JSON from post
		params, err := getPostParameters(r)
		if err != nil {
			fmt.Printf("%v\n", err)
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		if params == nil {
			missingParamErr(w, "parameters")
			return
		}

		if params["datasetLeft"] == nil {
			missingParamErr(w, "datasetLeft")
			return
		}

		if params["datasetRight"] == nil {
			missingParamErr(w, "datasetRight")
			return
		}

		// fetch vars for each dataset
		datasetLeft := params["datasetLeft"].(map[string]interface{})
		datasetRight := params["datasetRight"].(map[string]interface{})

		fmt.Printf("dsl: %+v\n\n\n\ndsr: %+v\n\n\n\n", datasetLeft["joinSuggestion"], datasetRight["joinSuggestion"])

		leftJoin := &task.JoinSpec{
			DatasetID:     datasetLeft["id"].(string),
			DatasetFolder: datasetLeft["folder"].(string),
			DatasetSource: metadata.DatasetSource(datasetLeft["source"].(string)),
		}

		rightJoin := &task.JoinSpec{
			DatasetID:     datasetRight["id"].(string),
			DatasetFolder: datasetRight["folder"].(string),
			DatasetSource: metadata.DatasetSource(datasetRight["source"].(string)),
		}

		leftVariables := datasetLeft["variables"].([]*model.Variable)
		rightVariables := datasetRight["variables"].([]*model.Variable)

		fmt.Printf("jsl: %+v\n\n\n\njsr: %+v\n\n\n\n", leftJoin, rightJoin)

		// need to find the right join suggestion since a single dataset
		// can have multiple join suggestions
		var origin model.DatasetOrigin
		if datasetRight["joinSuggestion"] != nil {
			joinSuggestion := datasetRight["joinSuggestion"].(map[string]interface{})
			origin = joinSuggestion["datasetOrigin"].(model.DatasetOrigin)
			fmt.Printf("%v\n\n\n", origin)
		}

		originRef := &origin

		// run joining pipeline
		data, err := task.Join(leftJoin, rightJoin, leftVariables, rightVariables, originRef)
		if err != nil {
			fmt.Printf("%v\n", err)
			handleError(w, err)
			return
		}

		// marshal output into JSON
		bytes, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("%v\n", err)
			handleError(w, errors.Wrap(err, "unable marshal filtered data result into JSON"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}
