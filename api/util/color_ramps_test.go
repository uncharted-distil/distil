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

package util

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	log "github.com/unchartedsoftware/plog"
)

func TestRamp(t *testing.T) {
	ramp := GenerateRamp([]RampEntry{
		{0.0, color.RGBA{0, 0, 0, 255}},
		{0.5, color.RGBA{128, 128, 128, 255}},
		{1.0, color.RGBA{255, 255, 255, 255}},
	}, 255, Lab)
	image := RampToImage(5, ramp)

	// compare to gold standard image
	testImage, err := LoadPNGImage("test/ramp_image.png")
	if err != nil {
		log.Error(err)
	}
	assert.Equal(t, testImage, image)
}
