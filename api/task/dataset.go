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
	"io/ioutil"
	"path"

	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	log "github.com/unchartedsoftware/plog"

	apicompute "github.com/uncharted-distil/distil/api/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
)

// DatasetConstructor is used to build a dataset.
type DatasetConstructor interface {
	CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*serialization.RawDataset, error)
	GetDefinitiveTypes() []*model.Variable
	CleanupTempFiles()
}

// CreateDataset structures a raw csv file into a valid D3M dataset.
func CreateDataset(dataset string, datasetCtor DatasetConstructor, outputPath string, config *env.Config) (string, string, error) {
	ingestConfig := NewConfig(*config)

	// save the csv file in the file system datasets folder
	if !config.IngestOverwrite {
		datasetUnique, err := GetUniqueOutputFolder(dataset, outputPath)
		if err != nil {
			return "", "", err
		}
		if datasetUnique != dataset {
			log.Infof("dataset changed to '%s' from '%s'", datasetUnique, dataset)
			dataset = datasetUnique
		}
	}
	outputDatasetPath := path.Join(outputPath, dataset)
	dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)
	dataPath := path.Join(outputDatasetPath, dataFilePath)

	log.Infof("running dataset creation for dataset '%s', writing output to '%s'", dataset, outputDatasetPath)
	ds, err := datasetCtor.CreateDataset(outputDatasetPath, dataset, config)
	if err != nil {
		return "", "", err
	}
	defer datasetCtor.CleanupTempFiles()

	datasetStorage := serialization.GetStorage(dataPath)
	err = datasetStorage.WriteData(dataPath, ds.Data)
	if err != nil {
		return "", "", err
	}

	schemaPath := path.Join(outputDatasetPath, compute.D3MDataSchema)
	err = datasetStorage.WriteMetadata(schemaPath, ds.Metadata, true, false)
	if err != nil {
		return "", "", err
	}

	// format the dataset into a D3M format
	formattedPath, err := Format(schemaPath, dataset, ingestConfig)
	if err != nil {
		return "", "", err
	}

	// if definitive types provided, write out the classification information
	if ds.DefinitiveTypes {
		outputPath := path.Join(formattedPath, config.ClassificationOutputPath)
		log.Infof("write definitve types to '%s'", outputPath)
		classification := buildClassificationFromMetadata(ds.Metadata.GetMainDataResource().Variables)
		classification.Path = outputPath
		err := metadata.WriteClassification(classification, outputPath)
		if err != nil {
			return "", "", err
		}
	}

	// copy to the original output location for consistency
	if formattedPath != outputDatasetPath {
		// copy the data file and the metadata doc
		err = util.Copy(formattedPath, path.Dir(schemaPath))
		if err != nil {
			return "", "", err
		}
	}

	return dataset, formattedPath, nil
}

// CopyDiskDataset copies an existing dataset on disk to a new location,
// updating the ID and the storage name.
func CopyDiskDataset(existingURI string, newURI string, newDatasetID string, newStorageName string) (*api.DiskDataset, error) {
	dsDisk, err := api.LoadDiskDatasetFromFolder(existingURI)
	if err != nil {
		return nil, err
	}

	dsDisk, err = dsDisk.Clone(newURI, newDatasetID, newStorageName)
	if err != nil {
		return nil, err
	}

	return dsDisk, nil
}

// ExportDataset extracts a dataset from the database and metadata storage, writing
// it to disk in D3M dataset format.
func ExportDataset(dataset string, metaStorage api.MetadataStorage, dataStorage api.DataStorage, filterParams *api.FilterParams) (string, string, error) {
	// TODO: most likely need to either get a unique folder name for output or error if already exists
	return exportDiskDataset(dataset, dataset, env.ResolvePath(metadata.Augmented, dataset), metaStorage, dataStorage, false, filterParams)
}

