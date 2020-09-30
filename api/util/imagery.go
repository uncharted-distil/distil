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

package util

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strings"

	lru "github.com/hashicorp/golang-lru"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/gdal"
	log "github.com/unchartedsoftware/plog"
)

const (
	// Sentinel2Max is the maximum expected value stored in a sentinel 2 satellite band.  Spec indicates a 12 bit
	// value.
	Sentinel2Max = 10000

	// Exponent is the exponent to apply to channel values during pre-processing.  A value of 1.0 will leave values
	// unchanged.
	Exponent = 0.3

	// NaturalColors identifies a band mapping that displays an image in natural color.
	NaturalColors = "natural_colors"

	// FalseColorInfrared identifies a band mapping that displays an image in false color for visualizing vegatation.
	FalseColorInfrared = "false_color_infrared"

	// FalseColorUrban identifies a band mapping that displays an image in false color for visualizing urban development.
	FalseColorUrban = "false_color_urban"

	// Agriculture identifies a band mapping that displays an image in false color for visualization agricultural activity.
	Agriculture = "agriculture"

	// AtmosphericPenetration identifies a band mapping that displays an image in false for visualizing atmospheric penetration.
	AtmosphericPenetration = "atmospheric_penetration"

	// HealthyVegetation identifies a band mapping that displays an image in false color for visualizing vegatation health.
	HealthyVegetation = "healthy_vegetation"

	// LandWater identifies a band mapping that displays an image in in false color that separates land and water.
	LandWater = "land_water"

	// AtmosphericRemoval identifies a band mapping that displays an image in near true color with atmoshperic effects reduced.
	AtmosphericRemoval = "atmospheric_removal"

	// ShortwaveInfrared identifies a band mapping that displays an image in shortwave infrared.
	ShortwaveInfrared = "shortwave_infrared"

	// VegetationAnalysis identifies a band mapping that displays an image in in false color for analyzing vegetation.
	VegetationAnalysis = "vegetation_analysis"

	// NDVI identifies a band mapping that displays Normalized Difference Vegetation Index mapped using an RGB ramp.
	NDVI = "ndvi"

	// NDWI identifies a band mapping that displays Normalized Difference Water Index mapped using an RGB ramp.
	NDWI = "ndwi"

	// NDMI idenfifies a band mapping that display Normalized Difference Moisture Index mapped using an RGB ramp.
	NDMI = "ndmi"
)

// BandCombinationID uniquely identifies a band combination
type BandCombinationID string

// BandCombination defines a mapping of satellite bands to image RGB channels.
type BandCombination struct {
	ID          BandCombinationID
	DisplayName string
	Mapping     []string
	Ramp        []uint8
	Transform   func(...uint16) float64
}

// ClampedNormalizingTransform transforms to a range of (-1, 1) and then clamps to (0, 1)
func ClampedNormalizingTransform(bandValues ...uint16) float64 {
	return math.Max(0, float64(int32(bandValues[0])-int32(bandValues[1]))/float64(int32(bandValues[0])+int32(bandValues[1])))
}

// NormalizingTransform transforms to a range of (-1, 1) and then normalizes to (0, 1)
func NormalizingTransform(bandValues ...uint16) float64 {
	return (1.0 + float64(int32(bandValues[0])-int32(bandValues[1]))/float64(int32(bandValues[0])+int32(bandValues[1]))) / 2.0
}

var (
	// SentinelBandCombinations defines a list of recommended band combinations for sentinel 2 satellite missions
	SentinelBandCombinations = map[string]*BandCombination{}

	// Cache to hold directory file type search results
	folderTypeCache *lru.Cache
)

