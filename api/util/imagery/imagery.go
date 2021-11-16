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

package imagery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"path"
	"sort"
	"strings"

	lru "github.com/hashicorp/golang-lru"
	"github.com/mitchellh/hashstructure"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/gdal"
	log "github.com/unchartedsoftware/plog"
)

const (
	// Sentinel2Max is the maximum expected value stored in a sentinel 2 satellite band.  Spec indicates a 12 bit
	// value.
	Sentinel2Max = 10000

	// NaturalColors1 identifies a band mapping that displays an image in natural color.
	NaturalColors1 = "natural_colors_1"

	// NaturalColors2 identifies a band mapping that displays an image in natural color.
	NaturalColors2 = "natural_colors_2"

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

	// Segmentation identifies a placeholder band mapping to display image segmentation output.
	Segmentation = "segmentation"

	// ShortwaveInfrared identifies a band mapping that displays an image in shortwave infrared.
	ShortwaveInfrared = "shortwave_infrared"

	// VegetationAnalysis identifies a band mapping that displays an image in in false color for analyzing vegetation.
	VegetationAnalysis = "vegetation_analysis"

	// ImageAttention identifies what the model is paying attention to in the image
	ImageAttention = "image_attention"
	// NDVI identifies a band mapping that displays Normalized Difference Vegetation Index mapped using an RGB ramp.
	NDVI = "ndvi"

	// NDWI identifies a band mapping that displays Normalized Difference Water Index mapped using an RGB ramp.
	NDWI = "ndwi"

	// NDMI identifies a band mapping that displays Normalized Difference Moisture Index mapped using an RGB ramp.
	NDMI = "ndmi"

	// NSMI identifies a band mapping that display Normalized Soil Moisture Index mapped using an RGB ramp.
	NSMI = "nsmi"

	// MNDWI identifies a band mapping that displays Modified Normalized Difference Moisture Index mapped using an RGB ramp
	MNDWI = "mndwi"

	// RSWIR identifies a band mapping that displays Red and Shortwave Infrared mapped using an RGB ramp
	RSWIR = "rswir"
	// OPTRAM identifies a band mapping that represents the soil moisture
	OPTRAM = "optram"
	// OPTRAMJSONFile contains the file name to look for in the learning folder to check if optram is supported
	OPTRAMJSONFile = "optram_variables.json"
)

var (
	cache *lru.Cache
)

// BandCombinationID uniquely identifies a band combination
type BandCombinationID string

// BandCombination defines a mapping of satellite bands to image RGB channels.
type BandCombination struct {
	ID                 BandCombinationID
	DisplayName        string
	Mapping            []string
	Ramp               []uint8
	Transform          func(*OptramEdges, ...uint16) float64
	AdvancedColorModel bool
}

// ImageScale defines what to scale the image size to. If one property is defined aspect ratio will be kept. If nil for both the func will determine the size.
type ImageScale struct {
	Width  int
	Height int
}

//OptramEdges is the struct that contains the edges needed for the optram calculation
type OptramEdges struct {
	IDryEdge float64 `json:"i_d"`
	SDryEdge float64 `json:"s_d"`
	IWetEdge float64 `json:"i_w"`
	SWetEdge float64 `json:"s_w"`
}

// ReadOptramFile will parse the optram json file into a map[string]OptramEdges struct and return the precision
func ReadOptramFile(file string) (map[string]OptramEdges, int, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, 0, err
	}
	result := map[string]OptramEdges{}
	err = json.Unmarshal(buffer, &result)
	if err != nil {
		return nil, 0, err
	}
	precision := 0
	for key := range result {
		precision = len(key)
		break
	}
	return result, precision, nil
}
func (imageScale *ImageScale) shouldScale() bool {
	return imageScale.Width != 0 && imageScale.Height != 0
}

// ClampedNormalizingTransform transforms to a range of (-1, 1) and then clamps to (0, 1)
func ClampedNormalizingTransform(edges *OptramEdges, bandValues ...uint16) float64 {
	return math.Max(0, float64(int32(bandValues[0])-int32(bandValues[1]))/float64(int32(bandValues[0])+int32(bandValues[1])))
}

