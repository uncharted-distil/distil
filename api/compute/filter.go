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
	"strings"

	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
)

func filterData(client *compute.Client, ds *api.Dataset, filterParams *api.FilterParams, dataStorage api.DataStorage) (string, error) {
	inputPath := env.ResolvePath(ds.Source, ds.Folder)

	// determine if filtering is needed
	preFilters := getPreFiltering(ds, filterParams)
	if preFilters.Empty(false) {
		return inputPath, nil
	}

	// check if the filtered results already exists
	hash, err := hashFilter(inputPath, preFilters)
	if preFilters.Empty(false) {
		return "", err
	}
	outputFolder := env.ResolvePath("tmp", fmt.Sprintf("%s-%v", ds.Folder, hash))
	outputDataFile := path.Join(outputFolder, compute.D3MDataFolder, compute.D3MLearningData)
	if util.FileExists(outputDataFile) {
		return outputFolder, nil
	}

	// read the data from the database
	data, err := dataStorage.FetchDataset(ds.ID, ds.StorageName, false, nil)
	if err != nil {
		return "", err
	}

	// write it out as a dataset
	dsRaw := &api.RawDataset{
		ID:       ds.ID,
		Name:     ds.Name,
		Data:     data,
		Metadata: ds.ToMetadata(),
	}
	err = serialization.WriteDataset(outputFolder, dsRaw)
	if err != nil {
		return "", err
	}

	// run the filtering pipeline
	pipeline, err := description.CreateDataFilterPipeline("Pre Filtering", "pre filter a dataset that has metadata features", ds.Variables, preFilters.Filters)
	if err != nil {
		return "", err
	}
	filteredData, err := SubmitPipeline(client, []string{outputFolder}, nil, nil, pipeline, true)
	if err != nil {
		return "", err
	}

	// output the filtered results as the data in the filtered dataset
	err = util.CopyFile(filteredData, outputDataFile)
	if err != nil {
		return "", err
	}

	return outputFolder, nil
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

func getPreFiltering(ds *api.Dataset, filterParams *api.FilterParams) *api.FilterParams {
	vars := map[string]*model.Variable{}
	for _, v := range ds.Variables {
		vars[v.Key] = v
	}
	// filter if a clustering or outlier detection metadata feature exist
	preFilters := &api.FilterParams{
		Filters: []*model.Filter{},
	}
	for _, f := range filterParams.Filters {
		if vars[f.Key].DistilRole == model.VarDistilRoleMetadata &&
			(strings.HasPrefix(f.Key, "_cluster") || strings.HasPrefix(f.Key, "_outlier")) {
			preFilters.Filters = append(preFilters.Filters, f)
		}
	}

	return preFilters
}
