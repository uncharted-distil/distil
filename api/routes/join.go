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
	return func(w http.ResponseWriter, r *http.Request) {
		// parse JSON from post
		params, err := getPostParameters(r)
		if err != nil {
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

		if params["searchResultIndex"] == nil {
			missingParamErr(w, "searchResultIndex")
			return
		}

		// fetch vars from params
		datasetLeft := params["datasetLeft"].(map[string]interface{})
		datasetRight := params["datasetRight"].(map[string]interface{})
		searchResultIndex := int(params["searchResultIndex"].(float64))

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

		leftVariables, err := parseVariables(datasetLeft["variables"].([]interface{}))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to parse left variables"))
			return
		}
		rightVariables, err := parseVariables(datasetRight["variables"].([]interface{}))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to parse right variables"))
			return
		}

		// need to find the right join suggestion since a single dataset
		// can have multiple join suggestions
		if datasetRight["joinSuggestion"] == nil {
			handleError(w, errors.Wrap(err, "Join Suggestion undefined"))
			return
		}

		joinSuggestions := datasetRight["joinSuggestion"].([]interface{})
		targetJoin := joinSuggestions[searchResultIndex].(map[string]interface{})
		if targetJoin == nil {
			handleError(w, errors.Wrap(err, "Unable to find join suggestion at search result index"))
			return
		}

		targetJoinOrigin := targetJoin["datasetOrigin"].(map[string]interface{})
		if targetJoinOrigin == nil {
			handleError(w, errors.Wrap(err, "Unable to find join origin"))
			return
		}

		targetOriginModel := model.DatasetOrigin{}
		err = json.MapToStruct(&targetOriginModel, targetJoinOrigin)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse join origin from JSON"))
			return
		}

		// run joining pipeline
		data, err := task.Join(leftJoin, rightJoin, leftVariables, rightVariables, &targetOriginModel)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		bytes, err := json.Marshal(data)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal filtered data result into JSON"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

func parseVariables(variablesRaw []interface{}) ([]*model.Variable, error) {
	variables := make([]*model.Variable, len(variablesRaw))
	for i, varRaw := range variablesRaw {
		v := model.Variable{}
		err := json.MapToStruct(&v, varRaw.(map[string]interface{}))
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse Variables")
		}
		variables[i] = &v
	}

	return variables, nil
}
