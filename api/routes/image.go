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

package routes

import (
	"bytes"
	"image"
	"image/draw"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"

	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	c_util "github.com/uncharted-distil/distil-image-upscale/c_util"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/imagery"
	"goji.io/v3/pat"
)

// ImageHandler provides a static file lookup route using simple directory mapping.
func ImageHandler(ctor api.MetadataStorageCtor, config *env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// resources can either be local or remote
		dataset := pat.Param(r, "dataset")
		file := pat.Param(r, "file")

		// check if a thumbnail is requested
		isThumbnail, err := strconv.ParseBool(pat.Param(r, "is-thumbnail"))
		if err != nil {
			handleError(w, err)
			return
		}
		scale, err := strconv.ParseInt(pat.Param(r, "scale"), 10, 32)
		if err != nil {
			handleError(w, err)
			return
		}

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
		metaDisk, err := metadata.LoadMetadataFromOriginalSchema(path.Join(sourcePath, compute.D3MDataSchema), false)
		if err != nil {
			handleError(w, err)
			return
		}
		// need to read the dataset doc to determine the path to the data resource
		for _, dr := range metaDisk.DataResources {
			if dr.IsCollection && dr.ResType == model.ResTypeImage {
				sourcePath = model.GetResourcePathFromFolder(sourcePath, dr)
				break
			}
		}
		data, err := ioutil.ReadFile(path.Join(sourcePath, file))
		if err != nil {
			handleError(w, err)
			return
		}

		if isThumbnail {
			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				handleError(w, err)
				return
			}
			img = resize.Thumbnail(ThumbnailDimensions, ThumbnailDimensions, img, resize.Lanczos3)
			rgbaImg := image.NewRGBA(image.Rect(0, 0, ThumbnailDimensions, ThumbnailDimensions))
			draw.Draw(rgbaImg, image.Rect(0, 0, ThumbnailDimensions, ThumbnailDimensions), img, img.Bounds().Min, draw.Src)
			imageBytes, err := imagery.ImageToJPEG(rgbaImg)
			if err != nil {
				handleError(w, err)
				return
			}
			data = imageBytes
		}
		if scale > 0 && config.ShouldScaleImages {
			if scale > 3 {
				// dont allow upscaling past factor of 6
				scale = 3
			}
			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				handleError(w, err)
				return
			}
			dimensions := img.Bounds()
			rgbaImg := image.NewRGBA(image.Rect(0, 0, dimensions.Max.X, dimensions.Max.Y))
			draw.Draw(rgbaImg, image.Rect(0, 0, dimensions.Max.X, dimensions.Max.Y), img, img.Bounds().Min, draw.Src)
			// multiple passes for increasing scale dramatically
			for i := 0; i < int(scale); i++ {
				rgbaImg = c_util.UpscaleImage(rgbaImg, c_util.GetModelType(config.ModelType))
				imageBytes, err := imagery.ImageToJPEG(rgbaImg)
				if err != nil {
					handleError(w, err)
					return
				}
				data = imageBytes
			}
		}
		_, err = w.Write(data)
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to write image resource bytes"))
			return
		}
	}
}