// NormalizingTransform transforms to a range of (-1, 1) and then normalizes to (0, 1)
func NormalizingTransform(edges *OptramEdges, bandValues ...uint16) float64 {
	return (1.0 + float64(int32(bandValues[0])-int32(bandValues[1]))/float64(int32(bandValues[0])+int32(bandValues[1]))) / 2.0
}

// OptramTransform calculates a moisture normalized value between (0, 1)
func OptramTransform(edges *OptramEdges, bandValues ...uint16) float64 {
	b08 := float64(bandValues[0])
	b04 := float64(bandValues[1])
	b12 := float64(bandValues[2])
	NDVI := (b08 - b04) / (b08 + b04)
	STR := (math.Pow((1.0-b12), 2.0) / (2.0 * b12))
	numerator := edges.IDryEdge + edges.SDryEdge*NDVI - STR
	denominator := edges.IDryEdge - edges.IWetEdge + (edges.IDryEdge-edges.SWetEdge)*NDVI
	return math.Max(0.0, math.Min((numerator/denominator)/float64(math.MaxUint8), 1.0))
}

var (
	// SentinelBandCombinations defines a list of recommended band combinations for sentinel 2 satellite missions
	SentinelBandCombinations = map[string]*BandCombination{}
)

func init() {
	config, _ := env.LoadConfig()
	// initialize the band combination structures - needs to be done in init so that referenced color ramps are built
	SentinelBandCombinations = map[string]*BandCombination{
		NaturalColors1:         {NaturalColors1, "Natural Colors", []string{"b04", "b03", "b02"}, nil, nil, false},
		NaturalColors2:         {NaturalColors2, "Natural Colors 2", []string{"b04", "b03", "b02"}, nil, nil, true},
		FalseColorInfrared:     {FalseColorInfrared, "False Color Infrared", []string{"b08", "b04", "b03"}, nil, nil, false},
		FalseColorUrban:        {FalseColorUrban, "False Color Urban", []string{"b12", "b11", "b04"}, nil, nil, false},
		Agriculture:            {Agriculture, "Agriculture", []string{"b11", "b08", "b02"}, nil, nil, false},
		AtmosphericPenetration: {AtmosphericPenetration, "Atmospheric Penetration", []string{"b12", "b11", "b8a"}, nil, nil, false},
		HealthyVegetation:      {HealthyVegetation, "Healthy Vegetation", []string{"b08", "b11", "b02"}, nil, nil, false},
		LandWater:              {LandWater, "Land/Water", []string{"b08", "b11", "b04"}, nil, nil, false},
		AtmosphericRemoval:     {AtmosphericRemoval, "Atmospheric Removal", []string{"b12", "b08", "b03"}, nil, nil, false},
		ShortwaveInfrared:      {ShortwaveInfrared, "Shortwave Infrared", []string{"b12", "b08", "b04"}, nil, nil, false},
		VegetationAnalysis:     {VegetationAnalysis, "Vegetation Analysis", []string{"b11", "b08", "b04"}, nil, nil, false},
		NDVI:                   {NDVI, "Normalized Difference Vegetation Index", []string{"b08", "b04"}, RedYellowGreenRamp, ClampedNormalizingTransform, false},
		NDMI:                   {NDMI, "Normalized Difference Moisture Index ", []string{"b08", "b11"}, BrownYellowBlueRamp, NormalizingTransform, false},
		NDWI:                   {NDWI, "Normalized Difference Water Index", []string{"b03", "b08"}, BrownYellowBlueRamp, NormalizingTransform, false},
		NSMI:                   {NSMI, "Normalized Soil Moisture Index", []string{"b11", "b12"}, BrownYellowBlueRamp, NormalizingTransform, false},
		MNDWI:                  {MNDWI, "Modified Normalized Difference Water Index", []string{"b03", "b11"}, BrownYellowBlueRamp, NormalizingTransform, false},
		RSWIR:                  {RSWIR, "Red and Shortwave Infrared", []string{"b04", "b11"}, BrownYellowBlueRamp, NormalizingTransform, false},
		OPTRAM:                 {OPTRAM, "OPTRAM", []string{"b08", "b04", "b12"}, RedYellowGreenRamp, OptramTransform, false},
	}

	if config.SegmentationEnabled {
		SentinelBandCombinations[Segmentation] = &BandCombination{Segmentation, "Segmentation", []string{}, nil, nil, false}
	}
}

