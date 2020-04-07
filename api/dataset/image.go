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

package dataset

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
)

const (
	defaultImageType = "jpeg"
)

var (
	imageTypeMap = map[string]string{
		"png":  "png",
		"jpeg": "jpeg",
		"jpg":  "jpeg",
	}
	imageTypeContentMap = map[string][]string{
		"jpeg": {"jpeg", "jpg"},
	}
)

// Image captures the data in an image dataset.
type Image struct {
	Dataset           string `json:"dataset"`
	ImageType         string `json:"imageType"`
	RawFilePath       string `json:"rawFilePath"`
	ExtractedFilePath string `json:"extractedFilePath"`
}

// NewImageDataset creates a new image dataset from raw byte data, assuming json.
func NewImageDataset(dataset string, imageType string, rawData []byte, config *env.Config) (*Image, error) {
	outputPath := path.Join(config.D3MOutputDir, config.AugmentedSubFolder)
	outputDatasetPath := path.Join(outputPath, dataset)

	// clear the output dataset path location
	err := util.RemoveContents(outputDatasetPath)
	if err != nil {
		log.Warnf("unable to remove contents: %v", err)
	}

	// store and expand raw data
	zipFilename := path.Join(outputDatasetPath, "raw.zip")
	err = util.WriteFileWithDirs(zipFilename, rawData, os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "unable to write raw image data archive")
	}
	extractedArchivePath := outputDatasetPath
	err = util.Unzip(zipFilename, extractedArchivePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to extract raw image data archive")
	}

	return &Image{
		Dataset:           dataset,
		ImageType:         imageType,
		RawFilePath:       zipFilename,
		ExtractedFilePath: extractedArchivePath,
	}, nil
}

// CreateDataset processes the raw image dataset and creates a raw D3M dataset.
func (i *Image) CreateDataset(rootDataPath string, config *env.Config) (*api.RawDataset, error) {
	outputPath := path.Join(config.D3MOutputDir, config.AugmentedSubFolder)
	outputDatasetPath := path.Join(outputPath, i.Dataset)
	dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)

	imageFolders := make([]string, 0)
	extractedFiles, err := ioutil.ReadDir(i.ExtractedFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read extracted data")
	}
	for _, f := range extractedFiles {
		if f.IsDir() {
			imageFolders = append(imageFolders, path.Join(i.ExtractedFilePath, f.Name()))
		}
	}

	csvData := make([][]string, 0)
	csvData = append(csvData, []string{model.D3MIndexFieldName, "image_file", "label"})
	mediaFolder := getUniqueFolder(path.Join(outputDatasetPath, "media"))

	// the folder name represents the label to apply for all containing images
	for _, imageFolder := range imageFolders {
		log.Infof("processing image folder '%s'", imageFolder)
		label := path.Base(imageFolder)

		imageFiles, err := ioutil.ReadDir(imageFolder)
		if err != nil {
			return nil, err
		}

		// copy images while building the csv data
		log.Infof("building csv data")
		for _, imageFile := range imageFiles {
			imageFilename := imageFile.Name()
			imageFilenameFull := path.Join(imageFolder, imageFilename)

			imageLoaded, err := readImage(imageFilenameFull, i.ImageType)
			if err != nil {
				return nil, err
			}

			targetImageFilename := imageFilename
			if path.Ext(targetImageFilename) != fmt.Sprintf(".%s", defaultImageType) {
				targetImageFilename = fmt.Sprintf("%s.%s", imageFilename, defaultImageType)
			}
			targetImageFilename = getUniqueName(path.Join(mediaFolder, targetImageFilename))

			imageOutput, err := toJPEG(&imageLoaded)
			if err != nil {
				return nil, err
			}

			err = util.WriteFileWithDirs(targetImageFilename, imageOutput, os.ModePerm)
			if err != nil {
				return nil, errors.Wrap(err, "unable to save processed image file")
			}

			csvData = append(csvData, []string{fmt.Sprintf("%d", len(csvData)-1), path.Base(targetImageFilename), label})
		}
	}

	log.Infof("creating metadata")

	// create the dataset schema doc
	datasetID := model.NormalizeDatasetID(i.Dataset)
	meta := model.NewMetadata(i.Dataset, i.Dataset, "", datasetID)
	dr := model.NewDataResource(compute.DefaultResourceID, model.ResTypeTable, map[string][]string{compute.D3MResourceFormat: {"csv"}})
	dr.ResPath = dataFilePath
	dr.Variables = append(dr.Variables,
		model.NewVariable(0, model.D3MIndexFieldName, model.D3MIndexFieldName,
			model.D3MIndexFieldName, model.IntegerType, model.IntegerType, "D3M index",
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
	imageTypeLookup := imageTypeMap[defaultImageType]
	refDR := model.NewDataResource("0", model.ResTypeImage, map[string][]string{fmt.Sprintf("image/%s", imageTypeLookup): imageTypeContentMap[imageTypeLookup]})
	refDR.ResPath = path.Base(mediaFolder)
	refDR.IsCollection = true

	meta.DataResources = []*model.DataResource{refDR, dr}

	return &api.RawDataset{
		ID:       datasetID,
		Name:     i.Dataset,
		Data:     csvData,
		Metadata: meta,
	}, nil
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

func readImage(imagePath string, defaultType string) (image.Image, error) {
	// decode the image
	imageRaw, err := ioutil.ReadFile(imagePath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read image '%s'", imagePath)
	}

	typ := path.Ext(imagePath)
	if typ == "" {
		typ = defaultType
	} else {
		typ = typ[1:]
	}

	switch typ {
	case "png":
		return png.Decode(bytes.NewReader(imageRaw))
	case "jpg", "jpeg":
		return jpeg.Decode(bytes.NewReader(imageRaw))
	default:
		return nil, errors.Errorf("unsupported image type '%s'", typ)
	}
}

func toJPEG(img *image.Image) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, *img, nil); err != nil {
		return nil, errors.Wrap(err, "unable to encode jpg")
	}

	return buf.Bytes(), nil
}

func toPNG(img *image.Image) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, *img); err != nil {
		return nil, errors.Wrap(err, "unable to encode png")
	}

	return buf.Bytes(), nil
}