// CreateDatasetFromResult creates a new dataset based on a result set & the input
// to the model
func CreateDatasetFromResult(newDatasetName string, predictionDataset string, sourceDataset string, features []string,
	targetName string, resultURI string, metaStorage api.MetadataStorage, dataStorage api.DataStorage, config env.Config) (string, error) {
	// get the prediction dataset
	predictionDS, err := metaStorage.FetchDataset(predictionDataset, true, true, true)
	if err != nil {
		return "", err
	}
	sourceDS, err := metaStorage.FetchDataset(sourceDataset, true, true, true)
	if err != nil {
		return "", err
	}

	// need to expand feature set to handle groups
	varsExpanded := map[string]bool{}
	varsSource := map[string]*model.Variable{}
	groups := []*model.Variable{}
	for _, v := range sourceDS.Variables {
		varsSource[v.Key] = v
		if v.IsGrouping() {
			groups = append(groups, v)
		}
	}
	featuresExpanded := []string{}
	for _, f := range features {
		expandedFeature := getComponentVariables(varsSource[f])
		// make sure each feature is only included once
		for _, ef := range expandedFeature {
			if !varsExpanded[ef] {
				featuresExpanded = append(featuresExpanded, ef)
				varsExpanded[ef] = true
			}
		}
	}

	// extract the data from the database (result + base)
	data, err := dataStorage.FetchResultDataset(predictionDataset, predictionDS.StorageName, targetName, featuresExpanded, resultURI, true)
	if err != nil {
		return "", err
	}

	// read the source DS metadata from disk for the new dataset
	sourceDSDatasetPath := path.Join(env.ResolvePath(sourceDS.Source, sourceDS.Folder), compute.D3MDataSchema)
	metaDisk, err := metadata.LoadMetadataFromOriginalSchema(sourceDSDatasetPath, false)
	if err != nil {
		return "", err
	}
	mainDR := metaDisk.GetMainDataResource()

	varsMeta := map[string]*model.Variable{}
	for _, v := range mainDR.Variables {
		varsMeta[v.Key] = v
	}

	// map variables to get type info from source dataset and index from data
	// need the current types from the source dataset to have the proper definitive types
	varsNewDataset := make([]*model.Variable, len(data[0]))
	varsClassification := make([]*model.Variable, len(data[0]))
	for i, v := range data[0] {
		// min meta datasets do not have every variable so use source ES dataset if not in metadata
		variableMeta := varsMeta[v]
		if variableMeta == nil {
			variableMeta = varsSource[v]
		}
		if variableMeta == nil {
			// assume explanaibility output and create a new variable for it
			variableMeta = model.NewVariable(i, v, v, v, v, model.StringType, model.StringType,
				"", []string{"attribute"}, []string{model.VarDistilRoleData}, nil, nil, true)
		}
		variableMeta.Index = i
		variableMeta.SuggestedTypes = nil
		varsNewDataset[i] = variableMeta

		variableClassification := varsSource[v]
		if variableClassification != nil {
			variableClassification.Index = i
			varsClassification[i] = variableClassification
		} else {
			varsClassification[i] = variableMeta
		}
	}
	metaDisk.GetMainDataResource().Variables = varsNewDataset

	// make the data resource paths absolute
	predictionDSDatasetPath := path.Join(env.ResolvePath(predictionDS.Source, predictionDS.Folder), compute.D3MDataSchema)
	predictionDSDisk, err := serialization.ReadMetadata(predictionDSDatasetPath)
	if err != nil {
		return "", err
	}
	predictionDSDiskDRs := map[string]*model.DataResource{}
	for _, dr := range predictionDSDisk.DataResources {
		predictionDSDiskDRs[dr.ResID] = dr
	}
	for _, dr := range metaDisk.DataResources {
		if dr != mainDR {
			// TODO: NOT SURE WE CAN ASSUME RES ID EQUALITY!
			dr.ResPath = predictionDSDiskDRs[dr.ResID].ResPath
		}
	}

	// store the dataset to disk
	outputPath := env.ResolvePath(metadata.Augmented, newDatasetName)
	writer := serialization.GetStorage(metaDisk.GetMainDataResource().ResPath)

	newStorageName, err := dataStorage.GetStorageName(newDatasetName)
	if err != nil {
		return "", err
	}

	// update the header of the data since the data from the database uses keys as header
	data[0] = metaDisk.GetMainDataResource().GenerateHeader()

	metaDisk.ID = newDatasetName
	metaDisk.Name = newDatasetName
	metaDisk.StorageName = newStorageName
	metaDisk.DatasetFolder = newDatasetName
	rawDS := &serialization.RawDataset{
		ID:              metaDisk.ID,
		Name:            metaDisk.Name,
		Metadata:        metaDisk,
		Data:            data,
		DefinitiveTypes: true,
	}
	err = writer.WriteDataset(outputPath, rawDS)
	if err != nil {
		return "", err
	}
	classificationOutputPath := path.Join(outputPath, config.ClassificationOutputPath)
	classification := buildClassificationFromMetadata(varsClassification)
	classification.Path = classificationOutputPath
	err = metadata.WriteClassification(classification, classificationOutputPath)
	if err != nil {
		return "", err
	}

	// store new dataset metadata
	steps := &IngestSteps{
		VerifyMetadata:       false,
		FallbackMerged:       false,
		CreateMetadataTables: false,
	}
	params := &IngestParams{
		Source: metadata.Augmented,
		Type:   api.DatasetTypeModelling,
	}
	ingestConfig := NewConfig(config)
	cloneSchemaPath := path.Join(outputPath, compute.D3MDataSchema)
	_, err = IngestMetadata(cloneSchemaPath, cloneSchemaPath, nil, metaStorage, params, ingestConfig, steps)
	if err != nil {
		return "", err
	}

	// add all groups
	newDS, err := metaStorage.FetchDataset(newDatasetName, true, true, true)
	if err != nil {
		return "", err
	}
	for _, v := range groups {
		v.Index = len(newDS.Variables)
		newDS.Variables = append(newDS.Variables, v)
	}

	// if the prediction data is prefeaturized, then update the target variable with the new values
	if predictionDS.LearningDataset != "" {
		targetFolder := env.ResolvePath(newDS.Source, CreateFeaturizedDatasetID(newDatasetName))
		err := util.Copy(predictionDS.GetLearningFolder(), targetFolder)
		if err != nil {
			return "", err
		}
		err = updatePrefeaturizedDatasetVariable(targetFolder, varsSource[targetName].HeaderName, rawDS)
		if err != nil {
			return "", err
		}
		newDS.LearningDataset = targetFolder
	}

	err = metaStorage.UpdateDataset(newDS)
	if err != nil {
		return "", err
	}

	// ingest to postgres from disk
	err = IngestPostgres(cloneSchemaPath, cloneSchemaPath, params, ingestConfig, steps)
	if err != nil {
		return "", err
	}

	err = dataStorage.VerifyData(newDatasetName, newStorageName)
	if err != nil {
		return "", err
	}

	return metaDisk.ID, nil
}

