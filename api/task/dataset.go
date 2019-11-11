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
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"

	"github.com/uncharted-distil/distil-ingest/metadata"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
)

const (
	baseMediaFolder = "media"
)

// CreateDataset structures a raw csv file into a valid D3M dataset.
func CreateDataset(dataset string, csvData []byte, outputPath string, config *IngestTaskConfig) (string, error) {
	// save the csv file in the file system datasets folder
	outputDatasetPath := path.Join(outputPath, dataset)
	dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)
	dataPath := path.Join(outputDatasetPath, dataFilePath)
	err := util.WriteFileWithDirs(dataPath, csvData, os.ModePerm)
	if err != nil {
		return "", err
	}

	// create the raw dataset schema doc
	datasetID := model.NormalizeDatasetID(dataset)
	meta := model.NewMetadata(dataset, dataset, "", datasetID)
	dr := model.NewDataResource("learningData", model.ResTypeRaw, []string{compute.D3MResourceFormat})
	dr.ResPath = dataFilePath
	meta.DataResources = []*model.DataResource{dr}

	schemaPath := path.Join(outputDatasetPath, compute.D3MDataSchema)
	err = metadata.WriteSchema(meta, schemaPath)
	if err != nil {
		return "", err
	}

	// format the dataset into a D3M format
	formattedPath, err := Format(metadata.Contrib, schemaPath, config)
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

// CreateImageDataset creates a D3M dataset from a collection of folders that
// each contain images. The name of the folder represents the label to give
// for the images within.
func CreateImageDataset(dataset string, imageFolders []string, imageType string, outputPath string, config *IngestTaskConfig) (string, error) {
	// generate all the image data for the csv table
	csvData := make([][]string, 0)
	csvData = append(csvData, []string{model.D3MIndexFieldName, "image_file", "label"})
	mediaFolder := getMediaFolder(imageFolders)

	// the folder name represents the label to apply for all containing images
	for _, imageFolder := range imageFolders {
		label := path.Base(imageFolder)
		datasetFolder := path.Dir(imageFolder)

		imageFiles, err := ioutil.ReadDir(imageFolder)
		if err != nil {
			return "", err
		}

		// copy images while building the csv data
		d3mIndex := 0
		for _, imageFile := range imageFiles {
			imageFilename := imageFile.Name()
			if path.Ext(imageFilename) != imageType {
				imageFilename = fmt.Sprintf("%s.%s", imageFilename, imageType)
			}

			err = util.Copy(path.Join(imageFolder, imageFile.Name()), getUniqueName(path.Join(datasetFolder, mediaFolder, imageFilename)))
			if err != nil {
				return "", err
			}

			csvData = append(csvData, []string{fmt.Sprintf("%d", d3mIndex), imageFilename, label})
		}
	}

	// create the data resource for the referenced images
	refDR := model.NewDataResource("0", "image", []string{fmt.Sprintf("image/%s", imageType)})

	// create the D3M dataset from the csv data
	meta := createMetadata(dataset, config)

	// add the image information
	meta.DataResources = append(meta.DataResources, refDR)
	meta.DataResources[0].Variables[1].RefersTo["resID"] = "0"
	meta.DataResources[0].Variables[1].RefersTo["resObject"] = "item"

	// write out the dataset
	buf := bytes.NewBuffer(nil)
	csvOutput := csv.NewWriter(buf)
	err := csvOutput.WriteAll(csvData)
	if err != nil {
		return "", err
	}

	outputPathWritten, err := writeDataset(meta, buf.Bytes(), outputPath, config)
	if err != nil {
		return "", err
	}

	return outputPathWritten, nil
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
	err = metadata.WriteSchema(meta, schemaPath)
	if err != nil {
		return "", err
	}

	// format the dataset into a D3M format
	formattedPath, err := Format(metadata.Contrib, schemaPath, config)
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

func createMetadata(dataset string, config *IngestTaskConfig) *model.Metadata {
	dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)

	// create the raw dataset schema doc
	datasetID := model.NormalizeDatasetID(dataset)
	meta := model.NewMetadata(dataset, dataset, "", datasetID)
	dr := model.NewDataResource("learningData", model.ResTypeRaw, []string{compute.D3MResourceFormat})
	dr.ResPath = dataFilePath
	meta.DataResources = []*model.DataResource{dr}

	return meta
}

func getMediaFolder(refFolders []string) string {
	duplicateCount := 0
	for _, folder := range refFolders {
		if strings.HasPrefix(folder, baseMediaFolder) {
			duplicateCount = duplicateCount + 1
		}
	}

	mediaFolder := baseMediaFolder
	if duplicateCount > 0 {
		mediaFolder = fmt.Sprintf("%s_%d", mediaFolder, duplicateCount)
	}

	return mediaFolder
}

func getUniqueName(filename string) string {
	currentFilename := filename
	for i := 1; util.FileExists(currentFilename); {
		currentFilename = fmt.Sprintf("%s_%d", filename, i)
	}

	return currentFilename
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
