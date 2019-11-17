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
	log "github.com/unchartedsoftware/plog"

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
	log.Infof("creating image dataset '%s' of type '%s'", dataset, imageType)
	outputDatasetPath := path.Join(outputPath, dataset)
	schemaPath := path.Join(outputDatasetPath, compute.D3MDataSchema)
	dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)
	dataPath := path.Join(outputDatasetPath, dataFilePath)

	csvData := make([][]string, 0)
	csvData = append(csvData, []string{model.D3MIndexFieldName, "image_file", "label"})
	mediaFolder := getUniqueFolder(path.Join(outputDatasetPath, "media"))

	err := os.MkdirAll(outputDatasetPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	// the folder name represents the label to apply for all containing images
	for _, imageFolder := range imageFolders {
		log.Infof("processing image folder '%s'", imageFolder)
		label := path.Base(imageFolder)

		imageFiles, err := ioutil.ReadDir(imageFolder)
		if err != nil {
			return "", err
		}

		// copy images while building the csv data
		log.Infof("building csv data")
		for _, imageFile := range imageFiles {
			imageFilename := path.Base(imageFile.Name())
			if path.Ext(imageFilename) != fmt.Sprintf(".%s", imageType) {
				imageFilename = fmt.Sprintf("%s.%s", imageFilename, imageType)
			}
			imageFilename = getUniqueName(path.Join(mediaFolder, imageFilename))

			err = util.Copy(path.Join(imageFolder, imageFile.Name()), imageFilename)
			if err != nil {
				return "", err
			}

			csvData = append(csvData, []string{fmt.Sprintf("%d", len(csvData)-1), imageFilename, label})
		}
	}

	log.Infof("creating metadata")

	// create the dataset schema doc
	datasetID := model.NormalizeDatasetID(dataset)
	meta := model.NewMetadata(dataset, dataset, "", datasetID)
	dr := model.NewDataResource("learningData", model.ResTypeTable, []string{compute.D3MResourceFormat})
	dr.ResPath = dataFilePath
	dr.Variables = append(dr.Variables,
		model.NewVariable(0, model.D3MIndexFieldName, model.D3MIndexFieldName,
			model.D3MIndexFieldName, model.IndexType, model.IndexType, "D3M index",
			[]string{model.RoleIndex}, model.VarRoleIndex, nil, dr.Variables, false),
	)
	dr.Variables = append(dr.Variables,
		model.NewVariable(1, "image_file", "image_file", "image_file", model.StringType,
			model.StringType, "Reference to image file", []string{"attribute"},
			model.VarRoleData, map[string]interface{}{"resID": "0", "resObject": "item"}, dr.Variables, false))
	dr.Variables = append(dr.Variables,
		model.NewVariable(2, "label", "label", "label", model.StringType,
			model.StringType, "Label of the image", []string{"suggestedTarget"},
			model.VarRoleData, nil, dr.Variables, false))

	// create the data resource for the referenced images
	refDR := model.NewDataResource("0", model.ResTypeImage, []string{fmt.Sprintf("image/%s", imageType)})
	refDR.ResPath = path.Base(mediaFolder)

	meta.DataResources = []*model.DataResource{refDR, dr}

	log.Infof("writing schema to '%s'", schemaPath)
	err = metadata.WriteSchema(meta, schemaPath)
	if err != nil {
		return "", err
	}

	// write out the dataset
	buf := bytes.NewBuffer(nil)
	csvOutput := csv.NewWriter(buf)
	err = csvOutput.WriteAll(csvData)
	if err != nil {
		return "", err
	}

	err = util.WriteFileWithDirs(dataPath, buf.Bytes(), os.ModePerm)
	if err != nil {
		return "", err
	}

	return outputDatasetPath, nil
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

func getUniqueName(filename string) string {
	extension := path.Ext(filename)
	baseFilename := strings.TrimSuffix(filename, extension)
	currentFilename := filename
	for i := 1; util.FileExists(currentFilename); i++ {
		currentFilename = fmt.Sprintf("%s_%d.%s", baseFilename, i, extension)
	}

	return currentFilename
}

func getUniqueFolder(folder string) string {
	currentFilename := folder
	for i := 1; util.FileExists(currentFilename); i++ {
		currentFilename = fmt.Sprintf("%s_%d", folder, i)
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
