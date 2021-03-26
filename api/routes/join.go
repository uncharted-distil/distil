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
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
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

		// fetch vars from params
		datasetLeft := params["datasetLeft"].(map[string]interface{})
		datasetRight := params["datasetRight"].(map[string]interface{})

		leftJoin := &task.JoinSpec{
			DatasetID:     datasetLeft["id"].(string),
			DatasetFolder: datasetLeft["datasetFolder"].(string),
			DatasetSource: metadata.DatasetSource(datasetLeft["source"].(string)),
		}

		rightJoin := &task.JoinSpec{
			DatasetID:     datasetRight["id"].(string),
			DatasetFolder: datasetRight["datasetFolder"].(string),
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

		// add d3m variables to left variables
		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		d3mIndexVar, err := meta.FetchVariable(datasetLeft["id"].(string), model.D3MIndexFieldName)
		if err != nil {
			handleError(w, err)
			return
		}
		leftVariables = append(leftVariables, d3mIndexVar)

		// run joining pipeline
		path, data, err := join(leftJoin, rightJoin, leftVariables, rightVariables, datasetRight, params, meta)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		bytes, err := json.Marshal(map[string]interface{}{"path": path, "data": data})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal filtered data result into JSON"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(bytes)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to write filtered data to response writer"))
			return
		}
	}
}

func parseVariables(variablesRaw []interface{}) ([]*model.Variable, error) {
	variables := make([]*model.Variable, len(variablesRaw))
	for i, varRaw := range variablesRaw {
		varData := varRaw.(map[string]interface{})
		// groups need to be handled separately as they depend on type
		var groupingParsed model.BaseGrouping
		if varData["grouping"] != nil {
			if model.IsTimeSeries(varData["colType"].(string)) {
				groupingTimeseries := model.TimeseriesGrouping{}
				err := json.MapToStruct(&groupingTimeseries, varData["grouping"].(map[string]interface{}))
				if err != nil {
					return nil, errors.Wrap(err, "Unable to parse timeseries grouping")
				}
				groupingParsed = &groupingTimeseries
			}
			varData["grouping"] = nil
		}
		v := model.Variable{}
		err := json.MapToStruct(&v, varData)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse Variables")
		}
		v.Grouping = groupingParsed
		variables[i] = &v
	}

	return variables, nil
}

func join(joinLeft *task.JoinSpec, joinRight *task.JoinSpec, varsLeft []*model.Variable, varsRight []*model.Variable,
	datasetRight map[string]interface{}, params map[string]interface{}, metaStorage api.MetadataStorage) (string, *api.FilteredData, error) {
	// determine if distil or datamart
	if params["searchResultIndex"] == nil {
		return joinDistil(joinLeft, joinRight, params, metaStorage)
	}

	return joinDatamart(joinLeft, joinRight, varsLeft, varsRight, datasetRight, params)
}

func joinDistil(joinLeft *task.JoinSpec, joinRight *task.JoinSpec,
	params map[string]interface{}, metaStorage api.MetadataStorage) (string, *api.FilteredData, error) {
	if params["datasetAColumn"] == nil {
		return "", nil, errors.Errorf("missing parameter 'datasetAColumn'")
	}
	leftCol := params["datasetAColumn"].(string)
	if params["datasetBColumn"] == nil {
		return "", nil, errors.Errorf("missing parameter 'datasetBColumn'")
	}
	rightCol := params["datasetBColumn"].(string)
	if params["accuracy"] == nil {
		return "", nil, errors.Errorf("missing parameter 'accuracy'")
	}
	accuracy := params["accuracy"].(float32)
	// need to read variables from disk for the variable list
	metaLeft, err := getDiskMetadata(joinLeft.DatasetID, metaStorage)
	if err != nil {
		return "", nil, err
	}
	metaRight, err := getDiskMetadata(joinRight.DatasetID, metaStorage)
	if err != nil {
		return "", nil, err
	}
	joinLeft.Variables = metaLeft.GetMainDataResource().Variables
	joinRight.Variables = metaRight.GetMainDataResource().Variables
	joinLeft.ExistingMetadata = metaLeft
	joinRight.ExistingMetadata = metaRight

	path, data, err := task.JoinDistil(joinLeft, joinRight, leftCol, rightCol, accuracy)
	if err != nil {
		return "", nil, err
	}

	return path, data, nil
}

func joinDatamart(joinLeft *task.JoinSpec, joinRight *task.JoinSpec, varsLeft []*model.Variable,
	varsRight []*model.Variable, datasetRight map[string]interface{}, params map[string]interface{}) (string, *api.FilteredData, error) {
	if params["searchResultIndex"] == nil {
		return "", nil, errors.Errorf("missing parameter 'searchResultIndex'")
	}
	searchResultIndex := int(params["searchResultIndex"].(float64))

	// need to find the right join suggestion since a single dataset
	// can have multiple join suggestions
	if datasetRight["joinSuggestion"] == nil {
		return "", nil, errors.Errorf("Join Suggestion undefined")
	}

	joinSuggestions := datasetRight["joinSuggestion"].([]interface{})
	targetJoin := joinSuggestions[searchResultIndex].(map[string]interface{})
	if targetJoin == nil {
		return "", nil, errors.Errorf("Unable to find join suggestion at search result index")
	}

	targetJoinOrigin := targetJoin["datasetOrigin"].(map[string]interface{})
	if targetJoinOrigin == nil {
		return "", nil, errors.Errorf("Unable to find join origin")
	}

	targetOriginModel := model.DatasetOrigin{}
	err := json.MapToStruct(&targetOriginModel, targetJoinOrigin)
	if err != nil {
		return "", nil, errors.Wrap(err, "Unable to parse join origin from JSON")
	}
	joinLeft.Variables = varsLeft
	joinRight.Variables = varsRight

	// run joining pipeline
	path, data, err := task.JoinDatamart(joinLeft, joinRight, &targetOriginModel)
	if err != nil {
		return "", nil, err
	}

	return path, data, nil
}

func getDiskMetadata(dataset string, metaStorage api.MetadataStorage) (*model.Metadata, error) {
	ds, err := metaStorage.FetchDataset(dataset, true, true, true)
	if err != nil {
		return nil, err
	}

	folderPath := env.ResolvePath(ds.Source, ds.Folder)
	dsDisk, err := serialization.ReadDataset(path.Join(folderPath, compute.D3MDataSchema))
	if err != nil {
		return nil, err
	}

	return dsDisk.Metadata, nil
}
