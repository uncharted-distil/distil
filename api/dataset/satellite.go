//
//   Copyright © 2021 Uncharted Software Inc.
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
	"path"
	"regexp"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/gdal"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
)

const (
	errorLogLimit = 50
)

var (
	satTypeMap = map[string]string{
		"tif":  "tiff",
		"tiff": "tiff",
	}
	satTypeContentMap = map[string][]string{
		"tiff": {"tif", "tiff"},
	}

	bandRegex      = regexp.MustCompile(`_B[0-9][0-9a-zA-Z][.]`)
	timestampRegex = regexp.MustCompile(`\d{8}T\d{4,6}`)

	// eurosat drops cloud layer, has the 8A layer and offsets everything else.
	eurosatBandMapping = map[int]string{
		10: "",
		13: "8A",
	}
)

// RemoteSensingDatasetProperties lists the data properties of a remote sensing dataset.
type RemoteSensingDatasetProperties struct {
	MultiClass     bool
	MultiTimestamp bool
}

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
func (b *BoundingBox) String() string {
	coords := []string{
		b.pointToString(b.LowerLeft, ","),
		b.pointToString(b.UpperLeft, ","),
		b.pointToString(b.UpperRight, ","),
		b.pointToString(b.LowerRight, ","),
	}
	return strings.Join(coords, ",")
}

// ToGeometryString writes out the bounding box to a geometry string (POSTGIS).
func (b *BoundingBox) ToGeometryString() string {
	coords := []string{
		b.pointToString(b.LowerLeft, " "),
		b.pointToString(b.UpperLeft, " "),
		b.pointToString(b.UpperRight, " "),
		b.pointToString(b.LowerRight, " "),
		b.pointToString(b.LowerLeft, " "),
	}
	return fmt.Sprintf("POLYGON((%s))", strings.Join(coords, ","))
}

func (b *BoundingBox) pointToString(point *Point, separator string) string {
	if point != nil {
		return fmt.Sprintf("%f%s%f", point.X, separator, point.Y)
	}

	return separator
}

// NewSatelliteDataset creates a new satelitte dataset from geotiff files
func NewSatelliteDataset(dataset string, imageType string, rawFilePath string) (*Satellite, error) {
	expandedInfo, err := ExpandZipDataset(rawFilePath, dataset)
	if err != nil {
		return nil, err
	}

	return &Satellite{
		Dataset:           dataset,
		ImageType:         imageType,
		RawFilePath:       expandedInfo.RawFilePath,
		ExtractedFilePath: expandedInfo.ExtractedFilePath,
	}, nil
}

// NewSatelliteDatasetFromExpanded creates a new satelitte dataset from geotiff files where the archive has already been expanded.
func NewSatelliteDatasetFromExpanded(dataset string, imageType string, rawFilePath string, extractedFilePath string) (*Satellite, error) {
	return &Satellite{
		Dataset:           dataset,
		ImageType:         imageType,
		RawFilePath:       rawFilePath,
		ExtractedFilePath: extractedFilePath,
	}, nil
}