func updatePrefeaturizedDatasetVariable(prefeaturizedPath string, variableName string, updatedData *serialization.RawDataset) error {
	log.Infof("updating variable '%s' in prefeaturized dataset found at '%s'", variableName, prefeaturizedPath)
	schemaPath := path.Join(prefeaturizedPath, compute.D3MDataSchema)
	dsDisk, err := serialization.ReadDataset(schemaPath)
	if err != nil {
		return err
	}

	dsDisk.Metadata.ID = updatedData.Metadata.ID
	dsDisk.Metadata.Name = updatedData.Metadata.Name
	dsDisk.Metadata.StorageName = updatedData.Metadata.StorageName

	// get the variable column to update and the d3m index column
	indicesPrefeaturized, err := getVariableIndices(dsDisk.Data[0], []string{variableName, model.D3MIndexFieldName})
	if err != nil {
		return err
	}
	indicesUpdated, err := getVariableIndices(updatedData.Data[0], []string{variableName, model.D3MIndexFieldName})
	if err != nil {
		return err
	}

	// create the updated data lookup
	updatedDataLookup := map[string]string{}
	for _, row := range updatedData.Data[1:] {
		updatedDataLookup[row[indicesUpdated[model.D3MIndexFieldName]]] = row[indicesUpdated[variableName]]
	}

	// cycle through updates
	for _, row := range dsDisk.Data[1:] {
		d3mIndexValue := row[indicesPrefeaturized[model.D3MIndexFieldName]]
		row[indicesPrefeaturized[variableName]] = updatedDataLookup[d3mIndexValue]
	}

	// output updated dataset
	err = serialization.WriteDataset(prefeaturizedPath, dsDisk)
	if err != nil {
		return err
	}

	return nil
}