func init() {
	// create an LRU cache to hold the results of time consuming directory content analysis
	var err error
	folderTypeCache, err = lru.New(100)
	if err != nil {
		log.Error(errors.Wrap(err, "failed to init directory type cache"))
	}

	// initialize the band combination structures - needs to be done in init so that referenced color ramps are built
	SentinelBandCombinations = map[string]*BandCombination{
		NaturalColors:          {NaturalColors, "Natural Colors", []string{"b04", "b03", "b02"}, nil, nil},
		FalseColorInfrared:     {FalseColorInfrared, "False Color Infrared", []string{"b08", "b04", "b03"}, nil, nil},
		FalseColorUrban:        {FalseColorUrban, "False Color Urban", []string{"b12", "b11", "b04"}, nil, nil},
		Agriculture:            {Agriculture, "Agriculture", []string{"b11", "b08", "b02"}, nil, nil},
		AtmosphericPenetration: {AtmosphericPenetration, "Atmospheric Penetration", []string{"b12", "b11", "b8A"}, nil, nil},
		HealthyVegetation:      {HealthyVegetation, "Healthy Vegetation", []string{"b08", "b11", "b02"}, nil, nil},
		LandWater:              {LandWater, "Land/Water", []string{"b08", "b11", "b04"}, nil, nil},
		AtmosphericRemoval:     {AtmosphericRemoval, "Atmospheric Removal", []string{"b12", "b08", "b03"}, nil, nil},
		ShortwaveInfrared:      {ShortwaveInfrared, "Shortwave Infrared", []string{"b12", "b08", "b04"}, nil, nil},
		VegetationAnalysis:     {VegetationAnalysis, "Vegetation Analysis", []string{"b11", "b08", "b04"}, nil, nil},
		NDVI:                   {NDVI, "Normalized Difference Vegetation Index", []string{"b08", "b04"}, RedYellowGreenRamp, ClampedNormalizingTransform},
		NDMI:                   {NDMI, "Normalized Difference Moisture Index ", []string{"b08", "b11"}, BlueYellowBrownRamp, NormalizingTransform},
		NDWI:                   {NDWI, "Normalized Difference Water Index", []string{"b03", "b08"}, BlueYellowBrownRamp, NormalizingTransform},
	}
}

// ImageFromCombination takes a base datsaet directory, fileID and a band combination label and
// returns a composed image.  NOTE: Currently a bit hardcoded for sentinel-2 data.
func ImageFromCombination(datasetDir string, fileID string, bandCombination BandCombinationID) (*image.RGBA, error) {
	// attempt to get the folder file type for the supplied dataset dir from the cache, if
	// not do the look up
	var fileType string
	cacheValue, ok := folderTypeCache.Get(datasetDir)
	if !ok {
		var err error
		fileType, err = GetFolderFileType(datasetDir)
		if err != nil {
			return nil, err
		}
		folderTypeCache.Add(datasetDir, fileType)
	} else {
		fileType = cacheValue.(string)
	}

	// map the band files to the inputs
	filePaths := []string{}
	if bandCombo, ok := SentinelBandCombinations[strings.ToLower(string(bandCombination))]; ok {
		for _, bandLabel := range bandCombo.Mapping {
			filePath := getFilePath(datasetDir, fileID, bandLabel, fileType)
			filePaths = append(filePaths, filePath)
		}
		return ImageFromBands(filePaths, bandCombo.Ramp, bandCombo.Transform)
	}

	return nil, errors.Errorf("unhandled band combination %s", bandCombination)
}

// ImageFromBands loads band data from the file paths array into a single RGB image,
// where the file names map to R,G,B in order.  The results are returned as a JPEG
// encoded byte stream. If errors are encountered processing a band an attempt will
// be made to create the image from the remaining bands, while logging an error.
func ImageFromBands(paths []string, ramp []uint8, transform func(...uint16) float64) (*image.RGBA, error) {
	bandImages := []*image.Gray16{}
	maxXSize := 0
	maxYSize := 0

	// Load each of the datasets as a Gray16 image
	for _, filePath := range paths {
		bandImage, err := loadAsGray16(filePath)
		bandImages = append(bandImages, bandImage)
		if err != nil {
			log.Error(err, "band file not loaded")
			continue
		}

		// extract input raster size and update max x,y
		xSize := bandImage.Bounds().Dx()
		ySize := bandImage.Bounds().Dy()
		if xSize > maxXSize {
			maxXSize = xSize
		}
		if ySize > maxYSize {
			maxYSize = ySize
		}
	}

	// Resize any images that are below the max size
	for i, bandImage := range bandImages {
		if bandImage != nil && (bandImage.Bounds().Dx() < maxXSize || bandImage.Bounds().Dy() < maxYSize) {
			// no need to check type assertion - guaranteed to be what as passed in by api
			bandImages[i] = resize.Resize(uint(maxXSize), uint(maxYSize), bandImage, resize.NearestNeighbor).(*image.Gray16)
		}
	}

	// Ceate the final image either as a direct mapping from the supplied bands, or by applying
	// a transform and color lookup
	if ramp == nil || transform == nil {
		// Create an RGBA image from the resized bands
		return createRGBAFromBands(maxXSize, maxYSize, bandImages), nil
	}
	return createRGBAFromRamp(maxXSize, maxYSize, bandImages, transform, ramp), nil
}

