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

	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
	"github.com/uncharted-distil/distil/api/util/json"
)

func submitForBatch(pip *description.FullySpecifiedPipeline) func(string) (string, error) {
	return func(schemaFile string) (string, error) {
		return submitPipeline([]string{schemaFile}, pip, true)
	}
}

// CreateFeaturizedDatasetID creates a dataset id for a learning dataset.
func CreateFeaturizedDatasetID(datasetID string) string {
	return fmt.Sprintf("%s-featurized", datasetID)
}

// FeaturizeDataset creates a featurized output of the data that can be used
// in simplified pipelines.
func FeaturizeDataset(originalSchemaFile string, schemaFile string, dataset string, metaStorage api.MetadataStorage, config *IngestTaskConfig) (string, string, error) {
	envConfig, err := env.LoadConfig()
	if err != nil {
		return "", "", err
	}

	// load the metadata from the metadata storage
	ds, err := metaStorage.FetchDataset(dataset, true, true, false)
	if err != nil {
		return "", "", err
	}

	// create & submit the featurize pipeline
	// determine if remote sensing or image
	var pip *description.FullySpecifiedPipeline
	imageDataset := false
	for _, v := range ds.Variables {
		if model.IsImage(v.Type) {
			imageDataset = true
			break
		}
	}
	if imageDataset {
		pip, err = description.CreateImageFeaturizationPipeline("Image featurization", "", ds.Variables)
	} else {
		pip, err = description.CreateMultiBandImageFeaturizationPipeline("Multiband image featurization", "", ds.Variables,
			envConfig.RemoteSensingNumJobs, envConfig.RemoteSensingGPUBatchSize, envConfig.PoolFeatures)
	}
	if err != nil {
		return "", "", err
	}

	// pipeline execution assumes datasetDoc.json as schema file
	datasetURI, err := batchSubmitDataset(schemaFile, dataset, config.DatasetBatchSize, submitForBatch(pip))
	if err != nil {
		return "", "", err
	}
	featurizedDataReader := serialization.GetStorage(datasetURI)
	featurizedData, err := featurizedDataReader.ReadData(datasetURI)
	if err != nil {
		return "", "", err
	}

	// create the dataset folder
	featurizedDatasetID := CreateFeaturizedDatasetID(dataset)
	featurizedDatasetID, err = GetUniqueOutputFolder(featurizedDatasetID, env.GetAugmentedPath())
	if err != nil {
		return "", "", err
	}
	featurizedOutputPath := path.Join(env.GetAugmentedPath(), featurizedDatasetID)

	// copy the output to the folder as the data
	dataOutputPath := path.Join(featurizedOutputPath, path.Join(compute.D3MDataFolder, compute.DistilParquetLearningData))
	featurizedDataWriter := serialization.GetStorage(dataOutputPath)
	err = featurizedDataWriter.WriteData(dataOutputPath, featurizedData)
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

	// keep only the fields in the output (including the new fields as floats)
	schemaOutputPath := path.Join(featurizedOutputPath, compute.D3MDataSchema)
	vars := []*model.Variable{}
	metadataVariables := map[string]*model.Variable{}
	for _, v := range mainDR.Variables {
		metadataVariables[v.HeaderName] = v
	}
	for index, field := range header {
		var v *model.Variable
		if metadataVariables[field] != nil {
			v = metadataVariables[field]
			v.Index = index
		} else {
			v = model.NewVariable(index, field, field, field, field, model.RealType,
				model.RealType, "featurized value", []string{model.RoleAttribute},
				model.VarDistilRoleSystemData, nil, mainDR.Variables, false)
		}
		vars = append(vars, v)
	}
	mainDR.Variables = vars
	mainDR.ResPath = dataOutputPath

	err = featurizedDataWriter.WriteMetadata(schemaOutputPath, meta, true, true)
	if err != nil {
		return "", "", err
	}

	return featurizedDatasetID, featurizedOutputPath, nil
}

