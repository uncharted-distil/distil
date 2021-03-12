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
	inputPath := ds.GetLearningFolder()

	log.Infof("checking if solution search for dataset %s found in '%s' needs prefiltering", ds.ID, inputPath)
	// determine if filtering is needed
	updatedParams, preFilters := getPreFiltering(ds, filterParams)
	if preFilters.Empty(false) {
		log.Infof("solution request for dataset %s does not need prefiltering", ds.ID)
		return inputPath, updatedParams, nil
	}

	// check if the filtered results already exists
	// TODO: JUST BECAUSE THE FILE EXISTS DOESNT MEAN THE CONTENTS IS GOOD
	// SHOULD PROBABLY WRITE TO A TMP FOLDER AND COPY THE RESULTS OVER IF EVERYTHING WORKED!
	// (OR DO SOMETHING ELSE TO GUARANTEE THAT FILE EXISTING = FILTERING WORKED)
	hash, err := hashFilter(inputPath, preFilters)
	if err != nil {
		return "", nil, err
	}
	outputFolder := env.ResolvePath("tmp", fmt.Sprintf("%s-%v", ds.Folder, hash))
	outputExists, _ := getPreFilteringOutputDataFile(outputFolder)
	if outputExists {
		log.Infof("solution request for dataset %s already prefiltered at '%s'", ds.ID, outputFolder)
		return outputFolder, updatedParams, nil
	}

	// prepare the data to use for filtering
	resultingVariables, err := preparePrefilteringDataset(outputFolder, ds, dataStorage)
	if err != nil {
		return "", nil, err
	}

	// run the filtering pipeline
	pipeline, err := description.CreateDataFilterPipeline("Pre Filtering", "pre filter a dataset that has metadata features", resultingVariables, preFilters.Filters)
	if err != nil {
		return "", nil, err
	}

	var allowableTypes []string
	if ds.LearningDataset == "" {
		allowableTypes = append(allowableTypes, compute.CSVURIValueType)
	} else {
		allowableTypes = append(allowableTypes, compute.ParquetURIValueType)
		allowableTypes = append(allowableTypes, compute.CSVURIValueType)
	}
	filteredData, err := SubmitPipeline(client, []string{outputFolder}, nil, nil, pipeline, allowableTypes, true)
	if err != nil {
		return "", nil, err
	}

	// output the filtered results as the data in the filtered dataset
	_, outputDataFile := getPreFilteringOutputDataFile(outputFolder)
	err = util.CopyFile(filteredData, outputDataFile)
	if err != nil {
		return "", nil, err
	}
	err = HarmonizeDataMetadata(outputFolder)
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

		// Pre-filters rows select by D3MIndex (i.e. row selection.)
		if variable.Key == model.D3MIndexFieldName {
			params = preFilters
		}

		params.Filters = append(params.Filters, f)
	}

	return clone, preFilters
}

func getPreFilteringOutputDataFile(folder string) (bool, string) {
	// make sure the folder exists
	if !util.FileExists(folder) {
		return false, ""
	}

	// read the schema doc (if it exists)
	schemaPath := path.Join(folder, compute.D3MDataSchema)
	if !util.FileExists(schemaPath) {
		return false, ""
	}
	ds, err := serialization.ReadMetadata(schemaPath)
	if err != nil {
		return false, ""
	}

	// get the main data resource path
	return true, model.GetResourcePath(schemaPath, ds.GetMainDataResource())
}

func preparePrefilteringDataset(outputFolder string, sourceDataset *api.Dataset, dataStorage api.DataStorage) ([]*model.Variable, error) {
	// read the data from the database
	data, err := dataStorage.FetchDataset(sourceDataset.ID, sourceDataset.StorageName, true, false, nil)
	if err != nil {
		return nil, err
	}

	// if learning dataset, then update that
	if sourceDataset.LearningDataset != "" {
		// load the dataset from disk
		dsDisk, err := api.LoadDiskDataset(sourceDataset)
		if err != nil {
			return nil, err
		}
		dsDisk = dsDisk.FeaturizedDataset

		// clone the feature dataset
		dsDisk, err = dsDisk.Clone(outputFolder, dsDisk.Dataset.Metadata.ID, dsDisk.Dataset.Metadata.StorageName)
		if err != nil {
			return nil, err
		}

		// update it
		err = dsDisk.UpdateOnDisk(sourceDataset, data)
		if err != nil {
			return nil, err
		}

		// get the variable list
		meta, err := serialization.ReadMetadata(path.Join(outputFolder, compute.D3MDataSchema))
		if err != nil {
			return nil, err
		}
		return meta.GetMainDataResource().Variables, nil
	}

	// read the metadata from disk to keep the reference data resources
	// TODO: MAY WANT TO MAKE THIS A FUNCTION SOMEWHERE!
	metadataSchemaPath := path.Join(env.ResolvePath(sourceDataset.Source, sourceDataset.Folder), compute.D3MDataSchema)
	metadataSource, err := serialization.ReadMetadata(metadataSchemaPath)
	if err != nil {
		return nil, err
	}
	metadataSourceDR := metadataSource.GetMainDataResource()
	metaVarMap := api.MapVariables(metadataSourceDR.Variables, func(variable *model.Variable) string { return variable.Key })
	sourceVarMap := api.MapVariables(sourceDataset.Variables, func(variable *model.Variable) string { return variable.Key })

	// update the metadata to match the data pulled from the data storage
	// (mostly matching column index and dropping columns not pulled)
	variablesData := make([]*model.Variable, len(data[0]))
	sourceVariables := []*model.Variable{}
	for i, f := range data[0] {
		metaVar := metaVarMap[f]
		if metaVar == nil {
			// variable does not exist in the disk metadata yet so add it
			metaVar = sourceVarMap[f]
			metadataSourceDR.Variables = append(metadataSourceDR.Variables, metaVar)
		}
		metaVar.Index = i
		variablesData[i] = metaVar
		sourceVariables = append(sourceVariables, sourceVarMap[f])
	}

	// update all non main data resources to be absolute
	for _, dr := range metadataSource.DataResources {
		if dr != metadataSourceDR {
			dr.ResPath = model.GetResourcePath(metadataSchemaPath, dr)
		}
	}

	// write it out as a dataset
	dsRaw := &serialization.RawDataset{
		ID:       sourceDataset.ID,
		Name:     sourceDataset.Name,
		Data:     data,
		Metadata: metadataSource,
	}
	err = serialization.WriteDataset(outputFolder, dsRaw)
	if err != nil {
		return nil, err
	}

	return sourceVariables, nil
}

