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

package routes

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	"goji.io/v3/pat"
)

const (
	// ThumbnailDimensions is hard coded thumbnail dimension -- could be refactored to be default if we want client to dictate size.
	ThumbnailDimensions = 125
)

func getOptions(requestURI string) string {
	idx := strings.LastIndex(requestURI, "/") + 1 // exclusive
	return requestURI[idx:]
}

// MultiBandImageHandler fetches individual band images and combines them into a single RGB image using the supplied mapping.
func MultiBandImageHandler(ctor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		imageID := pat.Param(r, "image-id")
		bandCombo := pat.Param(r, "band-combination")
		paramOption := getOptions(r.URL.Path)
		isThumbnail, err := strconv.ParseBool(pat.Param(r, "is-thumbnail"))
		if err != nil {
			handleError(w, err)
			return
		}
		imageScale := util.ImageScale{}
		// get metadata client
		storage, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}
		res, err := storage.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		sourcePath := env.ResolvePath(res.Source, res.Folder)

		// need to read the dataset doc to determine the path to the data resource
		metaDisk, err := metadata.LoadMetadataFromOriginalSchema(path.Join(sourcePath, compute.D3MDataSchema), false)
		if err != nil {
			handleError(w, err)
			return
		}
		for _, dr := range metaDisk.DataResources {
			if dr.IsCollection && dr.ResType == model.ResTypeImage {
				sourcePath = model.GetResourcePathFromFolder(sourcePath, dr)
				break
			}
		}

		if isThumbnail {
			imageScale = util.ImageScale{Width: ThumbnailDimensions, Height: ThumbnailDimensions}
		}
		options := util.Options{Gain: 2.5, Gamma: 2.2, GainL: 1.0} // default options for color correction
		if paramOption != "" {
			err := json.Unmarshal([]byte(paramOption), &options)
			if err != nil {
				handleError(w, err)
				return
			}
		}
		img, err := util.ImageFromCombination(sourcePath, imageID, bandCombo, imageScale, options)
		if err != nil {
			handleError(w, err)
			return
		}

		imageBytes, err := util.ImageToJPEG(img)
		if err != nil {
			handleError(w, err)
			return
		}
		_, err = w.Write(imageBytes)
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to write image resource bytes"))
			return
		}
	}
}