func loadAsGray16(filePath string) (*image.Gray16, error) {
	// Load each of the datasets
	dataset, err := gdal.Open(filePath, gdal.ReadOnly)
	if err != nil {
		return nil, errors.Wrap(err, "band file not loaded")
	}

	// Accept a single band.
	numBands := dataset.RasterCount()
	if numBands == 0 {
		log.Warnf("found 0 bands - skipping")
	} else if numBands > 1 {
		log.Warnf("found %d bands - using band 0 only", numBands)
	}
	inputBand0 := dataset.RasterBand(1)

	// extract input raster size and update max x,y
	xSize := dataset.RasterXSize()
	ySize := dataset.RasterYSize()

	// extract input band data type
	dataType := inputBand0.RasterDataType()

	// Accept 16 bit integer data
	var bandImage *image.Gray16
	if dataType == gdal.UInt16 {
		bandImage = image.NewGray16(image.Rect(0, 0, xSize, ySize))
	} else {
		log.Warnf("unhandled data type %s - skipping", dataType.Name())
		dataset.Close()
		return nil, nil
	}

	// read the band data into the image buffer
	buffer := make([]uint16, xSize*ySize)
	if err = inputBand0.IO(gdal.Read, 0, 0, xSize, ySize, buffer, xSize, ySize, 0, 0); err != nil {
		dataset.Close()
		return nil, errors.Wrapf(err, "failed to load band data for %s", filePath)
	}
	dataset.Close() // done with GDAL buffer

	// crappy for now - go image lib stores its gray16 as [uint8, uint8] so we need an extra copy here
	badCount := 0
	for i, grayVal := range buffer {
		if grayVal > Sentinel2Max {
			grayVal = Sentinel2Max
			badCount++
		}
		// decompose the 16-bit value into 8 bit values with a big endian ordering as per the image lib
		// documentation
		bandImage.Pix[i*2] = uint8(grayVal & 0xFF00 >> 8)
		bandImage.Pix[i*2+1] = uint8(grayVal & 0xFF)
	}
	if badCount > 0 {
		log.Warnf("truncated %d values from %s", badCount, filePath)
	}

	return bandImage, nil
}

func createRGBAFromBands(xSize int, ySize int, bandImages []*image.Gray16) *image.RGBA {
	// Create a new RGBA image to hold the collected bands
	outputImage := image.NewRGBA(image.Rect(0, 0, xSize, ySize))

	// Copy the 16 bit band images into the 8 bit target image.  If a band image couldn't be processed
	// earlier, we set to grey.
	outputIdx := 0
	for i := 0; i < (xSize * ySize * 2); i += 2 {
		for _, bandImage := range bandImages {
			if bandImage != nil {
				grayValue16 := uint16(bandImage.Pix[i])<<8 | uint16(bandImage.Pix[i+1])
				outputImage.Pix[outputIdx] = uint8(math.Pow(float64(grayValue16)/Sentinel2Max, Exponent) * 255)
			} else {
				outputImage.Pix[outputIdx] = uint8(math.MaxInt8 / 2)
			}
			outputIdx++
		}
		outputImage.Pix[outputIdx] = 0xFF // max out the A channel
		outputIdx++
	}
	return outputImage
}

