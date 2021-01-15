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
	CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*api.RawDataset, error)
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
		classification := buildClassificationFromMetadata(ds.Metadata)
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

// ExportDataset extracts a dataset from the database and metadata storage, writing
// it to disk in D3M dataset format.
func ExportDataset(dataset string, metaStorage api.MetadataStorage, dataStorage api.DataStorage, invert bool, filterParams *api.FilterParams) (string, string, error) {
	metaDataset, err := metaStorage.FetchDataset(dataset, true, false, false)
	if err != nil {
		return "", "", err
	}
	meta := metaDataset.ToMetadata()

	data, err := dataStorage.FetchDataset(dataset, meta.StorageName, invert, filterParams)
	if err != nil {
		return "", "", err
	}
	// update the header with the proper variable names
	data[0] = meta.GetMainDataResource().GenerateHeader()
	dataRaw := &api.RawDataset{
		Name:     meta.Name,
		ID:       meta.ID,
		Data:     data,
		Metadata: meta,
	}

	// TODO: most likely need to either get a unique folder name for output or error if already exists
	outputFolder := env.ResolvePath(metadata.Augmented, dataset)
	storage := serialization.GetCSVStorage()
	err = storage.WriteDataset(outputFolder, dataRaw)
	if err != nil {
		return "", "", err
	}

	return dataset, outputFolder, err
}

// CreateDatasetFromResult creates a new dataset based on a result set & the input
// to the model
func CreateDatasetFromResult(newDatasetName string, predictionDataset string, sourceDataset string, features []string,
	targetName string, resultURI string, metaStorage api.MetadataStorage, dataStorage api.DataStorage, config env.Config) (string, error) {
	newDatasetID := model.NormalizeDatasetID(newDatasetName)
	// get the prediction dataset
	predictionDS, err := metaStorage.FetchDataset(predictionDataset, true, true, true)
	if err != nil {
		return "", err
	}
	sourceDS, err := metaStorage.FetchDataset(sourceDataset, true, true, true)
	if err != nil {
		return "", err
	}

	// extract the data from the database (result + base)
	data, err := dataStorage.FetchResultDataset(predictionDataset, predictionDS.StorageName, targetName, features, resultURI)
	if err != nil {
		return "", err
	}

	// read the source DS metadata from disk for the new dataset
	predictionDSDatasetPath := env.ResolvePath(metadata.Augmented, path.Join(predictionDS.Folder, compute.D3MDataSchema))
	metaDisk, err := metadata.LoadMetadataFromOriginalSchema(predictionDSDatasetPath, false)
	if err != nil {
		return "", err
	}

	// map variables to get type info from source dataset and index from data
	varsSource := map[string]*model.Variable{}
	for _, v := range sourceDS.Variables {
		varsSource[v.Key] = v
	}
	varsNewDataset := make([]*model.Variable, len(metaDisk.GetMainDataResource().Variables))
	for i, v := range data[0] {
		variable := varsSource[v]
		variable.Index = i
		varsNewDataset[i] = variable
	}
	metaDisk.GetMainDataResource().Variables = varsNewDataset

	// store the dataset to disk
	outputPath := env.ResolvePath(metadata.Augmented, newDatasetName)
	writer := serialization.GetStorage(metaDisk.GetMainDataResource().ResPath)

	newStorageName, err := dataStorage.GetStorageName(newDatasetName)
	if err != nil {
		return "", err
	}

	metaDisk.ID = newDatasetID
	metaDisk.Name = newDatasetName
	metaDisk.StorageName = newStorageName
	metaDisk.DatasetFolder = newDatasetName
	rawDS := &api.RawDataset{
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

	// store new dataset metadata
	err = metaStorage.IngestDataset(metadata.Augmented, metaDisk)
	if err != nil {
		return "", err
	}

	// ingest to postgres from disk
	cloneSchemaPath := path.Join(outputPath, compute.D3MDataSchema)
	err = IngestPostgres(cloneSchemaPath, cloneSchemaPath, metadata.Augmented, nil, NewConfig(config), false, false, true)
	if err != nil {
		return "", err
	}

	return metaDisk.ID, nil
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

func buildClassificationFromMetadata(meta *model.Metadata) *model.ClassificationData {
	// cycle through the variables and collect the types
	mainDR := meta.GetMainDataResource()
	classification := &model.ClassificationData{
		Labels:        make([][]string, len(mainDR.Variables)),
		Probabilities: make([][]float64, len(mainDR.Variables)),
	}
	for _, v := range mainDR.Variables {
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
