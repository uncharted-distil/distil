//
//   Copyright © 2020 Uncharted Software Inc.
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
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	log "github.com/unchartedsoftware/plog"
)

func TestImageFromCombination(t *testing.T) {
	bandMap := map[string]string{
		"b01": "S2A_MSIL2A_20171121T112351_79_21_B01.tif",
		"b02": "S2A_MSIL2A_20171121T112351_79_21_B02.tif",
		"b03": "S2A_MSIL2A_20171121T112351_79_21_B03.tif",
		"b04": "S2A_MSIL2A_20171121T112351_79_21_B04.tif",
		"b05": "S2A_MSIL2A_20171121T112351_79_21_B05.tif",
		"b06": "S2A_MSIL2A_20171121T112351_79_21_B06.tif",
		"b07": "S2A_MSIL2A_20171121T112351_79_21_B07.tif",
		"b08": "S2A_MSIL2A_20171121T112351_79_21_B08.tif",
		"b8A": "S2A_MSIL2A_20171121T112351_79_21_B8A.tif",
		"b09": "S2A_MSIL2A_20171121T112351_79_21_B09.tif",
		"b11": "S2A_MSIL2A_20171121T112351_79_21_B11.tif",
		"b12": "S2A_MSIL2A_20171121T112351_79_21_B12.tif",
	}

	composedImage, err := ImageFromCombination("../test/bigearthnet", bandMap, NaturalColors2, ImageScale{}, &OptramEdges{}, "")

	assert.NoError(t, err)
	assert.NotNil(t, composedImage)
	assert.True(t, len(composedImage.Pix) > 0)

	// compare to gold standard image
	testImage, err := LoadPNGImage("../test/4_3_2.png")
	if err != nil {
		log.Error(err)
	}
	assert.Equal(t, testImage, composedImage)
}

func TestImageFromBandsTrueColor(t *testing.T) {
	// Test basic loading - all image sources are the same size.
	composedImage, err := ImageFromBands([]string{
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B04.tif",
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B03.tif",
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B02.tif",
	}, nil, nil, ImageScale{}, &OptramEdges{}, true)
	assert.NoError(t, err)
	assert.NotNil(t, composedImage)
	assert.True(t, len(composedImage.Pix) > 0)

	// compare to gold standard image
	testImage, err := LoadPNGImage("../test/4_3_2.png")
	if err != nil {
		log.Error(err)
	}
	assert.Equal(t, testImage, composedImage)
}

func TestImageFromBandsResize(t *testing.T) {
	// Tests loading data from sources that are 3 different sizes in terms
	// of pixels.
	composedImage, err := ImageFromBands([]string{
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B12.tif",
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B08.tif",
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B04.tif",
	}, nil, nil, ImageScale{}, &OptramEdges{}, true)

	assert.NoError(t, err)
	assert.NotNil(t, composedImage)
	assert.True(t, len(composedImage.Pix) > 0)

	// compare to gold standard image
	testImage, err := LoadPNGImage("../test/12_8_4.png")
	if err != nil {
		log.Error(err)
	}
	assert.Equal(t, testImage, composedImage)
}

func TestImageFromRamp(t *testing.T) {
	// Tests loading data from sources that are 3 different sizes in terms
	// of pixels.
	composedImage, err := ImageFromBands([]string{
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B08.tif",
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B11.tif",
	}, BrownYellowBlueRamp, SentinelBandCombinations[NDMI].Transform, ImageScale{}, &OptramEdges{}, true)

	assert.NoError(t, err)
	assert.NotNil(t, composedImage)
	assert.True(t, len(composedImage.Pix) > 0)

	// compare to gold standard image
	testImage, err := LoadPNGImage("../test/8_11_ramp.png")
	if err != nil {
		log.Error(err)
	}
	assert.Equal(t, testImage, composedImage)
}

func TestImageFromRampClamped(t *testing.T) {
	// Tests loading data from sources that are 3 different sizes in terms
	// of pixels.
	composedImage, err := ImageFromBands([]string{
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B08.tif",
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B04.tif",
	}, RedYellowGreenRamp, SentinelBandCombinations[NDVI].Transform, ImageScale{}, &OptramEdges{}, true)

	assert.NoError(t, err)
	assert.NotNil(t, composedImage)
	assert.True(t, len(composedImage.Pix) > 0)

	// compare to gold standard image
	testImage, err := LoadPNGImage("../test/8_4_ramp.png")
	if err != nil {
		log.Error(err)
	}
	assert.Equal(t, testImage, composedImage)
}

func TestImageFromBandsMissing(t *testing.T) {
	// Tests handling the case where one of the bands contains bad data.  The missing
	// band will be mapped to grey.
	composedImage, err := ImageFromBands([]string{
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B12.tif",
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B08.tif",
		"../test/bigearthnet/S2A_MSIL2A_20171121T112351_79_21_B04.tif",
	}, nil, SentinelBandCombinations[ShortwaveInfrared].Transform, ImageScale{}, &OptramEdges{}, true)
	assert.NoError(t, err)
	assert.NotNil(t, composedImage)
	assert.True(t, len(composedImage.Pix) > 0)

	// compare to gold standard image
	testImage, err := LoadPNGImage("../test/12_8_4.png")
	if err != nil {
		log.Error(err)
	}
	assert.Equal(t, testImage, composedImage)
}

func TestSaveAndLoadPng(t *testing.T) {
	// Tests saving and loading of png by creating a test image, writing it out, and then
	// reloading it to make sure everything matches.

	// create a 4x4 image with incrementing pixels values.
	image := generateTestImage(4, 4)
	assert.NoError(t, SavePNGImage(image, "../test/test.png"))

	resultImage, err := LoadPNGImage("../test/test.png")
	assert.NoError(t, err)
	assert.Equal(t, image, resultImage)
}

func TestImageToJPEG(t *testing.T) {
	testImage := generateTestImage(4, 4)
	jpegBytes, err := ImageToJPEG(testImage)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(jpegBytes), 1)
}

func generateTestImage(xSize int, ySize int) *image.RGBA {
	image := image.NewRGBA(image.Rect(0, 0, xSize, ySize))
	offset := 0
	for i := 0; i < xSize*ySize; i++ {
		for j := 0; j < 3; j++ {
			image.Pix[offset] = uint8(i)
			offset++
		}
		image.Pix[offset] = uint8(0xFF)
		offset++
	}
	return image
}