func createRGBAFromRamp(xSize int, ySize int, bandImages []*image.Gray16, transform func(...uint16) float64, ramp []uint8) *image.RGBA {
	// Create a new RGBA image to hold the collected bands
	outputImage := image.NewRGBA(image.Rect(0, 0, xSize, ySize))

	rampElements := len(ramp) / 3

	// Copy the 16 bit band images into the 8 bit target image.  If a band image couldn't be processed
	// earlier, we set to grey.
	outputIdx := 0
	bandImage0 := bandImages[0]
	bandImage1 := bandImages[1]
	for i := 0; i < (xSize * ySize * 2); i += 2 {
		// extract the 16 bit pixel values for each input band
		grayValue0 := uint16(bandImage0.Pix[i])<<8 | uint16(bandImage0.Pix[i+1])
		grayValue1 := uint16(bandImage1.Pix[i])<<8 | uint16(bandImage1.Pix[i+1])

		// compute NDVI ratio
		transformedValue := -1.0
		if grayValue0 != 0 || grayValue1 != 0 {
			transformedValue = transform(grayValue0, grayValue1)
		}
		pixelOffset := int(transformedValue * float64(rampElements))

		outputImage.Pix[outputIdx] = uint8(ramp[pixelOffset*3])
		outputImage.Pix[outputIdx+1] = uint8(ramp[pixelOffset*3+1])
		outputImage.Pix[outputIdx+2] = uint8(ramp[pixelOffset*3+2])
		outputImage.Pix[outputIdx+3] = 0xFF // max out the A channel
		outputIdx += 4
	}
	return outputImage
}

// LoadPNGImage loads an RGBA PNG from the caller supplied path, decodes it,
// and returns it as an RGBA image.  Return an error if the image is not RGBA.
func LoadPNGImage(filename string) (*image.RGBA, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load image")
	}
	defer file.Close()

	imageData, err := png.Decode(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode image")
	}

	rgbaImage, ok := imageData.(*image.RGBA)
	if !ok {
		return nil, errors.Errorf("image type %T is not RGBA", imageData)
	}
	return rgbaImage, nil
}

// SavePNGImage saves an RGBA image to disk in PNG format.
func SavePNGImage(image *image.RGBA, filename string) error {
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, image); err != nil {
		return errors.Wrap(err, "failed so encode png file")
	}
	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "failed to write png file to disk")
	}
	return nil
}

// ImageToJPEG encodes an RGBA image as a JPEG byte array for further processing or
// network transmission.
func ImageToJPEG(image image.Image) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, image, nil); err != nil {
		return nil, errors.Wrap(err, "failed so encode png file")
	}
	return buf.Bytes(), nil
}

// SplitMultiBandImage splits a multiband image into separate images, each
// being for a single band. Bands can be mapped and dropped.
func SplitMultiBandImage(dataset gdal.Dataset, outputFolder string, bandMapping map[int]string) ([]string, error) {
	// make the output folder
	if err := os.MkdirAll(outputFolder, os.ModePerm); err != nil {
		return nil, errors.Wrap(err, "failed to create dir for multiband image split")
	}
	filename := dataset.FileList()[0]
	tileName := path.Base(filename)
	tileName = strings.TrimSuffix(tileName, path.Ext(tileName))

	files := make([]string, 0)
	for band := 1; band <= dataset.RasterCount(); band++ {
		mappedBand, ok := bandMapping[band]
		if ok && mappedBand == "" {
			continue
		} else if !ok {
			mappedBand = fmt.Sprintf("%02d", band)
		}

		name := fmt.Sprintf("%s_B%s.tiff", tileName, mappedBand)
		fullName := path.Join(outputFolder, name)
		dst := gdal.GDALTranslate(fullName, dataset, []string{"-b", fmt.Sprintf("%d", band)})
		dst.Close()
		files = append(files, fullName)
	}

	return files, nil
}

// getFilePath takes a top level dataset directory, a file ID and a band label and composes them
// into a coherent path for a BigEarthNet file.
func getFilePath(datasetDir string, fileID string, bandLabel string, fileType string) string {
	fileName := fmt.Sprintf("%s_%s.%s", fileID, strings.ToUpper(bandLabel), fileType)
	return path.Join(datasetDir, fileName)
}
