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

	// read the data from the database
	data, err := dataStorage.FetchDataset(ds.ID, ds.StorageName, true, false, nil)
	if err != nil {
		return "", nil, err
	}

	// update the metadata to match the data pulled from the data storage
	// (mostly matching column index and dropping columns not pulled)
	metaVarMap := map[string]*model.Variable{}
	for _, v := range ds.Variables {
		metaVarMap[v.Key] = v
	}
	variablesData := make([]*model.Variable, len(data[0]))
	for i, f := range data[0] {
		variablesData[i] = metaVarMap[f]
		variablesData[i].Index = i
	}
	ds.Variables = variablesData

	// write it out as a dataset
	dsRaw := &api.RawDataset{
		ID:       ds.ID,
		Name:     ds.Name,
		Data:     data,
		Metadata: ds.ToMetadata(),
	}
	err = serialization.WriteDataset(outputFolder, dsRaw)
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