// CreateDataset processes the raw satellite dataset and creates a raw D3M dataset.
func (s *Satellite) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*serialization.RawDataset, error) {
	if datasetName == "" {
		datasetName = s.Dataset
	}
	outputDatasetPath := rootDataPath
	dataFilePath := path.Join(outputDatasetPath, compute.D3MDataFolder, compute.D3MLearningData)

	imageFolders, err := getLabelFolders(s.ExtractedFilePath)
	if err != nil {
		return nil, err
	}
	props := s.readProperties(imageFolders)
	labelHeader := "label"
	expectedHeaders := []string{model.D3MIndexFieldName, "image_file", "group_id", "band", "timestamp", "coordinates", "geo_coordinates"}
	if props.MultiClass {
		expectedHeaders = append(expectedHeaders, labelHeader)
	}
	headerNames := append([]string{}, expectedHeaders...)

	csvData := make([][]string, 0)
	csvData = append(csvData, headerNames)
	mediaFolder := util.GetUniqueFolder(path.Join(outputDatasetPath, "media"))

	// need to keep track of d3m Index values since they are shared for a whole group
	d3mIDs := make(map[string]int)
	d3mIDRunning := 1

	// the folder name represents the label to apply for all containing images
	errorCount := 0
	timestampType := model.DateTimeType
	indicesToKeep := getIndicesToKeep(expectedHeaders, headerNames)
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
				logWarning(errorCount, "'%s' is not a valid or supported satellite image", imageFilenameFull)
				errorCount++
				continue
			}

			filesToProcess, err := copyAndSplitMultiBandImage(imageFilenameFull, s.ImageType, mediaFolder)
			if err != nil {
				return nil, err
			}

			for _, targetImageFilename := range filesToProcess {
				coordinates, err := extractCoordinates(targetImageFilename)
				if err != nil {
					logWarning(errorCount, "unable to extract coordinates from '%s': %v", targetImageFilename, err)
					errorCount++
					continue
				}

				band, err := extractBand(targetImageFilename)
				if err != nil {
					logWarning(errorCount, "unable to extract band from '%s': %v", targetImageFilename, err)
					errorCount++
					continue
				}

				timestamp, err := extractTimestamp(targetImageFilename)
				if err != nil {
					logWarning(errorCount, "unable to extract timestamp from '%s': %v", targetImageFilename, err)
					errorCount++
					timestampType = model.StringType
				}

				groupID := extractGroupID(targetImageFilename, props)

				d3mID := d3mIDs[groupID]
				if d3mID == 0 {
					d3mID = d3mIDRunning
					d3mIDRunning = d3mIDRunning + 1
					d3mIDs[groupID] = d3mID
				}
				// remove values that are not needed based on the headerNames (expects values, expectedHeaders and headerNames to be IN ORDER)
				csvLine := removeMissingValues(indicesToKeep, []string{fmt.Sprintf("%d", d3mID), path.Base(targetImageFilename), groupID, band, timestamp, coordinates.String(), coordinates.ToGeometryString(), label})

				csvData = append(csvData, csvLine)
			}
		}
	}
	log.Infof("parsed all input data creating %d rows of data and %d errors", len(csvData)-1, errorCount)

	// create the dataset schema doc
	datasetID := model.NormalizeDatasetID(datasetName)
	meta := model.NewMetadata(datasetName, datasetName, "", datasetID)
	dr := model.NewDataResource(compute.DefaultResourceID, model.ResTypeTable, map[string][]string{compute.D3MResourceFormat: {"csv"}})
	dr.ResPath = dataFilePath
	varCounter := 0
	dr.Variables = append(dr.Variables,
		model.NewVariable(varCounter, model.D3MIndexFieldName, model.D3MIndexFieldName, model.D3MIndexFieldName,
			model.D3MIndexFieldName, model.IntegerType, model.IntegerType, "D3M index",
			[]string{model.RoleMultiIndex}, model.VarDistilRoleIndex, nil, dr.Variables, false),
	)
	varCounter++
	dr.Variables = append(dr.Variables,
		model.NewVariable(varCounter, "image_file", "image_file", "image_file", "image_file", model.MultiBandImageType,
			model.StringType, "Reference to image file", []string{"attribute"},
			model.VarDistilRoleData, map[string]interface{}{"resID": "0", "resObject": "item"}, dr.Variables, false))
	varCounter++
	dr.Variables = append(dr.Variables,
		model.NewVariable(varCounter, "group_id", "group_id", "group_id", "group_id", model.StringType,
			model.StringType, "ID linking all bands of a particular image set together", []string{"attribute", "suggestedGroupingKey"},
			model.VarDistilRoleGrouping, nil, dr.Variables, false))
	varCounter++
	dr.Variables = append(dr.Variables,
		model.NewVariable(varCounter, "band", "band", "band", "band", model.StringType,
			model.StringType, "Image band", []string{"attribute"},
			model.VarDistilRoleData, nil, dr.Variables, false))
	varCounter++
	dr.Variables = append(dr.Variables,
		model.NewVariable(varCounter, "timestamp", "timestamp", "timestamp", "timestamp", timestampType,
			model.StringType, "Image timestamp", []string{"attribute"},
			model.VarDistilRoleData, nil, dr.Variables, false))
	varCounter++
	dr.Variables = append(dr.Variables,
		model.NewVariable(varCounter, "coordinates", "coordinates", "coordinates", "coordinates", model.RealVectorType,
			model.RealVectorType, "Coordinates of the image defined by a bounding box", []string{"attribute"},
			model.VarDistilRoleData, nil, dr.Variables, false))
	varCounter++
	dr.Variables = append(dr.Variables,
		model.NewVariable(varCounter, "__geo_coordinates", "coordinates", "geo_coordinates", "__geo_coordinates", model.GeoBoundsType,
			model.GeoBoundsType, "postgis structure for the bounding box coordinates of the tile", []string{},
			model.VarDistilRoleMetadata, nil, dr.Variables, false))
	varCounter++
	if len(expectedHeaders) == len(headerNames) {
		dr.Variables = append(dr.Variables,
			model.NewVariable(varCounter, "label", "label", "label", "label", model.CategoricalType,
				model.StringType, "Label of the image", []string{"suggestedTarget"},
				model.VarDistilRoleData, nil, dr.Variables, false))
	}

	// create the data resource for the referenced images
	imageTypeLookup := satTypeMap[s.ImageType]
	refDR := model.NewDataResource("0", model.ResTypeImage, map[string][]string{fmt.Sprintf("image/%s", imageTypeLookup): satTypeContentMap[imageTypeLookup]})
	refDR.ResPath = mediaFolder
	refDR.IsCollection = true

	meta.DataResources = []*model.DataResource{refDR, dr}

	return &serialization.RawDataset{
		ID:              datasetID,
		Name:            datasetName,
		Data:            csvData,
		Metadata:        meta,
		DefinitiveTypes: true,
	}, nil
}