// SetGroups updates the dataset metadata (as stored) to capture group information.
func SetGroups(datasetID string, rawGroupings []map[string]interface{}, data api.DataStorage, meta api.MetadataStorage, config *IngestTaskConfig) error {
	// We currently only allow for one image to be present in a dataset.
	multiBandImageGroupings := getMultiBandImageGrouping(rawGroupings)
	numGroupings := len(multiBandImageGroupings)
	if numGroupings >= 1 {
		log.Warnf("found %d multiband image groupings - only first will be used", numGroupings)
	}
	if numGroupings > 0 {
		rsg := &model.MultiBandImageGrouping{}
		err := json.MapToStruct(rsg, multiBandImageGroupings[0])
		if err != nil {
			return err
		}
		// Set the name of the expected cluster column - it doesn't necessarily exist.
		varName := rsg.IDCol + "_group"
		rsg.ClusterCol = model.ClusterVarPrefix + rsg.IDCol
		err = meta.AddGroupedVariable(datasetID, varName, "Tile", model.MultiBandImageType, model.VarDistilRoleGrouping, rsg)
		if err != nil {
			return err
		}
	}

	geoBoundsGroupings := getGeoBoundsGrouping(rawGroupings)
	for _, rawGrouping := range geoBoundsGroupings {
		grouping := &model.GeoBoundsGrouping{}
		err := json.MapToStruct(grouping, rawGrouping)
		if err != nil {
			return err
		}

		ds, err := meta.FetchDataset(datasetID, true, true, true)
		if err != nil {
			return err
		}

		// confirm the existence of the underlying polygon field, creating it if necessary
		// (less than ideal because it hides a pretty big side effect)
		// (other option would be to error here and let calling code worry about it)
		exists, err := data.DoesVariableExist(datasetID, ds.StorageName, grouping.PolygonCol)
		if err != nil {
			return err
		}
		if !exists {
			err = createGeoboundsField(datasetID, ds.StorageName, grouping.CoordinatesCol, grouping.PolygonCol, data, meta)
			if err != nil {
				return err
			}
		}

		// Set the name of the expected cluster column
		varName := grouping.CoordinatesCol + "_group"
		err = meta.AddGroupedVariable(datasetID, varName, "coordinates", model.GeoBoundsType, model.VarDistilRoleGrouping, grouping)
		if err != nil {
			return err
		}
	}

	return nil
}

func getMultiBandImageGrouping(rawGroupings []map[string]interface{}) []map[string]interface{} {
	results := []map[string]interface{}{}
	for _, rawGrouping := range rawGroupings {
		if rawGrouping["type"] != nil && rawGrouping["type"].(string) == model.MultiBandImageType {
			results = append(results, rawGrouping)
		}
	}
	return results
}

func getGeoBoundsGrouping(rawGroupings []map[string]interface{}) []map[string]interface{} {
	results := []map[string]interface{}{}
	for _, rawGrouping := range rawGroupings {
		if rawGrouping["type"] != nil && rawGrouping["type"].(string) == model.GeoBoundsType {
			results = append(results, rawGrouping)
		}
	}
	return results
}

func canFeaturize(datasetID string, meta api.MetadataStorage) bool {
	ds, err := meta.FetchDataset(datasetID, true, true, false)
	if err != nil {
		log.Warnf("error fetching dataset to determine if it can be featurized: %+v", err)
		return false
	}

	for _, v := range ds.Variables {
		if model.IsMultiBandImage(v.Type) || model.IsImage(v.Type) {
			return true
		}
	}

	return false
}

func createGeoboundsField(datasetID string, storageName string, coordinateField string,
	geometryField string, data api.DataStorage, meta api.MetadataStorage) error {
	// pull the coordinate data from the database
	params := &api.FilterParams{Variables: []string{coordinateField}}
	coordinateData, err := data.FetchData(datasetID, storageName, params, false, nil)
	if err != nil {
		return err
	}

	// add the field
	err = data.AddVariable(datasetID, storageName, geometryField, model.GeoBoundsType, "")
	if err != nil {
		return err
	}

	// extract update data from the data fetched
	// coordinate data should be d3m index & coordinate field
	d3mIndexIndex := coordinateData.Columns[model.D3MIndexFieldName].Index
	coordinateFieldIndex := (d3mIndexIndex + 1) % 2
	updates := map[string]string{}
	for _, row := range coordinateData.Values {
		d3mIndexString := fmt.Sprintf("%.0f", row[d3mIndexIndex].Value.(float64))
		updates[d3mIndexString] = util.CreatePolygonFromCoordinates(row[coordinateFieldIndex].Value.([]float64))
	}

	// update the field data
	err = data.UpdateVariableBatch(storageName, geometryField, updates)
	if err != nil {
		return err
	}

	// add the field to the metadata
	err = meta.AddVariable(datasetID, geometryField, coordinateField, model.GeoBoundsType, model.VarDistilRoleMetadata)
	if err != nil {
		return err
	}

	return nil
}
