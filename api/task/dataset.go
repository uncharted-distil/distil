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

	// need to update metadata variable order to match extracted data
	header := data[0]
	exportedVariables := make([]*model.Variable, len(header))
	exportVarMap := mapVariables(metaDataset.Variables, func(variable *model.Variable) string { return variable.Key })
	for i, v := range header {
		variable := exportVarMap[v]
		variable.Index = i
		exportedVariables[i] = variable
	}
	meta.GetMainDataResource().Variables = exportedVariables

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

	// need to write the prefeaturized version of the dataset if it exists
	if metaDataset.ParentDataset != "" {
		err = updateLearningDataset(dataRaw, metaDataset, exportVarMap, metaStorage)
	}

	return dataset, outputFolder, err
}

func updateLearningDataset(newDataset *api.RawDataset, metaDataset *api.Dataset, exportVarMap map[string]*model.Variable, metaStorage api.MetadataStorage) error {
	parentDS, err := metaStorage.FetchDataset(metaDataset.ParentDataset, false, false, false)
	if err != nil {
		return err
	}

	if parentDS.LearningDataset == "" {
		return nil
	}

	// determine if there are new columns that were not part of the original dataset
	parentVarMap := mapVariables(parentDS.Variables, func(variable *model.Variable) string { return variable.Key })
	newVars := []*model.Variable{}
	for _, v := range metaDataset.Variables {
		if v.DistilRole == model.VarDistilRoleData {
			if parentVarMap[v.Key] == nil {
				newVars = append(newVars, v)
			}
		}
	}

	// copy the learning dataset
	learningFolder := fmt.Sprintf("%s-featurized", newDataset.Metadata.ID)
	learningFolder = path.Join(path.Dir(parentDS.LearningDataset), learningFolder)
	err = util.Copy(parentDS.LearningDataset, learningFolder)
	if err != nil {
		return err
	}

	// read the prefeaturized data (need to load the metadata to read the data)
	preFeaturizedDataset, err := serialization.ReadDataset(path.Join(learningFolder, compute.D3MDataSchema))
	if err != nil {
		return err
	}
	preFeaturizedMainDR := preFeaturizedDataset.Metadata.GetMainDataResource()
	preFeaturizedVarMap := mapVariables(preFeaturizedMainDR.Variables, func(variable *model.Variable) string { return variable.Key })
	preFeaturizedD3MIndex := preFeaturizedVarMap[model.D3MIndexFieldName].Index

	// add the missing columns row by row and only retain rows in the new dataset
	// first build up the new variables by d3m index map
	// then cycle through the featurized rows and append the variables
	newDSD3MIndex := exportVarMap[model.D3MIndexFieldName].Index
	newDataMap := map[string][]string{}
	for _, r := range newDataset.Data[1:] {
		newVarsData := []string{}
		for i := 0; i < len(newVars); i++ {
			newVarsData = append(newVarsData, r[newVars[i].Index])
		}
		newDataMap[r[newDSD3MIndex]] = newVarsData
	}

	// add the new fields to the metadata to generate the proper header
	for i := 0; i < len(newVars); i++ {
		newVar := newVars[i]
		newVar.Index = i + len(preFeaturizedMainDR.Variables)
		preFeaturizedMainDR.Variables = append(preFeaturizedMainDR.Variables, newVar)
	}

	preFeaturizedOutput := [][]string{preFeaturizedMainDR.GenerateHeader()}
	for _, row := range preFeaturizedDataset.Data[1:] {
		d3mIndexPre := row[preFeaturizedD3MIndex]
		if newDataMap[d3mIndexPre] != nil {
			rowComplete := append(row, newDataMap[d3mIndexPre]...)
			preFeaturizedOutput = append(preFeaturizedOutput, rowComplete)
		}
	}

	// output the new pre featurized data
	preFeaturizedDataset.Data = preFeaturizedOutput
	err = serialization.WriteDataset(learningFolder, preFeaturizedDataset)
	if err != nil {
		return err
	}

	// update the learning dataset for the new dataset
	metaDataset, err = metaStorage.FetchDataset(metaDataset.ID, true, true, true)
	if err != nil {
		return err
	}
	metaDataset.LearningDataset = learningFolder
	err = metaStorage.UpdateDataset(metaDataset)
	if err != nil {
		return err
	}

	return nil
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