func (s *Satellite) readProperties(imageFolders []string) *RemoteSensingDatasetProperties {
	// cycle through folders to determine the dataset properties
	props := &RemoteSensingDatasetProperties{
		MultiClass: len(imageFolders) > 1,
	}

	// cycle through image names to determine if there are at least 2 timestamps
	firstTimestamp := ""
	for _, imageFolder := range imageFolders {
		imageFiles, err := ioutil.ReadDir(imageFolder)
		if err != nil {
			break
		}

		for _, imageFile := range imageFiles {
			timestamp, _ := extractTimestamp(imageFile.Name())
			if firstTimestamp == "" {
				firstTimestamp = timestamp
			} else if firstTimestamp != timestamp {
				props.MultiTimestamp = true
				break
			}
		}
	}

	return props
}

// removeValues removes values not needed based on supplied headernames
func removeMissingValues(indices []int, values []string) []string {
	result := []string{}
	// build the value string based on the actual headers that exist
	for _, idx := range indices {
		result = append(result, values[idx])
	}
	return result
}
func getIndicesToKeep(expectedHeaders []string, headers []string) []int {
	result := []int{}
	headerMap := map[string]int{}
	// build map
	for i, header := range headers {
		headerMap[header] = i
	}
	// check what is missing and append indices
	for i, header := range expectedHeaders {
		if _, ok := headerMap[header]; ok {
			result = append(result, i)
		}
	}
	return result
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
		return strings.ToLower(band[2 : len(band)-1]), nil
	}

	return "", errors.New("unable to extract band from filename")
}

func extractTimestamp(filename string) (string, error) {
	timestampRaw := timestampRegex.Find([]byte(filename))
	if len(timestampRaw) == 0 {
		return "", errors.New("unable to extract band from filename")
	}

	parsed, err := dateparse.ParseAny(strings.Replace(string(timestampRaw), "T", "", -1))
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse timestamp")
	}

	return parsed.Format("2006-01-02 03:04:05"), nil
}

