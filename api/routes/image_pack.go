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
	"bytes"
	"encoding/json"
	"image"
	"image/draw"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/imagery"
	log "github.com/unchartedsoftware/plog"
)

// ImagePackRequest is the expected request format for the route (Band can be an empty string)
type ImagePackRequest struct {
	Dataset  string   `json:"dataset"`
	ImageIDs []string `json:"imageIds"`
	Band     string   `json:"band,omitempty"`
}

// ImagePackResult is the expected post result for MultiBandImagePackHandler
type ImagePackResult struct {
	ImagesBuffer [][]byte `json:"images"`
	ImageIDs     []string `json:"imageIds"`
	ErrorIDs     []string `json:"errorIds"`
}

type chanStruct struct {
	data     [][]byte
	IDs      []string
	errorIDs []string
}

func postParamsToImagePackRequest(r *http.Request) (*ImagePackRequest, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse POST request")
	}
	result := &ImagePackRequest{}
	err = json.Unmarshal(body, result)
	return result, err
}

// MultiBandImagePackHandler fetches individual band images and combines them into a single RGB image using the supplied mapping.
func MultiBandImagePackHandler(ctor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, config env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := postParamsToImagePackRequest(r)
		if err != nil {
			handleError(w, err)
			return
		}
		// default to getImages
		funcPointer := getImages
		if params.Band != "" {
			// if band is not empty then get multiBandImages
			funcPointer = getMultiBandImages
		}
		// channel for threads to communicate
		result := make(chan chanStruct)
		// ImageThreadPool is an environment variable defaults to 2 (works great with 6)
		numOfThreads := config.ImageThreadPool
		// if no IDs return
		if len(params.ImageIDs) == 0 {
			err = handleJSON(w, ImagePackResult{
				ImagesBuffer: [][]byte{},
				ImageIDs:     []string{},
				ErrorIDs:     []string{},
			})
			if err != nil {
				handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
				return
			}
			return
		}
		// reduce numOfThreads to the number of ImageIDs if it is lower than 6
		if numOfThreads > len(params.ImageIDs) {
			numOfThreads = len(params.ImageIDs)
		}

		for i := 0; i < numOfThreads; i++ {
			go funcPointer(params, i, numOfThreads, result, ctor, dataCtor)
		}
		imagesBuffer := [][]byte{}
		IDs := []string{}
		errorIDs := []string{}
		for i := 0; i < numOfThreads; i++ {
			// no guaruntee of threads finishing in order so we supply the IDs back as well
			r := <-result
			imagesBuffer = append(imagesBuffer, r.data...)
			IDs = append(IDs, r.IDs...)
			errorIDs = append(errorIDs, r.errorIDs...)
		}
		// close channel
		close(result)
		if len(imagesBuffer) == 0 {
			handleError(w, errors.New("Server error"))
			return
		}
		err = handleJSON(w, ImagePackResult{
			ImagesBuffer: imagesBuffer,
			ImageIDs:     IDs,
			ErrorIDs:     errorIDs,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
func getImages(imagePackRequest *ImagePackRequest, threadID int, numThreads int, result chan chanStruct, ctor api.MetadataStorageCtor, dataCtor api.DataStorageCtor) {
	temp := [][]byte{}
	IDs := []string{}
	errorIDs := []string{}

	// get common storage
	storage, err := ctor()
	if err != nil {
		log.Error(err)
		return
	}

	res, err := storage.FetchDataset(imagePackRequest.Dataset, false, false, false)
	if err != nil {
		log.Error(err)
		return
	}

	sourcePath := env.ResolvePath(res.Source, res.Folder)
	metaDisk, err := metadata.LoadMetadataFromOriginalSchema(path.Join(sourcePath, compute.D3MDataSchema), false)
	if err != nil {
		log.Error(err)
		return
	}
	// need to read the dataset doc to determine the path to the data resource
	for _, dr := range metaDisk.DataResources {
		if dr.IsCollection && dr.ResType == model.ResTypeImage {
			sourcePath = model.GetResourcePathFromFolder(sourcePath, dr)
			break
		}
	}
	// loop through image info
	for i := threadID; i < len(imagePackRequest.ImageIDs); i += numThreads {
		imageID := imagePackRequest.ImageIDs[i]

		data, err := ioutil.ReadFile(path.Join(sourcePath, imageID))
		if err != nil {
			handleThreadError(&errorIDs, &imageID, &err)
			continue
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			handleThreadError(&errorIDs, &imageID, &err)
			continue
		}
		img = resize.Thumbnail(ThumbnailDimensions, ThumbnailDimensions, img, resize.Lanczos3)
		rgbaImg := image.NewRGBA(image.Rect(0, 0, ThumbnailDimensions, ThumbnailDimensions))
		draw.Draw(rgbaImg, image.Rect(0, 0, ThumbnailDimensions, ThumbnailDimensions), img, img.Bounds().Min, draw.Src)
		imageBytes, err := imagery.ImageToJPEG(rgbaImg)
		if err != nil {
			handleThreadError(&errorIDs, &imageID, &err)
			continue
		}
		temp = append(temp, imageBytes)
		IDs = append(IDs, imageID)
	}
	result <- chanStruct{data: temp, IDs: IDs, errorIDs: errorIDs}
}
func getMultiBandImages(multiBandPackRequest *ImagePackRequest, threadID int, numThreads int, result chan chanStruct, ctor api.MetadataStorageCtor, dataCtor api.DataStorageCtor) {
	temp := [][]byte{}
	IDs := []string{}
	errorIDs := []string{}
	// get common storage
	storage, err := ctor()
	if err != nil {
		log.Error(err)
		return
	}
	dataStorage, err := dataCtor()
	if err != nil {
		log.Error(err)
		return
	}

	res, err := storage.FetchDataset(multiBandPackRequest.Dataset, false, false, false)
	if err != nil {
		log.Error(err)
		return
	}
	sourcePath := env.ResolvePath(res.Source, res.Folder)
	metaDisk, err := metadata.LoadMetadataFromOriginalSchema(path.Join(sourcePath, compute.D3MDataSchema), false)
	if err != nil {
		log.Error(err)
		return
	}
	// need to read the dataset doc to determine the path to the data resource
	for _, dr := range metaDisk.DataResources {
		if dr.IsCollection && dr.ResType == model.ResTypeImage {
			sourcePath = model.GetResourcePathFromFolder(sourcePath, dr)
			break
		}
	}
	options := imagery.Options{Gain: 2.5, Gamma: 2.2, GainL: 1.0, Scale: false} // default options for color correction

	imageScale := imagery.ImageScale{Width: ThumbnailDimensions, Height: ThumbnailDimensions}

	// get the image file names
	imageIDs := []string{}
	for i := threadID; i < len(multiBandPackRequest.ImageIDs); i += numThreads {
		imageIDs = append(imageIDs, multiBandPackRequest.ImageIDs[i])
	}

	// need to get the band -> filename from the data
	bandMapping, err := getBandMapping(res, imageIDs, dataStorage)
	if err != nil {
		log.Error(err)
		return
	}

	// loop through image info
	for i := threadID; i < len(multiBandPackRequest.ImageIDs); i += numThreads {
		imageID := multiBandPackRequest.ImageIDs[i]

		img, err := imagery.ImageFromCombination(sourcePath, bandMapping[imageID], multiBandPackRequest.Band, imageScale, options)
		if err != nil {
			handleThreadError(&errorIDs, &imageID, &err)
			continue
		}

		imageBytes, err := imagery.ImageToJPEG(img)
		if err != nil {
			handleThreadError(&errorIDs, &imageID, &err)
			continue
		}
		temp = append(temp, imageBytes)
		IDs = append(IDs, imageID)
	}

	result <- chanStruct{data: temp, IDs: IDs, errorIDs: errorIDs}
}

func handleThreadError(errorIDs *[]string, imageID *string, err *error) {
	*errorIDs = append(*errorIDs, *imageID)
	log.Error(*err)
}
