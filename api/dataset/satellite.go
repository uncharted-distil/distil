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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/lukeroth/gdal"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
)

var (
	satTypeMap = map[string]string{
		"tif":  "tiff",
		"tiff": "tiff",
	}
	satTypeContentMap = map[string][]string{
		"tiff": {"tif", "tiff"},
	}

	bandRegex = regexp.MustCompile(`_B[0-9][0-9a-zA-Z][.]`)
)

// Satellite captures the data in a satellite (remote sensing) dataset.
type Satellite struct {
	Dataset           string `json:"dataset"`
	ImageType         string `json:"imageType"`
	RawFilePath       string `json:"rawFilePath"`
	ExtractedFilePath string `json:"extractedFilePath"`
}

// BoundingBox is a box delineated by four corners.
type BoundingBox struct {
	UpperLeft  *Point
	UpperRight *Point
	LowerLeft  *Point
	LowerRight *Point
}

// Point represents a coordinate in 2d space.
type Point struct {
	X float64
	Y float64
}

// ToString writes out the bounding box to a string.
func (b *BoundingBox) ToString() string {
	coords := []string{
		b.pointToString(b.LowerLeft),
		b.pointToString(b.UpperLeft),
		b.pointToString(b.UpperRight),
		b.pointToString(b.LowerRight),
	}
	return strings.Join(coords, ",")
}

func (b *BoundingBox) pointToString(point *Point) string {
	if point != nil {
		return fmt.Sprintf("%f,%f", point.X, point.Y)
	}

	return ","
}

// NewSatelliteDataset creates a new satelitte dataset from geotiff files
func NewSatelliteDataset(dataset string, imageType string, rawData []byte, config *env.Config) (*Satellite, error) {
	// store and expand raw data
	tmpPath := env.GetTmpPath()
	zipFilename := path.Join(tmpPath, fmt.Sprintf("%s_raw.zip", dataset))
	zipFilename = getUniqueName(zipFilename)
	err := util.WriteFileWithDirs(zipFilename, rawData, os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "unable to write raw image data archive")
	}

	extractedArchivePath := getUniqueFolder(path.Join(tmpPath, dataset))
	err = util.Unzip(zipFilename, extractedArchivePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to extract raw image data archive")
	}

	return &Satellite{
		Dataset:           dataset,
		ImageType:         imageType,
		RawFilePath:       zipFilename,
		ExtractedFilePath: extractedArchivePath,
	}, nil
}

// CreateDataset processes the raw satellite dataset and creates a raw D3M dataset.
func (s *Satellite) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*api.RawDataset, error) {
	if datasetName == "" {
		datasetName = s.Dataset
	}
	outputDatasetPath := rootDataPath
	dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)

	imageFolders, err := getImageFolders(s.ExtractedFilePath)
	if err != nil {
		return nil, err
	}

	csvData := make([][]string, 0)
	csvData = append(csvData, []string{model.D3MIndexFieldName, "image_file", "group_id", "band", "coordinates", "label"})
	mediaFolder := getUniqueFolder(path.Join(outputDatasetPath, "media"))

	// need to keep track of d3m Index values since they are shared for a whole group
	d3mIDs := make(map[string]int)
	d3mIDRunning := 1

	// the folder name represents the label to apply for all containing images
	for _, imageFolder := range imageFolders {
		log.Infof("processing satellite image folder '%s'", imageFolder)
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

			ok := verifySatelliteImage(imageFilenameFull, s.ImageType)
			if !ok {
				log.Warnf("'%s' is not a valid or supported satellite image", imageFilenameFull)
				continue
			}

			targetImageFilename := imageFilename
			extension := path.Ext(targetImageFilename)
			if extension != fmt.Sprintf(".%s", s.ImageType) {
				targetImageFilename = fmt.Sprintf("%s.%s", strings.TrimSuffix(targetImageFilename, extension), s.ImageType)
			}
			targetImageFilename = getUniqueName(path.Join(mediaFolder, targetImageFilename))

			err = util.CopyFile(imageFilenameFull, targetImageFilename)
			if err != nil {
				return nil, errors.Wrap(err, "unable to copy image file")
			}

			coordinates, err := extractCoordinates(targetImageFilename)
			if err != nil {
				log.Warnf("unable to extract coordinates from '%s': %v", targetImageFilename, err)
				continue
			}

			band, err := extractBand(targetImageFilename)
			if err != nil {
				log.Warnf("unable to extract band from '%s': %v", targetImageFilename, err)
				continue
			}

			groupID := extractGroupID(targetImageFilename)

			d3mID := d3mIDs[groupID]
			if d3mID == 0 {
				d3mID = d3mIDRunning
				d3mIDRunning = d3mIDRunning + 1
				d3mIDs[groupID] = d3mID
			}

			csvData = append(csvData, []string{fmt.Sprintf("%d", d3mID), path.Base(targetImageFilename), groupID, band, coordinates.ToString(), label})
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
			[]string{model.RoleMultiIndex}, model.VarDistilRoleIndex, nil, dr.Variables, false),
	)
	dr.Variables = append(dr.Variables,
		model.NewVariable(1, "image_file", "image_file", "image_file", model.StringType,
			model.StringType, "Reference to image file", []string{"attribute"},
			model.VarRoleData, map[string]interface{}{"resID": "0", "resObject": "item"}, dr.Variables, false))
	dr.Variables = append(dr.Variables,
		model.NewVariable(2, "group_id", "group_id", "group_id", model.StringType,
			model.StringType, "Image band", []string{"attribute"},
			model.VarRoleData, nil, dr.Variables, false))
	dr.Variables = append(dr.Variables,
		model.NewVariable(3, "band", "band", "band", model.StringType,
			model.StringType, "Image band", []string{"attribute"},
			model.VarRoleData, nil, dr.Variables, false))
	dr.Variables = append(dr.Variables,
		model.NewVariable(4, "coordinates", "coordinates", "coordinates", model.RealVectorType,
			model.RealVectorType, "Coordinates of the image defined by a bounding box", []string{"attribute"},
			model.VarRoleData, nil, dr.Variables, false))
	dr.Variables = append(dr.Variables,
		model.NewVariable(5, "label", "label", "label", model.StringType,
			model.StringType, "Label of the image", []string{"suggestedTarget"},
			model.VarRoleData, nil, dr.Variables, false))

	// create the data resource for the referenced images
	imageTypeLookup := satTypeMap[s.ImageType]
	refDR := model.NewDataResource("0", model.ResTypeImage, map[string][]string{fmt.Sprintf("image/%s", imageTypeLookup): satTypeContentMap[imageTypeLookup]})
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

