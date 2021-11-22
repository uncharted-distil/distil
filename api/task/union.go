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

package task

import (
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	apiModel "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
)

// VerticalConcat will bring mastery.
func VerticalConcat(dataStorage apiModel.DataStorage, joinLeft *JoinSpec, joinRight *JoinSpec) (string, *apiModel.FilteredData, error) {
	unionPaths, deletePaths, err := reorderFields(joinLeft.DatasetPath, joinRight.DatasetPath)
	if err != nil {
		return "", nil, err
	}
	for _, d := range deletePaths {
		defer util.Delete(d)
	}

	pipelineDesc, err := description.CreateVerticalConcatPipeline("Unioner", "Combine existing data")
	if err != nil {
		return "", nil, err
	}

	datasetPath, _, err := join(joinLeft, joinRight, pipelineDesc, unionPaths, defaultSubmitter{}, true)
	if err != nil {
		return "", nil, err
	}

	// rewrite dataset to have unique d3m index
	data, err := rewriteD3MIndex(datasetPath)
	if err != nil {
		return "", nil, err
	}

	return datasetPath, data, nil
}

func reorderFields(dsAPath string, dsBPath string) ([]string, []string, error) {
	log.Infof("reordering fields to have datasets found at '%s' and '%s' match", dsAPath, dsBPath)
	dsA, err := apiModel.LoadDiskDatasetFromFolder(dsAPath)
	if err != nil {
		return nil, nil, err
	}

	dsB, err := apiModel.LoadDiskDatasetFromFolder(dsBPath)
	if err != nil {
		return nil, nil, err
	}

	clonePath := ""
	unionPaths := []string{}
	if len(dsB.Dataset.Data) > len(dsA.Dataset.Data) {
		clonePath, err = reorderDatasetFields(dsA, dsB)
		unionPaths = append(unionPaths, dsBPath)
	} else {
		clonePath, err = reorderDatasetFields(dsB, dsA)
		unionPaths = append(unionPaths, dsAPath)
	}
	if err != nil {
		return nil, nil, err
	}
	unionPaths = append(unionPaths, clonePath)

	return unionPaths, []string{clonePath}, nil
}

func reorderDatasetFields(dsToReorder *apiModel.DiskDataset, dsOrder *apiModel.DiskDataset) (string, error) {
	pathClone := path.Join(env.GetTmpPath(), fmt.Sprintf("%s-reorder", dsToReorder.Dataset.ID))
	dsA, err := dsToReorder.Clone(pathClone, dsToReorder.Dataset.ID, dsToReorder.Dataset.ID)
	if err != nil {
		return "", err
	}
	err = dsA.ReorderFields(dsOrder.Dataset.Metadata.GetMainDataResource().Variables)
	if err != nil {
		return "", err
	}
	err = dsA.SaveDataset()
	if err != nil {
		return "", err
	}

	return pathClone, nil
}

func rewriteD3MIndex(datasetPath string) (*apiModel.FilteredData, error) {
	// read the raw dataset
	ds, err := serialization.ReadDataset(path.Join(datasetPath, compute.D3MDataSchema))
	if err != nil {
		return nil, err
	}

	// if d3m index is a multi index, need to use the grouping variable to reindex
	// NOTE: THIS CURRENTLY ASSUMES A SINGLE GROUPING VARIABLE IS USED TO DEFINE A GROUP!
	indexingVariable := ds.GetVariableMetadata(model.D3MIndexFieldName)
	if indexingVariable == nil {
		return nil, errors.Errorf("no d3m index field in dataset")
	}
	isMulti := false
	for _, r := range indexingVariable.Role {
		if r == model.RoleMultiIndex {
			isMulti = true
			break
		}
	}
	indexingVariableIndices := []int{}
	if isMulti {
		for _, v := range ds.Metadata.GetMainDataResource().Variables {
			if v.HasRole(model.VarDistilRoleGrouping) || v.HasRole(model.VarDistilRoleGroupingSupplemental) {
				indexingVariableIndices = append(indexingVariableIndices, v.Index)
			}
		}
	} else {
		indexingVariableIndices = append(indexingVariableIndices, indexingVariable.Index)
	}

	// find the d3m index field
	d3mIndexIndex := ds.GetVariableIndex(model.D3MIndexFieldName)

	// rewrite the index to make all rows unique (skipping header)
	count := 1
	reindexedValues := map[string]string{}
	for _, r := range ds.Data[1:] {
		keyValue := getKeyValue(r, indexingVariableIndices)
		if reindexedValues[keyValue] != "" {
			r[d3mIndexIndex] = reindexedValues[keyValue]
		} else {
			indexValue := fmt.Sprintf("%d", count)
			count++
			r[d3mIndexIndex] = indexValue
			reindexedValues[keyValue] = indexValue
		}
	}

	// save the updated dataset
	err = serialization.WriteDataset(datasetPath, ds)
	if err != nil {
		return nil, err
	}

	dataParsed, err := apiModel.CreateFilteredData(ds.Data, ds.Metadata.GetMainDataResource().Variables, false, 100)
	if err != nil {
		return nil, err
	}

	return dataParsed, nil
}

func getKeyValue(row []string, groupingIndices []int) string {
	keyValues := make([]string, len(row))
	for i, v := range groupingIndices {
		keyValues[i] = row[v]
	}

	return strings.Join(keyValues, "|")
}
