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

var (
	mediaTypeMap = map[string]string{
		"png":  "image",
		"jpeg": "image",
		"jpg":  "image",
		"txt":  "text",
	}
	mediaFormatMap = map[string]string{
		"jpeg": "image/jpeg",
		"jpg":  "image/jpeg",
		"txt":  "text/plain",
	}
	mediaTypeContentMap = map[string][]string{
		"jpeg": {"jpeg", "jpg"},
		"txt":  {"txt"},
	}
)

// Media captures the data in a media dataset.
type Media struct {
	Dataset           string `json:"dataset"`
	MediaType         string `json:"mediaType"`
	TargetMediaType   string `json:"targetMediaType"`
	RawFilePath       string `json:"rawFilePath"`
	ExtractedFilePath string `json:"extractedFilePath"`
}

// NewMediaDataset creates a new media dataset from raw byte data, assuming json.
func NewMediaDataset(dataset string, mediaType string, targetMediaType string, rawData []byte) (*Media, error) {
	// store and expand raw data
	zipPath, err := StoreZipDataset(dataset, rawData)
	if err != nil {
		return nil, err
	}

	expandedInfo, err := ExpandZipDataset(zipPath, dataset)
	if err != nil {
		return nil, err
	}

	return &Media{
		Dataset:           dataset,
		MediaType:         mediaType,
		TargetMediaType:   targetMediaType,
		RawFilePath:       expandedInfo.RawFilePath,
		ExtractedFilePath: expandedInfo.ExtractedFilePath,
	}, nil
}

// NewMediaDatasetFromExpanded creates a new media dataset from raw byte data, assuming json.
func NewMediaDatasetFromExpanded(dataset string, mediaType string, targetMediaType string, zipFileName string, extractedArchivePath string) (*Media, error) {
	return &Media{
		Dataset:           dataset,
		MediaType:         mediaType,
		TargetMediaType:   targetMediaType,
		RawFilePath:       zipFileName,
		ExtractedFilePath: extractedArchivePath,
	}, nil
}

// CreateDataset processes the raw media dataset and creates a raw D3M dataset.
func (m *Media) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*api.RawDataset, error) {
	if datasetName == "" {
		datasetName = m.Dataset
	}
	outputDatasetPath := rootDataPath
	dataFilePath := path.Join(outputDatasetPath, compute.D3MDataFolder, compute.D3MLearningData)

	labelFolders, err := getLabelFolders(m.ExtractedFilePath)
	if err != nil {
		return nil, err
	}

	csvData := make([][]string, 0)
	csvData = append(csvData, []string{model.D3MIndexFieldName, "media_file", "label"})
	mediaFolder := util.GetUniqueFolder(path.Join(outputDatasetPath, "media"))

	// the folder name represents the label to apply for all containing media
	totalCounts := make(map[string]int)
	successCounts := make(map[string]int)
	for _, folder := range labelFolders {
		log.Infof("processing label folder '%s'", folder)
		label := path.Base(folder)

		mediaFiles, err := ioutil.ReadDir(folder)
		if err != nil {
			return nil, err
		}

		// copy media while building the csv data
		log.Infof("building csv data")
		for _, file := range mediaFiles {
			mediaFilename := file.Name()
			mediaFilenameFull := path.Join(folder, mediaFilename)
			totalCounts[label] = totalCounts[label] + 1

			mediaLoaded, err := loadMedia(mediaFilenameFull, m.MediaType)
			if err != nil {
				log.Warnf("unable to load media '%s': %v", mediaFilenameFull, err)
				continue
			}

			targetMediaFilename := mediaFilename
			extension := path.Ext(targetMediaFilename)
			if extension != fmt.Sprintf(".%s", m.TargetMediaType) {
				targetMediaFilename = fmt.Sprintf("%s.%s", strings.TrimSuffix(targetMediaFilename, extension), m.TargetMediaType)
			}
			targetMediaFilename = util.GetUniqueName(path.Join(mediaFolder, targetMediaFilename))

			err = util.WriteFileWithDirs(targetMediaFilename, mediaLoaded, os.ModePerm)
			if err != nil {
				log.Warnf("unable to save processed media file '%s': %v", mediaFilenameFull, err)
				continue
			}

			csvData = append(csvData, []string{fmt.Sprintf("%d", len(csvData)-1), path.Base(targetMediaFilename), label})
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
		model.NewVariable(0, model.D3MIndexFieldName, model.D3MIndexFieldName, model.D3MIndexFieldName,
			model.D3MIndexFieldName, model.IntegerType, model.IntegerType, "D3M index",
			[]string{model.RoleIndex}, model.VarDistilRoleIndex, nil, dr.Variables, false),
	)
	dr.Variables = append(dr.Variables,
		model.NewVariable(1, "media_file", "media_file", "media_file", "media_file", model.StringType,
			model.StringType, "Reference to media file", []string{"attribute"},
			model.VarDistilRoleData, map[string]interface{}{"resID": "0", "resObject": "item"}, dr.Variables, false))
	dr.Variables = append(dr.Variables,
		model.NewVariable(2, "label", "label", "label", "label", model.StringType,
			model.StringType, "Label of the media", []string{"suggestedTarget"},
			model.VarDistilRoleData, nil, dr.Variables, false))

	// create the data resource for the referenced media
	refDR := model.NewDataResource("0", mediaTypeMap[m.TargetMediaType], map[string][]string{mediaFormatMap[m.TargetMediaType]: mediaTypeContentMap[m.TargetMediaType]})
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

func loadMedia(filename string, typ string) ([]byte, error) {
	var data []byte
	var err error
	switch typ {
	case "png", "jpg", "jpeg":
		imageLoaded, err := readImage(filename, typ)
		if err != nil {
			return nil, err
		}

		data, err = toJPEG(&imageLoaded)
		if err != nil {
			return nil, err
		}
		break
	case "txt":
		data, err = ioutil.ReadFile(filename)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read file")
		}
		break
	}

	return data, nil
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

func getLabelFolders(folderPath string) ([]string, error) {
	labelFolders := make([]string, 0)
	extractedFiles, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read extracted data")
	}
	for _, f := range extractedFiles {
		if f.IsDir() {
			labelFolders = append(labelFolders, path.Join(folderPath, f.Name()))
		}
	}

	return labelFolders, nil
}