// UpdatePrefeaturizedDataset updates a featurized dataset that already exists
// on disk to have new variables included
func UpdatePrefeaturizedDataset(outputFolder string, prefeaturizedPath string, sourceDataset *api.Dataset, storedData [][]string, updateMetadata bool) ([]*model.Variable, error) {
	// copy the prefeaturized dataset to the output folder
	err := util.Copy(prefeaturizedPath, outputFolder)
	if err != nil {
		return nil, err
	}

	// load the dataset from disk
	schemaPath := path.Join(outputFolder, compute.D3MDataSchema)
	dsDisk, err := serialization.ReadDataset(schemaPath)
	if err != nil {
		return nil, err
	}
	metaDiskMainDR := dsDisk.Metadata.GetMainDataResource()

	// determine if there are new columns that were not part of the original dataset
	metaDiskVarMap := api.MapVariables(metaDiskMainDR.Variables, func(variable *model.Variable) string { return variable.Key })

	// get the index of the new fields in the extracted data
	storedVarMap := api.MapVariables(sourceDataset.Variables, func(variable *model.Variable) string { return variable.Key })
	storedDataD3MIndex := -1
	newVars := []*model.Variable{}
	for i, v := range storedData[0] {
		if v == model.D3MIndexFieldName {
			storedDataD3MIndex = i
		} else if storedVarMap[v] != nil && metaDiskVarMap[v] == nil {
			storedVarMap[v].Index = i
			newVars = append(newVars, storedVarMap[v])
		}
	}

	// add the missing columns row by row and only retain rows in the new dataset
	// first build up the new variables by d3m index map
	// then cycle through the featurized rows and append the variables
	newDataMap := map[string][]string{}
	for _, r := range storedData[1:] {
		newVarsData := []string{}
		for i := 0; i < len(newVars); i++ {
			newKey := newVars[i].Key
			newIndex := storedVarMap[newKey].Index
			newVarsData = append(newVarsData, r[newIndex])
		}
		newDataMap[r[storedDataD3MIndex]] = newVarsData
	}

	// add the new fields to the metadata to generate the proper header
	for i := 0; i < len(newVars); i++ {
		newVar := newVars[i]
		newVar.Index = len(metaDiskMainDR.Variables)
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

	// make sure the ids and names match
	if updateMetadata {
		dsDisk.Metadata.ID = sourceDataset.ID
		dsDisk.Metadata.Name = sourceDataset.Name
		dsDisk.Metadata.StorageName = sourceDataset.StorageName
	}

	// output the new pre featurized data
	dsDisk.Data = preFeaturizedOutput
	err = serialization.WriteDataset(outputFolder, dsDisk)
	if err != nil {
		return nil, err
	}

	// capture the final set of variables to use
	storedVarMap = api.MapVariables(sourceDataset.Variables, func(variable *model.Variable) string { return variable.HeaderName })
	metaDiskVarMap = api.MapVariables(metaDiskMainDR.Variables, func(variable *model.Variable) string { return variable.HeaderName })
	outputVariables := make([]*model.Variable, len(dsDisk.Data[0]))
	for i, v := range dsDisk.Data[0] {
		var variable *model.Variable
		if storedVarMap[v] != nil {
			variable = storedVarMap[v]
		} else {
			variable = metaDiskVarMap[v]
		}
		variable.Index = i
		outputVariables[i] = variable
	}

	return outputVariables, nil
}

// HarmonizeDataMetadata updates a dataset on disk to have the schema info
// match the header of the backing data file, as well as limit
// variables to valid auto ml fields.
func HarmonizeDataMetadata(datasetFolder string) error {
	// load the dataset
	schemaPath := path.Join(datasetFolder, compute.D3MDataSchema)
	ds, err := serialization.ReadDataset(schemaPath)
	if err != nil {
		return err
	}

	// assume metadata has the correct info, but with superflous metadata variables
	// drop metadata variables
	mainDR := ds.Metadata.GetMainDataResource()
	finalVariables := []*model.Variable{}
	for _, v := range mainDR.Variables {
		if model.IsTA2Field(v.DistilRole, v.SelectedRole) || v.DistilRole == model.VarDistilRoleSystemData {
			v.Index = len(finalVariables)
			finalVariables = append(finalVariables, v)
		}
	}
	mainDR.Variables = finalVariables

	// set the header to match what is in the metadata
	ds.Data[0] = mainDR.GenerateHeader()

	// output the dataset
	err = serialization.WriteDataset(datasetFolder, ds)
	if err != nil {
		return err
	}

	return nil
}