// Initialize sets up the necessary structures for imagery processing.
func Initialize(config *env.Config) {
	log.Infof("initializing imagery utils...")
	cache, _ = lru.New(config.MultiBandImageCacheSize)
	log.Infof("imagery utils initialized")
}

// ImageCacheKey stores the fields used to generate an image hash key.
type ImageCacheKey struct {
	DatasetDir      string
	BandCombination string
	ImageScale      *ImageScale
	Ramp            string
	Options         []Options
	BandsMapped     []string
}

// ImageFromCombination takes a base dataset directory, fileID and a band combination label and
// returns a composed image.  NOTE: Currently a bit hardcoded for sentinel-2 data.
func ImageFromCombination(datasetDir string, bandFileMapping map[string]string, bandCombo string, imageScale ImageScale, edges *OptramEdges, ramp string, options ...Options) (*image.RGBA, error) {
	// attempt to get the folder file type for the supplied dataset dir from the cache, if
	// not do the look up
	bandCombination := strings.ToLower(string(BandCombinationID(bandCombo)))

	keyStruct := ImageCacheKey{
		DatasetDir:      datasetDir,
		BandCombination: bandCombination,
		ImageScale:      &imageScale,
		Ramp:            ramp,
		Options:         options,
	}

	var cacheKey uint64
	var err error

	if cache != nil {
		// need a constant ordering for the band mapping to prevent incorrect cache lookups
		bandsMapped := []string{}
		for bk, bf := range bandFileMapping {
			bandsMapped = append(bandsMapped, fmt.Sprintf("%s:%s", bk, bf))
		}
		sort.Strings(bandsMapped)
		keyStruct.BandsMapped = bandsMapped
		cacheKey, err = hashstructure.Hash(keyStruct, nil)
		if err != nil {
			return nil, err
		}

		if cache.Contains(cacheKey) {
			cached, ok := cache.Get(cacheKey)
			if ok {
				return cached.(*image.RGBA), nil
			}
		}
	}

	// map the band files to the inputs
	filePaths := []string{}
	if bandCombo, ok := SentinelBandCombinations[bandCombination]; ok {
		for _, bandLabel := range bandCombo.Mapping {
			filePaths = append(filePaths, path.Join(datasetDir, bandFileMapping[bandLabel]))
		}

		imageRamp := bandCombo.Ramp

		if ramp != "" {
			imageRamp = GetColorRamp(ramp)
		}

		image, err := ImageFromBands(filePaths, imageRamp, bandCombo.Transform, imageScale, edges, bandCombo.AdvancedColorModel, options...)

		if err != nil {
			return nil, err
		}
		if cache != nil {
			cache.Add(cacheKey, image)
		}

		return image, nil
	}

	return nil, errors.Errorf("unhandled band combination %s", bandCombination)
}

// ImageFromBands loads band data from the file paths array into a single RGB image,
// where the file names map to R,G,B in order.  The results are returned as a JPEG
// encoded byte stream. If errors are encountered processing a band an attempt will
// be made to create the image from the remaining bands, while logging an error.
func ImageFromBands(paths []string, ramp []uint8, transform func(*OptramEdges, ...uint16) float64, imageScale ImageScale, edges *OptramEdges, advancedColorModel bool, options ...Options) (*image.RGBA, error) {
	bandImages := []*image.Gray16{}
	maxXSize := 0
	maxYSize := 0

	for _, filePath := range paths {
		bandImage, err := loadAsGray16(filePath)
		bandImages = append(bandImages, bandImage)
		if err != nil {
			return nil, err
		}
	}
	if imageScale.shouldScale() {
		width, height := getMaxDimensions(&bandImages)
		aspectRatio := float64(height) / float64(width)
		maxXSize = imageScale.Width
		maxYSize = int(float64(maxXSize) * aspectRatio)
	} else {
		maxXSize, maxYSize = getMaxDimensions(&bandImages)
	}
	// Resize any images that are below the max size
	for i, bandImage := range bandImages {
		if bandImage != nil && (imageScale.shouldScale() || bandImage.Bounds().Dx() < maxXSize || bandImage.Bounds().Dy() < maxYSize) {
			// no need to check type assertion - guaranteed to be what as passed in by api
			bandImages[i] = resize.Resize(uint(maxXSize), uint(maxYSize), bandImage, resize.NearestNeighbor).(*image.Gray16)
		}
	}

	// Ceate the final image either as a direct mapping from the supplied bands, or by applying
	// a transform and color lookup
	if ramp == nil || transform == nil {
		// Create an RGBA image from the resized bands
		return createRGBAFromBands(maxXSize, maxYSize, bandImages, advancedColorModel, options...), nil
	}
	return createRGBAFromRamp(maxXSize, maxYSize, bandImages, transform, ramp, edges), nil
}