func verifySatelliteImage(filename string, defaultType string) bool {
	typ := path.Ext(filename)
	if len(typ) > 0 {
		typ = typ[1:]
	} else {
		typ = defaultType
	}

	return satTypeMap[typ] != ""
}

func extractBand(filename string) (string, error) {
	bandRaw := bandRegex.Find([]byte(filename))
	if len(bandRaw) > 0 {
		band := string(bandRaw)
		return band[2 : len(band)-1], nil
	}

	return "", errors.New("unable to extract band from filename")
}

func extractGroupID(filename string) string {
	bandRaw := bandRegex.Find([]byte(filename))
	adjustedFilename := path.Base(filename)
	if len(bandRaw) > 0 {
		adjustedFilename = strings.Replace(adjustedFilename, string(bandRaw), ".", 1)
	}

	adjustedFilename = strings.TrimSuffix(adjustedFilename, path.Ext(adjustedFilename))

	return adjustedFilename
}

func extractCoordinates(filename string) (*BoundingBox, error) {
	ds, err := gdal.OpenEx(filename, gdal.OFReadOnly, []string{"GTIFF"}, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open geotiff file")
	}

	width := float64(ds.RasterXSize())
	height := float64(ds.RasterYSize())
	gt := ds.GeoTransform()

	minX := gt[0]
	minY := gt[3] + width*gt[4] + height*gt[5]
	maxX := gt[0] + width*gt[1] + height*gt[2]
	maxY := gt[3]

	source := gdal.CreateSpatialReference("")
	err = source.FromWKT(ds.Projection())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create source spatial reference from projection")
	}

	target := gdal.CreateSpatialReference("")
	err = target.FromEPSG(4326)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create source spatial reference EPSG code")
	}

	pointsX := []float64{minX, maxX}
	pointsY := []float64{minY, maxY}
	transform := gdal.CreateCoordinateTransform(source, target)
	success := transform.Transform(len(pointsX), pointsX, pointsY, []float64{0, 0})
	if !success {
		return nil, errors.New("unable to transform points")
	}

	return &BoundingBox{
		LowerLeft: &Point{
			X: pointsX[0],
			Y: pointsY[0],
		},
		UpperLeft: &Point{
			X: pointsX[0],
			Y: pointsY[1],
		},
		UpperRight: &Point{
			X: pointsX[1],
			Y: pointsY[1],
		},
		LowerRight: &Point{
			X: pointsX[1],
			Y: pointsY[0],
		},
	}, nil
}
