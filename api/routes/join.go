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
	log "github.com/unchartedsoftware/plog"
)

func missingParamErr(w http.ResponseWriter, paramName string) {
	handleError(w, errors.Errorf(paramName+" needed for joined dataset import"))
}

// JoinHandler generates a route handler that joins two datasets using caller supplied
// columns.  The joined data is returned to the caller, but is NOT added to storage.
func JoinHandler(dataCtor api.DataStorageCtor, metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
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
			DatasetSource: metadata.DatasetSource(datasetLeft["source"].(string)),
		}
		leftJoin.DatasetPath = env.ResolvePath(leftJoin.DatasetSource, datasetLeft["datasetFolder"].(string))

		rightJoin := &task.JoinSpec{
			DatasetID:     datasetRight["id"].(string),
			DatasetSource: metadata.DatasetSource(datasetRight["source"].(string)),
		}
		rightJoin.DatasetPath = env.ResolvePath(rightJoin.DatasetSource, datasetRight["datasetFolder"].(string))

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

		// When joining a multiband image dataset with another we always force the multiband dataset
		// to be the left.  Because we perform a left outer join, this ensures that we effectively clip the
		// data to the area we have imagery for.
		for _, v := range rightVariables {
			if model.IsMultiBandImage(v.Type) {
				log.Warnf("Multiband image set %s used as right join argument and will be forced to the left.", rightJoin.DatasetID)
				temp := leftJoin
				leftJoin = rightJoin
				rightJoin = temp

				tempVars := leftVariables
				leftVariables = rightVariables
				rightVariables = tempVars

				tempDataset := datasetLeft
				datasetLeft = datasetRight
				datasetRight = tempDataset

				joinPairsRaw, ok := json.Array(params, "joinPairs")
				if !ok {
					handleError(w, errors.Errorf("joinPairs not a list of join pairs"))
					return
				}
				for _, p := range joinPairsRaw {
					temp := p["first"]
					p["first"] = p["second"]
					p["second"] = temp
				}

				break
			}
		}

		dataStorage, err := dataCtor()
		if err != nil {
			handleError(w, err)
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
		path, data, err := join(leftJoin, rightJoin, leftVariables, rightVariables, datasetRight, params, dataStorage, meta)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		bytes, err := json.Marshal(map[string]interface{}{"path": path, "data": transformDataForClient(data, api.EmptyString)})
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
			groupingType := varData["colType"].(string)
			if model.IsTimeSeries(groupingType) {
				groupingTimeseries := model.TimeseriesGrouping{}
				err := json.MapToStruct(&groupingTimeseries, varData["grouping"].(map[string]interface{}))
				if err != nil {
					return nil, errors.Wrap(err, "Unable to parse timeseries grouping")
				}
				groupingParsed = &groupingTimeseries
			} else if model.IsGeoBounds(groupingType) {
				groupingGeo := model.GeoBoundsGrouping{}
				err := json.MapToStruct(&groupingGeo, varData["grouping"].(map[string]interface{}))
				if err != nil {
					return nil, errors.Wrap(err, "Unable to parse geobounds grouping")
				}
				groupingParsed = &groupingGeo
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

func join(joinLeft *task.JoinSpec, joinRight *task.JoinSpec, varsLeft []*model.Variable,
	varsRight []*model.Variable, datasetRight map[string]interface{}, params map[string]interface{},
	dataStorage api.DataStorage, metaStorage api.MetadataStorage) (string, *api.FilteredData, error) {
	// determine if distil or datamart
	if params["searchResultIndex"] == nil {
		return joinDistil(joinLeft, joinRight, params, dataStorage, metaStorage)
	}

	joinLeft.ExistingMetadata = &model.Metadata{
		DataResources: []*model.DataResource{{
			Variables: varsLeft,
		}},
	}
	joinRight.ExistingMetadata = &model.Metadata{
		DataResources: []*model.DataResource{{
			Variables: varsRight,
		}},
	}

	return joinDatamart(joinLeft, joinRight, varsLeft, varsRight, datasetRight, params)
}

func joinDistil(joinLeft *task.JoinSpec, joinRight *task.JoinSpec, params map[string]interface{},
	dataStorage api.DataStorage, metaStorage api.MetadataStorage) (string, *api.FilteredData, error) {
	if params["joinPairs"] == nil {
		return "", nil, errors.Errorf("missing parameter 'joinPairs'")
	}

	accuracy, ok := params["accuracy"].([]interface{})
	if !ok {
		return "", nil, errors.Errorf("error converting accuracy to array interface")
	}
	absoluteAccuracy, ok := params["absoluteAccuracy"].([]interface{})
	if !ok {
		return "", nil, errors.Errorf("error converting absolute accuracy to array interface")
	}

	joinPairsRaw, ok := json.Array(params, "joinPairs")
	if !ok {
		return "", nil, errors.Errorf("joinPairs not a list of join pairs")
	}
	if len(accuracy) != len(joinPairsRaw) {
		return "", nil, errors.Errorf("accuracy length does not match join pairs length")
	}
	if len(accuracy) != len(absoluteAccuracy) {
		return "", nil, errors.Errorf("accuracy length does not match absolute accuracy length")
	}
	joinPairs := make([]*task.JoinPair, len(joinPairsRaw))
	for i, p := range joinPairsRaw {
		leftColName, ok := p["first"].(string)
		if !ok {
			return "", nil, errors.Errorf("join pair 'first' value is not a string")
		}

		rightColName, ok := p["second"].(string)
		if !ok {
			return "", nil, errors.Errorf("join pair 'second' value is not a string")
		}

		acc, ok := accuracy[i].(float64)
		if !ok {
			return "", nil, errors.Errorf("error converting accuracy to float64")
		}

		absolute, ok := absoluteAccuracy[i].(bool)
		if !ok {
			return "", nil, errors.Errorf("error converting absolute accuracy to bool")
		}
		joinPairs[i] = &task.JoinPair{
			Left:             leftColName,
			Right:            rightColName,
			Accuracy:         acc,
			AbsoluteAccuracy: absolute,
		}
	}

	// need to read variables from disk for the variable list
	metaLeft, err := getDiskMetadata(joinLeft.DatasetID, metaStorage, false)
	if err != nil {
		return "", nil, err
	}
	metaRight, err := getDiskMetadata(joinRight.DatasetID, metaStorage, false)
	if err != nil {
		return "", nil, err
	}

	dsLeft, err := metaStorage.FetchDataset(joinLeft.DatasetID, true, true, true)
	if err != nil {
		return "", nil, err
	}
	dsRight, err := metaStorage.FetchDataset(joinRight.DatasetID, true, true, true)
	if err != nil {
		return "", nil, err
	}

	joinLeft.UpdatedVariables = dsLeft.Variables
	joinRight.UpdatedVariables = dsRight.Variables
	joinLeft.ExistingMetadata = metaLeft
	joinRight.ExistingMetadata = metaRight

	var path string
	var data *api.FilteredData
	if dsLeft.LearningDataset != "" {
		path, data, err = joinPrefeaturized(dataStorage, metaStorage, joinLeft, joinRight, joinPairs)
	} else {
		path, data, err = task.JoinDistil(dataStorage, joinLeft, joinRight, joinPairs, false)
	}
	if err != nil {
		return "", nil, err
	}

	return path, data, nil
}

func joinPrefeaturized(dataStorage api.DataStorage, metaStorage api.MetadataStorage, joinLeft *task.JoinSpec,
	joinRight *task.JoinSpec, joinPairs []*task.JoinPair) (string, *api.FilteredData, error) {

	log.Infof("joining a prefeaturized dataset")
	// switch the left join info to point to the learning dataset
	sourceVarMap := api.MapVariables(joinLeft.UpdatedVariables, func(variable *model.Variable) string { return variable.Key })
	metaLeft, err := getDiskMetadata(joinLeft.DatasetID, metaStorage, true)
	if err != nil {
		return "", nil, err
	}
	joinLeft.ExistingMetadata = metaLeft
	joinLeft.DatasetPath = metaLeft.LearningDataset

	// join as normal
	path, data, err := task.JoinDistil(dataStorage, joinLeft, joinRight, joinPairs, true)
	if err != nil {
		return "", nil, err
	}

	// update the source data to have the joined data
	// build header for data to add & extract columns to keep
	prefeaturizedUpdates := [][]string{{}}
	newCols := []int{}
	for vName, v := range data.Columns {
		if v.Key == model.D3MIndexFieldName || sourceVarMap[v.Key] == nil {
			newCols = append(newCols, v.Index)
			prefeaturizedUpdates[0] = append(prefeaturizedUpdates[0], vName)
		}
	}

	// cycle through the data and copy over the new fields
	for _, r := range data.Values {
		newRow := []string{}
		for _, c := range newCols {
			newRow = append(newRow, r[c].Value.(string))
		}
		prefeaturizedUpdates = append(prefeaturizedUpdates, newRow)
	}

	// read the source dataset
	diskDataset, err := api.LoadDiskDatasetFromFolder(joinLeft.DatasetPath)
	if err != nil {
		return "", nil, err
	}

	// update the base dataset with the changes and write the updated data to disk
	err = diskDataset.UpdateRawData(sourceVarMap, prefeaturizedUpdates, false)
	if err != nil {
		return "", nil, err
	}

	// store the raw data to disk
	log.Infof("writing updated source data to %s", path)
	err = diskDataset.SaveDataset()
	if err != nil {
		return "", nil, err
	}

	dataParsed, err := api.CreateFilteredData(diskDataset.Dataset.Data, diskDataset.Dataset.Metadata.GetMainDataResource().Variables, false, 100)
	if err != nil {
		return "", nil, err
	}

	return path, dataParsed, nil
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
	joinLeft.UpdatedVariables = varsLeft
	joinRight.UpdatedVariables = varsRight

	// run joining pipeline
	path, data, err := task.JoinDatamart(joinLeft, joinRight, &targetOriginModel)
	if err != nil {
		return "", nil, err
	}

	return path, data, nil
}

func getDiskMetadata(dataset string, metaStorage api.MetadataStorage, useLearningFolder bool) (*model.Metadata, error) {
	ds, err := metaStorage.FetchDataset(dataset, true, true, true)
	if err != nil {
		return nil, err
	}

	folderPath := env.ResolvePath(ds.Source, ds.Folder)
	if useLearningFolder {
		folderPath = ds.GetLearningFolder()
	}

	dsDisk, err := serialization.ReadDataset(path.Join(folderPath, compute.D3MDataSchema))
	if err != nil {
		return nil, err
	}
	dsDisk.Metadata.LearningDataset = ds.LearningDataset

	return dsDisk.Metadata, nil
}
