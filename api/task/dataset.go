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
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
)

const (
	baseMediaFolder = "media"
)

var (
	imageTypeMap = map[string]string{
		"png":  "png",
		"jpeg": "jpeg",
		"jpg":  "jpeg",
	}

	datasetStorage serialization.Storage
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
		datasetUnique, err := getUniqueOutputFolder(dataset, outputPath)
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

	err = datasetStorage.WriteData(dataPath, ds.Data)
	if err != nil {
		return "", "", err
	}

	schemaPath := path.Join(outputDatasetPath, compute.D3MDataSchema)
	err = datasetStorage.WriteMetadata(schemaPath, ds.Metadata, true)
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
		err = os.RemoveAll(outputDatasetPath)
		if err != nil {
			return "", "", err
		}

		err = util.Copy(formattedPath, path.Dir(schemaPath))
		if err != nil {
			return "", "", err
		}
	}

	return dataset, formattedPath, nil
}

func writeDataset(meta *model.Metadata, csvData []byte, outputPath string, config *IngestTaskConfig) (string, error) {
	// save the csv file in the file system datasets folder
	outputDatasetPath := path.Join(outputPath, meta.Name)
	dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)
	dataPath := path.Join(outputDatasetPath, dataFilePath)
	err := util.WriteFileWithDirs(dataPath, csvData, os.ModePerm)
	if err != nil {
		return "", err
	}

	schemaPath := path.Join(outputDatasetPath, compute.D3MDataSchema)
	err = datasetStorage.WriteMetadata(schemaPath, meta, true)
	if err != nil {
		return "", err
	}

	// format the dataset into a D3M format
	formattedPath, err := Format(schemaPath, meta.Name, config)
	if err != nil {
		return "", err
	}

	// copy to the original output location for consistency
	if formattedPath != outputDatasetPath {
		err = os.RemoveAll(outputDatasetPath)
		if err != nil {
			return "", err
		}

		err = util.Copy(formattedPath, path.Dir(schemaPath))
		if err != nil {
			return "", err
		}
	}

	return formattedPath, nil
}

// UpdateExtremas will update every field's extremas in the specified dataset.
func UpdateExtremas(dataset string, metaStorage api.MetadataStorage, dataStorage api.DataStorage) error {
	d, err := metaStorage.FetchDataset(dataset, false, false)
	if err != nil {
		return err
	}

	for _, v := range d.Variables {
		err = api.UpdateExtremas(dataset, v.Name, metaStorage, dataStorage)
		if err != nil {
			return err
		}
	}

	return nil
}

func getUniqueOutputFolder(dataset string, outputPath string) (string, error) {
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
