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

package compute

import (
	"fmt"
	"path"

	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
)

func filterData(client *compute.Client, ds *api.Dataset, filterParams *api.FilterParams, dataStorage api.DataStorage) (string, *api.FilterParams, error) {
	inputPath := env.ResolvePath(ds.Source, ds.Folder)

	log.Infof("checking if solution search for dataset %s found in '%s' needs prefiltering", ds.ID, inputPath)
	// determine if filtering is needed
	updatedParams, preFilters := getPreFiltering(ds, filterParams)
	if preFilters.Empty(false) {
		log.Infof("solution request for dataset %s does not need prefiltering", ds.ID)
		return inputPath, updatedParams, nil
	}

	// check if the filtered results already exists
	hash, err := hashFilter(inputPath, preFilters)
	if err != nil {
		return "", nil, err
	}
	outputFolder := env.ResolvePath("tmp", fmt.Sprintf("%s-%v", ds.Folder, hash))
	outputDataFile := path.Join(outputFolder, compute.D3MDataFolder, compute.D3MLearningData)
	if util.FileExists(outputDataFile) {
		log.Infof("solution request for dataset %s already prefiltered at '%s'", ds.ID, outputFolder)
		return outputFolder, updatedParams, nil
	}

	// prepare the data to use for filtering
	outputFolder, err = preparePrefilteringDataset(outputFolder, ds, dataStorage)
	if err != nil {
		return "", nil, err
	}

	// run the filtering pipeline
	pipeline, err := description.CreateDataFilterPipeline("Pre Filtering", "pre filter a dataset that has metadata features", ds.Variables, preFilters.Filters)
	if err != nil {
		return "", nil, err
	}
	filteredData, err := SubmitPipeline(client, []string{outputFolder}, nil, nil, pipeline, true)
	if err != nil {
		return "", nil, err
	}

	// output the filtered results as the data in the filtered dataset
	err = util.CopyFile(filteredData, outputDataFile)
	if err != nil {
		return "", nil, err
	}

	log.Infof("solution request for dataset %s filtered data written to '%s'", ds.ID, outputDataFile)

	return outputFolder, updatedParams, nil
}

func hashFilter(schemaFile string, filterParams *api.FilterParams) (uint64, error) {
	// generate the hash from the params
	hashStruct := struct {
		Schema       string
		FilterParams *api.FilterParams
	}{
		Schema:       schemaFile,
		FilterParams: filterParams,
	}
	hash, err := hashstructure.Hash(hashStruct, nil)
	if err != nil {
		return 0, errors.Wrap(err, "failed to generate filter data hash")
	}
	return hash, nil
}

func getPreFiltering(ds *api.Dataset, filterParams *api.FilterParams) (*api.FilterParams, *api.FilterParams) {
	vars := map[string]*model.Variable{}
	for _, v := range ds.Variables {
		vars[v.Key] = v
	}
	clone := filterParams.Clone()
	// filter if a clustering or outlier detection metadata feature exist
	// remove pre filters from the rest of the filters since they should not be in the main pipeline
	// TODO: NEED TO HANDLE OUTLIER FILTERS!
	preFilters := &api.FilterParams{
		Filters: []*model.Filter{},
	}
	filters := clone.Filters
	clone.Filters = []*model.Filter{}
	for _, f := range filters {
		variable := vars[f.Key]
		params := clone
		if variable.IsGrouping() {
			cg, ok := variable.Grouping.(model.ClusteredGrouping)
			if ok {
				f.Key = cg.GetClusterCol()
				params = preFilters
			}
		}
		params.Filters = append(params.Filters, f)
	}

	return clone, preFilters
}

