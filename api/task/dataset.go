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
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
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
)

// DatasetConstructor is used to build a dataset.
type DatasetConstructor interface {
	CreateDataset(rootDataPath string, config *env.Config) (*api.RawDataset, error)
}

// CreateDataset structures a raw csv file into a valid D3M dataset.
func CreateDataset(dataset string, datasetCtor DatasetConstructor, outputPath string, typ api.DatasetType, config *env.Config) (string, error) {
	ingestConfig := NewConfig(*config)

	// save the csv file in the file system datasets folder
	var err error
	outputDatasetPath := path.Join(outputPath, dataset)
	if !config.IngestOverwrite {
		outputDatasetPath, err = getUniqueOutputFolder(outputDatasetPath, outputPath)
		if err != nil {
			return "", err
		}
	}

	dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)
	dataPath := path.Join(outputDatasetPath, dataFilePath)

	ds, err := datasetCtor.CreateDataset(outputDatasetPath, config)
	if err != nil {
		return "", err
	}

	var outputBuffer bytes.Buffer
	csvWriter := csv.NewWriter(&outputBuffer)
	err = csvWriter.WriteAll(ds.Data)
	if err != nil {
		return "", errors.Wrap(err, "unable to write csv data to buffer")
	}
	csvWriter.Flush()

	err = util.WriteFileWithDirs(dataPath, outputBuffer.Bytes(), os.ModePerm)
	if err != nil {
		return "", err
	}

	schemaPath := path.Join(outputDatasetPath, compute.D3MDataSchema)
	err = metadata.WriteSchema(ds.Metadata, schemaPath, true)
	if err != nil {
		return "", err
	}

	// format the dataset into a D3M format
	formattedPath, err := Format(metadata.Contrib, schemaPath, dataset, ingestConfig)
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
	err = metadata.WriteSchema(meta, schemaPath, true)
	if err != nil {
		return "", err
	}

	// format the dataset into a D3M format
	formattedPath, err := Format(metadata.Contrib, schemaPath, meta.Name, config)
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

func getUniqueOutputFolder(datasetPath string, outputPath string) (string, error) {
	// read the folders in the output path
	datasets, err := util.GetDirectories(outputPath)
	if err != nil {
		return "", err
	}

	uniqueDataset := getUniqueString(datasetPath, datasets)

	return uniqueDataset, nil
}

func getUniqueString(base string, existing []string) string {
	// create a unique name if the current name is already in use
	existingMap := make(map[string]bool)
	for _, e := range existing {
		existingMap[e] = true
	}

	unique := base
	for count := 1; !existingMap[unique]; count++ {
		unique = fmt.Sprintf("%s_%d", base, count)
	}

	return unique
}
