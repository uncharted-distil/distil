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

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	"github.com/uncharted-distil/distil/api/util/json"
)

// FeaturizeDataset creates a featurized output of the data that can be used
// in simplified pipelines.
func FeaturizeDataset(originalSchemaFile string, schemaFile string, dataset string, metaStorage api.MetadataStorage, config *IngestTaskConfig) (string, string, error) {
	// load the metadata from the metadata storage
	ds, err := metaStorage.FetchDataset(dataset, true, true)
	if err != nil {
		return "", "", err
	}

	// create & submit the featurize pipeline
	pip, err := description.CreateMultiBandImageFeaturizationPipeline("Euler", "", ds.Variables)
	if err != nil {
		return "", "", err
	}

	// pipeline execution assumes datasetDoc.json as schema file
	datasetURI, err := submitPipeline([]string{originalSchemaFile}, pip)
	if err != nil {
		return "", "", err
	}

	// create the dataset folder
	featurizedDatasetID := fmt.Sprintf("%s-featurized", dataset)
	featurizedDatasetID, err = getUniqueOutputFolder(featurizedDatasetID, env.GetAugmentedPath())
	if err != nil {
		return "", "", err
	}
	featurizedOutputPath := path.Join(env.GetAugmentedPath(), featurizedDatasetID)

	// copy the output to the folder as the data
	dataOutputPath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)
	err = util.Copy(datasetURI, path.Join(featurizedOutputPath, dataOutputPath))
	if err != nil {
		return "", "", err
	}

	// read the header to get all the featurized fields
	header, err := util.ReadCSVHeader(datasetURI)
	if err != nil {
		return "", "", err
	}

	// load the metadata from the source schema file
	meta, err := metadata.LoadMetadataFromClassification(schemaFile, path.Join(path.Dir(schemaFile), config.ClassificationOutputPathRelative), false, true)
	if err != nil {
		return "", "", err
	}
	mainDR := meta.GetMainDataResource()

	// update the metadata to have all the new fields as floats
	schemaOutputPath := path.Join(featurizedOutputPath, compute.D3MDataSchema)
	for i := len(mainDR.Variables); i < len(header); i++ {
		mainDR.Variables = append(mainDR.Variables, model.NewVariable(i, header[i], header[i],
			header[i], model.RealType, model.RealType, "featurized value",
			[]string{model.RoleAttribute}, model.VarDistilRoleData, nil, mainDR.Variables, false))
	}
	err = metadata.WriteSchema(meta, schemaOutputPath, false)
	if err != nil {
		return "", "", err
	}

	return featurizedDatasetID, featurizedOutputPath, nil
}

// SetGroups updates the dataset metadata (as stored) to capture group information.
func SetGroups(datasetID string, rawGrouping map[string]interface{}, meta api.MetadataStorage, config *IngestTaskConfig) error {
	ds, err := meta.FetchDataset(datasetID, true, true)
	if err != nil {
		return err
	}
	if isRemoteSensingDataset(ds) {
		rsg := &model.RemoteSensingGrouping{}
		err = json.MapToStruct(rsg, rawGrouping)
		if err != nil {
			return err
		}

		err = meta.AddGroupedVariable(datasetID, rsg.IDCol+"_group", "Tile", model.RemoteSensingType, model.VarDistilRoleGrouping, rsg)
		if err != nil {
			return err
		}
	}

	return nil
}

func isRemoteSensingDataset(ds *api.Dataset) bool {
	for _, v := range ds.Variables {
		if model.IsMultiBandImage(v.Type) {
			return true
		}
	}

	return false
}