// ScaleConfidenceMatrix scales confidence matrix to desired size using linear scaling
func ScaleConfidenceMatrix(width int, height int, confidence *[][]float64) [][]float64 {
	resultMatrix := make([][]float64, height)
	confWidth := len((*confidence)[0]) - 1
	adjustedWidth := width - 1 //adjust width to work with arrays
	// create 2d matrix
	for i := range resultMatrix {
		resultMatrix[i] = make([]float64, width)
	}
	// first lerp columns, then using the column data lerp the rows
	columns := getConfidenceColumns(height, confidence)
	for y := 0; y < len(columns); y++ {
		for x := 0; x < confWidth; x++ {
			next := x + 1
			xStartPos := int(lerp(0.0, float64(adjustedWidth), float64(x)/float64(confWidth)))
			xEndPos := int(lerp(0.0, float64(adjustedWidth), float64(next)/float64(confWidth)))
			chunk := getConfidenceChunk(xEndPos-xStartPos, columns[y][x], columns[y][next])
			for k, v := range chunk {
				resultMatrix[y][xStartPos+k] = v
			}
		}
	}
	return resultMatrix
}

func getConfidenceColumns(height int, confidence *[][]float64) [][]float64 {
	result := make([][]float64, height)
	xSize := len((*confidence)[0])
	arrEnd := len(*confidence) - 1
	adjustedHeight := height - 1 // adjusts height to work with arrays
	// init array
	for i := range result {
		result[i] = make([]float64, xSize)
	}
	for i := 0; i < arrEnd; i++ {
		for j := 0; j < xSize; j++ {
			next := i + 1
			start := (*confidence)[i][j]                                                      // get start confidence value for chunk
			end := (*confidence)[next][j]                                                     // get end confidence value for chunk
			yStartPos := int(lerp(0.0, float64(adjustedHeight), float64(i)/float64(arrEnd)))  // lerp indices to get start index
			yEndPos := int(lerp(0.0, float64(adjustedHeight), float64(next)/float64(arrEnd))) // lerp indices to get end index
			numElements := yEndPos - yStartPos                                                // calculate number of elements for chunk
			chunk := getConfidenceChunk(numElements, start, end)
			for k, v := range chunk {
				result[yStartPos+k][j] = v
			}
		}
	}
	return result

}
func getConfidenceChunk(numElements int, start float64, end float64) []float64 {
	result := make([]float64, numElements+1) // adds one to buffer because there is overlap in the scaling of the matrix for example
	// first chunk could be from idx 0-2 (inclusive) the next chunk would start from 2-4 (also inclusive), thus overlap
	for i := 0; i <= numElements; i++ {
		result[i] = lerp(start, end, float64(i)/float64(numElements))
	}
	return result
}

// ConfidenceMatrixToImage takes the confidences matrix and a supplied colorScale function and returns an image.
func ConfidenceMatrixToImage(confidence [][]float64, colorScale func(float64) *color.RGBA, opacity uint8) *image.RGBA {
	height := len(confidence)
	width := len(confidence[0])
	resultImage := image.NewRGBA(image.Rect(0, 0, width, height))
	outputIdx := 0
	step := 4
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := colorScale(float64(confidence[y][x]))
			// premultiplied alpha
			alpha := float64(opacity) / 255.0
			r := uint8(float64(color.R) * alpha)
			b := uint8(float64(color.B) * alpha)
			g := uint8(float64(color.G) * alpha)
			resultImage.Pix[outputIdx] = r
			resultImage.Pix[outputIdx+1] = g
			resultImage.Pix[outputIdx+2] = b
			resultImage.Pix[outputIdx+3] = opacity
			outputIdx += step
		}
	}
	return resultImage
}

