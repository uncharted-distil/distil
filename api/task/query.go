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

package task

import (
	"fmt"
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
)

type QueryParams struct {
	Dataset     string
	TargetName  string
	DataStorage api.DataStorage
	MetaStorage api.MetadataStorage
	Filters     *api.FilterParams
}

// Query uses a query pipeline to rank data by nearness to a target.
func Query(params QueryParams) (string, error) {
	// get the dataset metadata
	ds, err := params.MetaStorage.FetchDataset(params.Dataset, true, true)
	if err != nil {
		return "", err
	}

	// only prefeaturized datasets can be queried
	if ds.LearningDataset == "" {
		return "", errors.Errorf("only prefeaturized datasets support querying")
	}

	// extract the dataset from the database
	data, err := params.DataStorage.FetchDataset(params.Dataset, ds.StorageName, false, params.Filters)
	if err != nil {
		return "", err
	}

	// keep only the d3m index and the target column (1 row / index)
	dataToStore := extractQueryDataset(params.TargetName, data)

	// store it to disk
	datasetPath, err := writeQueryDataset(ds, dataToStore)
	if err != nil {
		return "", err
	}

	// create the image retrieval pipeline
	desc, err := description.CreateImageQueryPipeline("image query", "pipeline to retrieve pertinent images")
	if err != nil {
		return "", err
	}

	// submit the pipeline
	resultURI, err := submitPipeline([]string{params.Dataset, datasetPath}, desc)
	if err != nil {
		return "", err
	}
	storageResult := serialization.GetStorage(resultURI)
	resultData, err := storageResult.ReadData(resultURI)
	if err != nil {
		return "", err
	}

	// update the database to have the results
	// the results are the score for the search, between 0 and 1
	// it is stored in a separate column from the label itself
	err = persistQueryResults(params, ds.StorageName, resultData)
	if err != nil {
		return "", err
	}

	return ds.ID, nil
}

func extractQueryDataset(targetName string, data [][]string) [][]string {
	// get the needed column indices
	targetIndex := -1
	d3mIndex := -1
	for i, c := range data[0] {
		if c == targetName {
			targetIndex = i
		} else if c == model.D3MIndexFieldName {
			d3mIndex = i
		}
	}

	// need to reduce to 1 row / d3m index (labels should match across the whole group)
	reducedData := map[string]string{}
	dataToStore := [][]string{[]string{model.D3MIndexFieldName, targetName}}
	for i := 1; i < len(data); i++ {
		key := data[i][d3mIndex]
		_, ok := reducedData[key]
		if !ok {
			label := data[i][targetIndex]
			reducedData[key] = label
			dataToStore = append(dataToStore, []string{key, label})
		}
	}

	return dataToStore
}

func writeQueryDataset(ds *api.Dataset, data [][]string) (string, error) {
	// path to store to should be consistent to be overwritten every run
	// (although this does not play nice with simultaneous requests)
	datasetIDTarget := fmt.Sprintf("%s-query", ds.ID)
	datasetPathTarget := path.Join(env.GetTmpPath(), datasetIDTarget)
	dataPathTarget := path.Join(datasetPathTarget, compute.D3MDataFolder, compute.D3MLearningData)
	storage := serialization.GetStorage(dataPathTarget)
	err := storage.WriteData(dataPathTarget, data)
	if err != nil {
		return "", err
	}

	// create the metadata for the dataset that contains the target info
	meta := model.NewMetadata(datasetIDTarget, datasetIDTarget, "query dataset", ds.StorageName)
	dr := model.NewDataResource(compute.DefaultResourceID, model.ResTypeTable, map[string][]string{compute.D3MResourceFormat: {"csv"}})
	dr.Variables = []*model.Variable{
		model.NewVariable(0, model.D3MIndexFieldName, model.D3MIndexFieldName,
			model.D3MIndexFieldName, model.IntegerType, model.IntegerType, "D3M index",
			[]string{model.RoleIndex}, model.VarDistilRoleIndex, nil, dr.Variables, false),
		model.NewVariable(2, "label", "label", "label", model.StringType,
			model.StringType, "Label for the query", []string{"suggestedTarget"},
			model.VarDistilRoleData, nil, dr.Variables, false),
	}
	dr.ResPath = dataPathTarget
	meta.DataResources = []*model.DataResource{dr}

	// output the metadata
	metadataPathTarget := path.Join(datasetPathTarget, compute.D3MDataSchema)
	err = storage.WriteMetadata(metadataPathTarget, meta, false, false)
	if err != nil {
		return "", err
	}

	return datasetPathTarget, nil
}

func persistQueryResults(params QueryParams, storageName string, resultData [][]string) error {
	targetScore := fmt.Sprintf("__query_%s", params.TargetName)
	// results should be d3mindex, score
	exists, err := params.DataStorage.DoesVariableExist(params.Dataset, storageName, targetScore)
	if err != nil {
		return err
	}

	if !exists {
		// create the variable to hold the rank
		err = params.DataStorage.AddField(params.Dataset, storageName, targetScore, model.RealType, "0")
		if err != nil {
			return err
		}
	}

	// restructure the results to match expected collection format
	updates := map[string]string{}
	for _, r := range resultData[1:] {
		updates[r[0]] = r[1]
	}

	// overwrite the stored ranking
	err = params.DataStorage.UpdateData(params.Dataset, storageName, targetScore, updates, nil)
	if err != nil {
		return err
	}

	return nil
}
