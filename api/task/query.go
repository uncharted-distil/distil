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
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	"path"
)

const (
	// image retrieval primitive has hardcoded field name
	queryFieldName = "annotations"
	score          = "score"
)

// QueryParams helper struct to simplify query task calling.
type QueryParams struct {
	Dataset     string
	TargetName  string
	DataStorage api.DataStorage
	MetaStorage api.MetadataStorage
	Filters     *api.FilterParams
}

// Query uses a query pipeline to rank data by nearness to a target.
func Query(params QueryParams) (map[string]interface{}, error) {
	// get the dataset metadata
	ds, err := params.MetaStorage.FetchDataset(params.Dataset, true, true, false)
	if err != nil {
		return nil, err
	}

	// only prefeaturized datasets can be queried
	if ds.LearningDataset == "" {
		return nil, errors.Errorf("only prefeaturized datasets support querying")
	}

	// extract the dataset from the database
	data, err := params.DataStorage.FetchDataset(params.Dataset, ds.StorageName, false, params.Filters)
	if err != nil {
		return nil, err
	}

	// keep only the d3m index and the target column (1 row / index)
	dataToStore := extractQueryDataset(params.TargetName, data)

	// store it to disk
	datasetPath, err := writeQueryDataset(ds, dataToStore)
	if err != nil {
		return nil, err
	}

	// create the image retrieval pipeline
	desc, err := description.CreateImageQueryPipeline("image query", "pipeline to retrieve pertinent images")
	if err != nil {
		return nil, err
	}

	// submit the pipeline with no cache
	resultURI, err := submitPipeline([]string{ds.LearningDataset, datasetPath}, desc, false)
	if err != nil {
		return nil, err
	}
	storageResult := serialization.GetStorage(resultURI)
	resultData, err := storageResult.ReadData(resultURI)
	if err != nil {
		return nil, err
	}
	// update the database to have the results
	// the results are the score for the search, between 0 and 1
	// it is stored in a separate column from the label itself
	err = persistQueryResults(params, ds.StorageName, resultData)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// getColumnIndices returns: target, d3mIndex
func getColumnIndices(targetName string, data [][]string) (int, int) {
	targetIndex := -1
	d3mIndex := -1
	for i, c := range data[0] {
		if c == targetName {
			targetIndex = i
		} else if c == model.D3MIndexFieldName {
			d3mIndex = i
		}
	}
	return targetIndex, d3mIndex
}
func extractQueryDataset(targetName string, data [][]string) [][]string {
	// get the needed column indices
	targetIndex, d3mIndex := getColumnIndices(targetName, data)

	// need to reduce to 1 row / d3m index (labels should match across the whole group)
	valueMap := map[string]int{"unlabeled": -1, "negative": 0, "positive": 1}
	reducedData := map[string]string{}
	dataToStore := [][]string{{model.D3MIndexFieldName, queryFieldName}}
	for i := 1; i < len(data); i++ {
		key := data[i][d3mIndex]
		_, ok := reducedData[key]
		if !ok {
			label := data[i][targetIndex]
			reducedData[key] = label
			dataToStore = append(dataToStore, []string{key, fmt.Sprintf("%d", valueMap[label])})
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
		model.NewVariable(0, model.D3MIndexFieldName, model.D3MIndexFieldName, model.D3MIndexFieldName,
			model.D3MIndexFieldName, model.IntegerType, model.IntegerType, "D3M index",
			[]string{model.RoleMultiIndex}, model.VarDistilRoleIndex, nil, dr.Variables, false),
		model.NewVariable(1, queryFieldName, queryFieldName, queryFieldName, queryFieldName, model.StringType,
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
		err = params.DataStorage.AddVariable(params.Dataset, storageName, targetScore, model.RealType, "0")
		if err != nil {
			return err
		}
		err = params.MetaStorage.AddVariable(params.Dataset, targetScore, targetScore, model.RealType, model.VarDistilRoleSystemData)
		if err != nil {
			return err
		}
	} else {
		err = params.DataStorage.SetVariableValue(storageName, targetScore, "0")
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
	err = params.DataStorage.UpdateVariableBatch(storageName, targetScore, updates)
	if err != nil {
		return err
	}

	return nil
}