func preparePrefilteringDataset(outputFolder string, sourceDataset *api.Dataset, dataStorage api.DataStorage) (string, error) {
	// read the data from the database
	data, err := dataStorage.FetchDataset(sourceDataset.ID, sourceDataset.StorageName, true, false, nil)
	if err != nil {
		return "", err
	}

	// if learning dataset, then update that
	if sourceDataset.LearningDataset != "" {
		return UpdatePrefeaturizedDataset(sourceDataset.LearningDataset, sourceDataset, data)
	}

	// update the metadata to match the data pulled from the data storage
	// (mostly matching column index and dropping columns not pulled)
	metaVarMap := map[string]*model.Variable{}
	for _, v := range sourceDataset.Variables {
		metaVarMap[v.Key] = v
	}
	variablesData := make([]*model.Variable, len(data[0]))
	for i, f := range data[0] {
		variablesData[i] = metaVarMap[f]
		variablesData[i].Index = i
	}
	sourceDataset.Variables = variablesData

	// write it out as a dataset
	dsRaw := &api.RawDataset{
		ID:       sourceDataset.ID,
		Name:     sourceDataset.Name,
		Data:     data,
		Metadata: sourceDataset.ToMetadata(),
	}
	err = serialization.WriteDataset(outputFolder, dsRaw)
	if err != nil {
		return "", err
	}

	return outputFolder, nil
}

// UpdatePrefeaturizedDataset updates a featurized dataset that already exists
// on disk to have new variables included
func UpdatePrefeaturizedDataset(prefeaturizedPath string, sourceDataset *api.Dataset, storedData [][]string) (string, error) {
	// load the dataset from disk
	schemaPath := path.Join(prefeaturizedPath, compute.D3MDataSchema)
	dsDisk, err := serialization.ReadDataset(schemaPath)
	if err != nil {
		return "", err
	}
	metaDiskMainDR := dsDisk.Metadata.GetMainDataResource()

	// determine if there are new columns that were not part of the original dataset
	metaDiskVarMap := MapVariables(metaDiskMainDR.Variables, func(variable *model.Variable) string { return variable.Key })
	newVars := []*model.Variable{}
	for _, v := range sourceDataset.Variables {
		if metaDiskVarMap[v.Key] == nil {
			newVars = append(newVars, v)
		}
	}

	// get the index of the new fields in the extracted data
	storedVarMap := MapVariables(sourceDataset.Variables, func(variable *model.Variable) string { return variable.Key })
	storedDataD3MIndex := -1
	for i, v := range storedData[0] {
		if v == model.D3MIndexFieldName {
			storedDataD3MIndex = i
		} else if storedVarMap[v] != nil {
			storedVarMap[v].Index = i
		}
	}

	// add the missing columns row by row and only retain rows in the new dataset
	// first build up the new variables by d3m index map
	// then cycle through the featurized rows and append the variables
	newDataMap := map[string][]string{}
	for _, r := range storedData[1:] {
		newVarsData := []string{}
		for i := 0; i < len(newVars); i++ {
			newVarsData = append(newVarsData, r[newVars[i].Index])
		}
		newDataMap[r[storedDataD3MIndex]] = newVarsData
	}

	// add the new fields to the metadata to generate the proper header
	for i := 0; i < len(newVars); i++ {
		newVar := newVars[i]
		newVar.Index = i + len(metaDiskMainDR.Variables)
		metaDiskMainDR.Variables = append(metaDiskMainDR.Variables, newVar)
	}

	preFeaturizedOutput := [][]string{metaDiskMainDR.GenerateHeader()}
	metaDiskD3MIndex := metaDiskVarMap[model.D3MIndexFieldName].Index
	for _, row := range dsDisk.Data[1:] {
		d3mIndexPre := row[metaDiskD3MIndex]
		if newDataMap[d3mIndexPre] != nil {
			rowComplete := append(row, newDataMap[d3mIndexPre]...)
			preFeaturizedOutput = append(preFeaturizedOutput, rowComplete)
		}
	}

	// output the new pre featurized data
	dsDisk.Data = preFeaturizedOutput
	err = serialization.WriteDataset(prefeaturizedPath, dsDisk)
	if err != nil {
		return "", err
	}

	return prefeaturizedPath, nil
}

// MapVariables creates a variable map using the mapper function to create the key.
func MapVariables(variables []*model.Variable, mapper func(variable *model.Variable) string) map[string]*model.Variable {
	mapped := map[string]*model.Variable{}
	for _, d := range variables {
		mapped[mapper(d)] = d
	}

	return mapped
}
