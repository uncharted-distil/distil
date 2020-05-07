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
func NewImageDataset(dataset string, imageType string, rawData []byte) (*Image, error) {
	// store and expand raw data
	expandedInfo, err := ExpandZipDataset(dataset, rawData)
	if err != nil {
		return nil, err
	}
	return &Image{
		Dataset:           dataset,
		ImageType:         imageType,
		RawFilePath:       expandedInfo.RawFilePath,
		ExtractedFilePath: expandedInfo.ExtractedFilePath,
	}, nil
}

// NewImageDataset creates a new image dataset from raw byte data, assuming json.
func NewImageDatasetFromExpanded(dataset string, imageType string, zipFileName string, extractedArchivePath string) (*Image, error) {
	return &Image{
		Dataset:           dataset,
		ImageType:         imageType,
		RawFilePath:       zipFileName,
		ExtractedFilePath: extractedArchivePath,
	}, nil
}

// CreateDataset processes the raw image dataset and creates a raw D3M dataset.
func (i *Image) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*api.RawDataset, error) {
	if datasetName == "" {
		datasetName = i.Dataset
	}
	outputDatasetPath := rootDataPath
	dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)

	imageFolders, err := getImageFolders(i.ExtractedFilePath)
	if err != nil {
		return nil, err
	}

	csvData := make([][]string, 0)
	csvData = append(csvData, []string{model.D3MIndexFieldName, "image_file", "label"})
	mediaFolder := getUniqueFolder(path.Join(outputDatasetPath, "media"))

	// the folder name represents the label to apply for all containing images
	totalCounts := make(map[string]int)
	successCounts := make(map[string]int)
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
			totalCounts[label] = totalCounts[label] + 1

			imageLoaded, err := readImage(imageFilenameFull, i.ImageType)
			if err != nil {
				log.Warnf("unable to read image '%s': %v", imageFilename, err)
				continue
			}

			targetImageFilename := imageFilename
			extension := path.Ext(targetImageFilename)
			if extension != fmt.Sprintf(".%s", defaultImageType) {
				targetImageFilename = fmt.Sprintf("%s.%s", strings.TrimSuffix(targetImageFilename, extension), defaultImageType)
			}
			targetImageFilename = getUniqueName(path.Join(mediaFolder, targetImageFilename))

			imageOutput, err := toJPEG(&imageLoaded)
			if err != nil {
				log.Warnf("unable to convert image '%s': %v", imageFilename, err)
				continue
			}

			err = util.WriteFileWithDirs(targetImageFilename, imageOutput, os.ModePerm)
			if err != nil {
				log.Warnf("unable to save processed image file '%s': %v", imageFilename, err)
				continue
			}

			csvData = append(csvData, []string{fmt.Sprintf("%d", len(csvData)-1), path.Base(targetImageFilename), label})
			successCounts[label] = successCounts[label] + 1
		}
	}

	// check counts and flag if too many errors
	for label, count := range totalCounts {
		successes := successCounts[label]
		if successes < int(float64(count)*(1.0-config.ImportErrorThreshold)) {
			return nil, errors.Errorf("too many errors when processing label '%s' (%d out of %d failed)", label, count-successes, count)
		}
	}

	log.Infof("creating metadata")

	// create the dataset schema doc
	datasetID := model.NormalizeDatasetID(datasetName)
	meta := model.NewMetadata(datasetName, datasetName, "", datasetID)
	dr := model.NewDataResource(compute.DefaultResourceID, model.ResTypeTable, map[string][]string{compute.D3MResourceFormat: {"csv"}})
	dr.ResPath = dataFilePath
	dr.Variables = append(dr.Variables,
		model.NewVariable(0, model.D3MIndexFieldName, model.D3MIndexFieldName,
			model.D3MIndexFieldName, model.IntegerType, model.IntegerType, "D3M index",
			[]string{model.RoleIndex}, model.VarDistilRoleIndex, nil, dr.Variables, false),
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
		Name:     datasetName,
		Data:     csvData,
		Metadata: meta,
	}, nil
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

func getImageFolders(folderPath string) ([]string, error) {
	imageFolders := make([]string, 0)
	extractedFiles, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read extracted data")
	}
	for _, f := range extractedFiles {
		if f.IsDir() {
			imageFolders = append(imageFolders, path.Join(folderPath, f.Name()))
		}
	}

	return imageFolders, nil
}
