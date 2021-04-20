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

func filterData(client *compute.Client, ds *api.Dataset, filterParams *api.FilterParamsRaw, dataStorage api.DataStorage) (string, *api.FilterParamsRaw, error) {
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
	pipeline, err := description.CreateDataFilterPipeline("Pre Filtering", "pre filter a dataset that has metadata features", resultingVariables, preFilters.Filters.List)
	if err != nil {
		return "", nil, err
	}

	// allowable types are prioritized in order
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

func mapFilterKeys(dataset string, filters *api.FilterParamsRaw, variables []*model.Variable) *api.FilterParamsRaw {
	filtersUpdated := filters.Clone()

	varsMapped := api.MapVariables(variables, func(variable *model.Variable) string { return variable.Key })
	for _, f := range filtersUpdated.Filters.List {
		variable := varsMapped[f.Key]
		if variable.Key == f.Key && variable.IsGrouping() {
			if _, ok := variable.Grouping.(*model.GeoBoundsGrouping); ok {
				grouping := variable.Grouping.(*model.GeoBoundsGrouping)
				f.Key = grouping.CoordinatesCol
			}
		}
	}

	return filtersUpdated
}

func hashFilter(schemaFile string, filterParams *api.FilterParamsRaw) (uint64, error) {
	// generate the hash from the params
	hashStruct := struct {
		Schema       string
		FilterParams *api.FilterParamsRaw
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

func getPreFiltering(ds *api.Dataset, filterParams *api.FilterParamsRaw) (*api.FilterParamsRaw, *api.FilterParamsRaw) {
	vars := map[string]*model.Variable{}
	for _, v := range ds.Variables {
		vars[v.Key] = v
	}
	clone := filterParams.Clone()
	// filter if a clustering or outlier detection metadata feature exist
	// remove pre filters from the rest of the filters since they should not be in the main pipeline
	// TODO: NEED TO HANDLE OUTLIER FILTERS!
	preFilters := &api.FilterParamsRaw{
		Filters: api.FilterObject{List: []*model.Filter{}, Invert: false},
	}
	filters := clone.Filters
	clone.Filters = api.FilterObject{List: []*model.Filter{}, Invert: false}
	for _, f := range filters.List {
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

		params.Filters.List = append(params.Filters.List, f)
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
	data, err := dataStorage.FetchDataset(sourceDataset.ID, sourceDataset.StorageName, true, nil)
	if err != nil {
		return nil, err
	}

	// load the dataset from disk
	dsDisk, err := api.LoadDiskDataset(sourceDataset)
	if err != nil {
		return nil, err
	}

	// if learning dataset, then update that
	if sourceDataset.LearningDataset != "" {
		dsDisk = dsDisk.FeaturizedDataset
	}

	// clone the feature dataset
	dsDisk, err = dsDisk.Clone(outputFolder, dsDisk.Dataset.Metadata.ID, dsDisk.Dataset.Metadata.StorageName)
	if err != nil {
		return nil, err
	}

	// update it
	err = dsDisk.UpdateOnDisk(sourceDataset, data, false)
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