func extractGroupID(filename string, props *RemoteSensingDatasetProperties) string {
	bandRaw := bandRegex.Find([]byte(filename))
	adjustedFilename := path.Base(filename)
	if len(bandRaw) > 0 {
		adjustedFilename = strings.Replace(adjustedFilename, string(bandRaw), ".", 1)
	}

	// remove the timestamp from the group id if there is only one timestamp.
	if !props.MultiTimestamp {
		timestampRaw := timestampRegex.Find([]byte(adjustedFilename))
		if len(timestampRaw) > 0 {
			timestampString := fmt.Sprintf("_%s", string(timestampRaw))
			adjustedFilename = strings.Replace(adjustedFilename, timestampString, "", 1)
		}
	}

	adjustedFilename = strings.TrimSuffix(adjustedFilename, path.Ext(adjustedFilename))

	return adjustedFilename
}

func extractCoordinates(filename string) (*BoundingBox, error) {
	ds, err := gdal.OpenEx(filename, gdal.OFReadOnly, []string{"GTIFF"}, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open geotiff file")
	}
	defer ds.Close()

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

	defer transform.Destroy()

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

func logWarning(currentCount int, warning string, params ...interface{}) {
	if currentCount < errorLogLimit {
		log.Warnf(warning, params...)
	} else if currentCount == errorLogLimit {
		log.Warnf("reached error log limit (%d) so no further parsing errors will be logged", errorLogLimit)
	}
}

func copyAndSplitMultiBandImage(imageFilename string, imageType string, outputFolder string) ([]string, error) {
	files := make([]string, 0)

	// open file
	dataset, err := gdal.Open(imageFilename, gdal.ReadOnly)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to load geotiff")
	}
	defer dataset.Close()

	// check number of bands
	bandCount := dataset.RasterCount()

	if bandCount == 1 {
		// only one band means a simple copy of the file
		targetImageFilename := path.Base(imageFilename)
		extension := path.Ext(targetImageFilename)
		if extension != fmt.Sprintf(".%s", imageType) {
			targetImageFilename = fmt.Sprintf("%s.%s", strings.TrimSuffix(targetImageFilename, extension), imageType)
		}
		targetImageFilename = util.GetUniqueName(path.Join(outputFolder, targetImageFilename))

		err := util.CopyFile(imageFilename, targetImageFilename)
		if err != nil {
			return nil, errors.Wrap(err, "unable to copy image file")
		}
		files = append(files, targetImageFilename)
	} else {
		// multiband so need to split it into separate files
		files, err = util.SplitMultiBandImage(dataset, outputFolder, eurosatBandMapping)
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// CreateSatelliteGrouping dumps the satellite grouping structure into a map.
// It assumes that the dataset has the same structure as during upload.
func CreateSatelliteGrouping() map[string]interface{} {
	// assume dataset structure matches what would be created during ingest
	grouping := map[string]interface{}{}
	grouping["bandCol"] = "band"
	grouping["idCol"] = "group_id"
	grouping["imageCol"] = "image_file"
	grouping["type"] = model.MultiBandImageType
	grouping["hidden"] = []string{"image_file", "band", "group_id"}

	return grouping
}

// CreateGeoBoundsGrouping dumps the geobounds grouping structure into a map.
// It assumes that the dataset has the same structure as during upload.
func CreateGeoBoundsGrouping() map[string]interface{} {
	grouping := map[string]interface{}{}
	grouping["type"] = model.GeoBoundsType
	grouping["coordinatesCol"] = "coordinates"
	grouping["polygonCol"] = "__geo_coordinates"
	grouping["hidden"] = []string{"coordinates", "__geo_coordinates"}

	return grouping
}

// GetSatelliteIndexFields lists the fields to index for satellite datasets.
func GetSatelliteIndexFields() []string {
	return []string{"band", "__geo_coordinates"}
}
