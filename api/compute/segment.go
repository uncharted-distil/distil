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

package compute

import (
	"strconv"

	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil/api/util/imagery"
)

// BuildSegmentationImage uses the raw segmentation output to build an image layer.
func BuildSegmentationImage(rawSegmentation [][]interface{}) (map[string][]byte, error) {
	// output is mapping of d3m index to new segmentation layer
	output := map[string][]byte{}
	// need to output all the masks as images
	for _, r := range rawSegmentation[1:] {
		// create the image that captures the mask
		d3mIndex := r[0].(string)
		rawMask := r[1].([]interface{})
		rawFloats := make([][]float64, len(rawMask))
		for i, f := range rawMask {
			dataF := f.([]interface{})
			nestedFloats := make([]float64, len(dataF))
			for j, nf := range dataF {
				fp, err := strconv.ParseFloat(nf.(string), 64)
				if err != nil {
					return nil, errors.Wrapf(err, "unable to parse mask")
				}
				nestedFloats[j] = fp
			}
			rawFloats[i] = nestedFloats
		}

		filter := imagery.ConfidenceMatrixToImage(rawFloats, imagery.MagmaColorScale, uint8(100))
		imageBytes, err := imagery.ImageToPNG(filter)
		if err != nil {
			return nil, err
		}
		output[d3mIndex] = imageBytes
	}

	return output, nil
}