func getVariableIndices(header []string, variables []string) (map[string]int, error) {
	indices := map[string]int{}
	for _, v := range variables {
		varIndex := getFieldIndex(header, v)
		if varIndex == -1 {
			return nil, errors.Errorf("variable '%s' does not exist in header", v)
		}
		indices[v] = varIndex
	}

	return indices, nil
}

// UpdateExtremas will update every field's extremas in the specified dataset.
func UpdateExtremas(dataset string, metaStorage api.MetadataStorage, dataStorage api.DataStorage) error {
	d, err := metaStorage.FetchDataset(dataset, false, false, false)
	if err != nil {
		return err
	}

	for _, v := range d.Variables {
		err = api.UpdateExtremas(dataset, v.Key, metaStorage, dataStorage)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetUniqueOutputFolder produces a unique name for a dataset in a folder.
func GetUniqueOutputFolder(dataset string, outputPath string) (string, error) {
	// read the folders in the output path
	files, err := ioutil.ReadDir(outputPath)
	if err != nil {
		return "", errors.Wrap(err, "unable to list output path content")
	}

	dirs := make([]string, 0)
	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		}
	}

	return getUniqueString(dataset, dirs), nil
}

// DeleteDataset deletes a dataset from metadata and, if not a soft delete, from the database.
func DeleteDataset(ds *api.Dataset, metaStorage api.MetadataStorage, dataStorage api.DataStorage, softDelete bool) error {
	// delete meta
	err := metaStorage.DeleteDataset(ds.ID, softDelete)
	if err != nil {
		return err
	}

	// delete the query cache associated wit this dataset if it exists
	DeleteQueryCache(ds.ID)

	if !softDelete {
		// delete db tables
		err = dataStorage.DeleteDataset(ds.StorageName)
		if err != nil {
			return err
		}
		// delete files
		err = util.RemoveContents(env.ResolvePath(ds.Source, ds.Folder), true)
		if err != nil {
			return err
		}

		// remove prefeaturized disk dataset
		if ds.LearningDataset != "" {
			err = util.RemoveContents(ds.LearningDataset, true)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getUniqueString(base string, existing []string) string {
	// create a unique name if the current name is already in use
	existingMap := make(map[string]bool)
	for _, e := range existing {
		existingMap[e] = true
	}

	unique := base
	for count := 1; existingMap[unique]; count++ {
		unique = fmt.Sprintf("%s_%d", base, count)
	}

	return unique
}

func buildClassificationFromMetadata(variables []*model.Variable) *model.ClassificationData {
	// cycle through the variables and collect the types
	classification := &model.ClassificationData{
		Labels:        make([][]string, len(variables)),
		Probabilities: make([][]float64, len(variables)),
	}
	for _, v := range variables {
		classification.Labels[v.Index] = []string{v.Type}
		classification.Probabilities[v.Index] = []float64{1}
	}

	return classification
}

func batchSubmitDataset(schemaFile string, dataset string, size int, submitFunc func(string) (string, error)) (string, error) {
	// get the storage to use
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile, false)
	if err != nil {
		return "", err
	}
	dataStorage := serialization.GetStorage(meta.GetMainDataResource().ResPath)

	// split the source dataset into batches
	schemaFiles, err := apicompute.CreateBatches(schemaFile, size)
	if err != nil {
		return "", err
	}
	defer func() {
		for _, sf := range schemaFiles {
			util.Delete(sf)
		}
	}()

	// submit each batch
	batchedResultSchemaFiles := []string{}
	for _, b := range schemaFiles {
		newFile, err := submitFunc(b)
		if err != nil {
			return "", err
		}

		batchedResultSchemaFiles = append(batchedResultSchemaFiles, newFile)
	}

	// join the results together
	completeData := [][]string{}
	for _, resultFile := range batchedResultSchemaFiles {
		data, err := dataStorage.ReadData(resultFile)
		if err != nil {
			return "", err
		}

		// grab the header off first batch read
		if len(completeData) == 0 {
			completeData = append(completeData, data[0])
		}
		completeData = append(completeData, data[1:]...)
	}

	// store the complete data
	hash, err := hashstructure.Hash([]interface{}{size, schemaFile, dataset}, nil)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate hashcode for %s", dataset)
	}
	hashFileName := fmt.Sprintf("%s-%0x", dataset, hash)
	outputURI := path.Join(env.GetTmpPath(), fmt.Sprintf("%s%s", hashFileName, path.Ext(meta.GetMainDataResource().ResPath)))
	outputURI = util.GetUniqueName(outputURI)
	err = dataStorage.WriteData(outputURI, completeData)
	if err != nil {
		return "", err
	}

	return outputURI, nil
}

func exportDiskDataset(dataset string, newDatasetID string, outputFolder string, metaStorage api.MetadataStorage,
	dataStorage api.DataStorage, limitSelectedFields bool, filterParams *api.FilterParams) (string, string, error) {
	metaDataset, err := metaStorage.FetchDataset(dataset, true, false, false)
	if err != nil {
		return "", "", err
	}
	meta := metaDataset.ToMetadata()
	meta.ID = newDatasetID

	data, err := dataStorage.FetchDataset(dataset, meta.StorageName, false, limitSelectedFields, filterParams)
	if err != nil {
		return "", "", err
	}

	// need to update metadata variable order to match extracted data
	header := data[0]
	exportedVariables := make([]*model.Variable, len(header))
	exportVarMap := api.MapVariables(metaDataset.Variables, func(variable *model.Variable) string { return variable.Key })
	for i, v := range header {
		variable := exportVarMap[v]
		variable.Index = i
		exportedVariables[i] = variable
	}
	meta.GetMainDataResource().Variables = exportedVariables

	// update the header with the proper variable names
	data[0] = meta.GetMainDataResource().GenerateHeader()
	dataRaw := &serialization.RawDataset{
		Name:     meta.Name,
		ID:       meta.ID,
		Data:     data,
		Metadata: meta,
	}

	// need to write the prefeaturized version of the dataset if it exists
	if metaDataset.ParentDataset != "" {
		log.Infof("exporting dataset %s that has parent dataset %s", dataset, metaDataset.ParentDataset)
		parentDS, err := metaStorage.FetchDataset(metaDataset.ParentDataset, false, false, false)
		if err != nil {
			return "", "", err
		}

		err = api.UpdateDiskDataset(metaDataset, data)
		if err != nil {
			return "", "", err
		}

		// read metadata of the parent from disk to get the non main data resources
		parentDatasetDoc := path.Join(env.ResolvePath(parentDS.Source, parentDS.Folder), compute.D3MDataSchema)
		parentMetaDisk, err := serialization.ReadMetadata(parentDatasetDoc)
		if err != nil {
			return "", "", err
		}
		parentMetaDiskMainDR := parentMetaDisk.GetMainDataResource()
		for _, dr := range parentMetaDisk.DataResources {
			if dr != parentMetaDiskMainDR {
				dr.ResPath = model.GetResourcePath(parentDatasetDoc, dr)
				meta.DataResources = append(meta.DataResources, dr)
			}
		}

		// main data resources need to be checked for any resource references
		parentVariablesMap := api.MapVariables(parentMetaDiskMainDR.Variables, func(variable *model.Variable) string { return variable.Key })
		for _, v := range dataRaw.Metadata.GetMainDataResource().Variables {
			parentVar := parentVariablesMap[v.Key]
			if parentVar != nil && parentVar.RefersTo != nil {
				v.RefersTo = parentVar.RefersTo
			}
		}
	}

	storage := serialization.GetCSVStorage()
	err = storage.WriteDataset(outputFolder, dataRaw)
	if err != nil {
		return "", "", err
	}

	return dataset, outputFolder, nil
}