// getMaxDimensions return max from array. Return order width, height
func getMaxDimensions(bandImages *[]*image.Gray16) (int, int) {
	width := 0
	height := 0
	for _, bandImage := range *bandImages {
		// extract input raster size and update max x,y
		xSize := bandImage.Bounds().Dx()
		ySize := bandImage.Bounds().Dy()
		if xSize > width {
			width = xSize
		}
		if ySize > height {
			height = ySize
		}
	}
	return width, height
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

func createRGBAFromBands(xSize int, ySize int, bandImages []*image.Gray16, advancedColorModel bool, options ...Options) *image.RGBA {
	// Create a new RGBA image to hold the collected bands
	outputImage := image.NewRGBA(image.Rect(0, 0, xSize, ySize))

	// Copy the 16 bit band images into the 8 bit target image.  If a band image couldn't be processed
	// earlier, we set to grey.
	outputIdx := 0
	bandBuffer := [3]float64{0, 0, 0}
	for i := 0; i < (xSize * ySize * 2); i += 2 {
		for j, bandImage := range bandImages {
			if bandImage != nil {
				grayValue16 := uint16(bandImage.Pix[i])<<8 | uint16(bandImage.Pix[i+1])
				bandBuffer[j] = float64(grayValue16) / Sentinel2Max
			} else {
				bandBuffer[j] = 0.5
			}
		}
		rgb := ConvertS2ToRgb(bandBuffer, advancedColorModel, options...)
		outputImage.Pix[outputIdx] = uint8(rgb[0] * 255)   // r
		outputImage.Pix[outputIdx+1] = uint8(rgb[1] * 255) // g
		outputImage.Pix[outputIdx+2] = uint8(rgb[2] * 255) // b
		outputImage.Pix[outputIdx+3] = 0xFF                // max out the A channel
		outputIdx += len(bandImages) + 1                   // +1 for alpha channel
	}
	return outputImage
}

func createRGBAFromRamp(xSize int, ySize int, bandImages []*image.Gray16, transform func(*OptramEdges, ...uint16) float64, ramp []uint8, edges *OptramEdges) *image.RGBA {
	// Create a new RGBA image to hold the collected bands
	outputImage := image.NewRGBA(image.Rect(0, 0, xSize, ySize))

	rampElements := (len(ramp) - 2) / 3

	// Copy the 16 bit band images into the 8 bit target image.  If a band image couldn't be processed
	// earlier, we set to grey.
	outputIdx := 0
	for i := 0; i < (xSize * ySize * 2); i += 2 {
		grayValues := make([]uint16, len(bandImages))
		valid := false
		for ii, v := range bandImages {
			grayValues[ii] = uint16(v.Pix[i])<<8 | uint16(v.Pix[i+1])
			if !valid {
				valid = grayValues[ii] != 0
			}
		}
		// compute NDVI ratio
		transformedValue := 0.0
		if valid {
			transformedValue = transform(edges, grayValues...)
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
func ImageToJPEG(image *image.RGBA) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, image, nil); err != nil {
		return nil, errors.Wrap(err, "failed so encode png file")
	}
	return buf.Bytes(), nil
}

// ImageToPNG encodes RGBA image as PNG byte array
func ImageToPNG(image *image.RGBA) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, image); err != nil {
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

// CreatePolygonFromCoordinates creates a string that captures the polygon defined
// by the coordinates.
func CreatePolygonFromCoordinates(coordinates []float64) string {
	if len(coordinates) == 0 {
		return "POLYGON EMPTY"
	}

	// polygon must be closed, so if it isnt, then add the first point at the end to close it
	if coordinates[0] != coordinates[len(coordinates)-2] || coordinates[1] != coordinates[len(coordinates)-1] {
		coordinates = append(coordinates, coordinates[0], coordinates[1])
	}

	geometryString := ""
	for i := 0; i < len(coordinates); i += 2 {
		geometryString = fmt.Sprintf("%s,%f %f", geometryString, coordinates[i], coordinates[i+1])
	}
	return fmt.Sprintf("POLYGON((%s))", geometryString[1:])
}

func lerp(v0 float64, v1 float64, t float64) float64 {
	return (1.0-t)*v0 + t*v1
}

// ParseGeoHashFromID will extrapolate the geohash from the path
func ParseGeoHashFromID(ID string, precision int) string {
	geoHash := strings.Split(ID, "_")[0]
	return geoHash[:precision]
}
