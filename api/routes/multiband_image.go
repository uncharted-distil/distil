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
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

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

const (
	// ThumbnailDimensions is hard coded thumbnail dimension -- could be refactored to be default if we want client to dictate size.
	ThumbnailDimensions = 125
)

func getOptions(requestURI string) string {
	idx := strings.LastIndex(requestURI, "/") + 1 // exclusive
	return requestURI[idx:]
}

// MultiBandImageHandler fetches individual band images and combines them into a single RGB image using the supplied mapping.
func MultiBandImageHandler(ctor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, config env.Config) func(http.ResponseWriter, *http.Request) {
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
		imageScale := imagery.ImageScale{}
		// get metadata client
		storage, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}
		dataStorage, err := dataCtor()
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
		options := imagery.Options{Gain: 2.5, Gamma: 2.2, GainL: 1.0, Scale: false} // default options for color correction
		if paramOption != "" {
			err := json.Unmarshal([]byte(paramOption), &options)
			if err != nil {
				handleError(w, err)
				return
			}
		}
		if isThumbnail {
			imageScale = imagery.ImageScale{Width: ThumbnailDimensions, Height: ThumbnailDimensions}
			// if thumbnail scale should be 0
			options.Scale = false
		}

		// need to get the band -> filename from the data
		bandMapping, err := getBandMapping(res, []string{imageID}, dataStorage)
		if err != nil {
			handleError(w, err)
			return
		}
		var optramMap map[string]imagery.OptramEdges
		optramPath := ""
		edge := imagery.OptramEdges{}
		precision := 0
		if bandCombo == imagery.OPTRAM {
			optramPath = strings.Join([]string{env.ResolvePath(res.Source, res.Folder), imagery.OPTRAMJSONFile}, "/")
			optramMap, precision, err = imagery.ReadOptramFile(optramPath)
			if err != nil {
				handleError(w, err)
				return
			}
			geoHash := imagery.ParseGeoHashFromID(imageID, precision)
			edge = optramMap[geoHash]
		}

		img, err := imagery.ImageFromCombination(sourcePath, bandMapping[imageID], bandCombo, imageScale, &edge, options)
		if err != nil {
			handleError(w, err)
			return
		}
		if options.Scale && config.ShouldScaleImages {
			img = c_util.UpscaleImage(img, c_util.GetModelType(config.ModelType))
		}
		imageBytes, err := imagery.ImageToJPEG(img)
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

func getBandMapping(ds *api.Dataset, groupKeys []string, dataStorage api.DataStorage) (map[string]map[string]string, error) {
	// build a filter to only include rows matching a group id
	var groupingCol string
	var bandCol string
	var fileCol string

	// find the components of the multiband image
	for _, v := range ds.Variables {
		if v.IsGrouping() && model.IsMultiBandImage(v.Type) {
			gcg := v.Grouping.(*model.MultiBandImageGrouping)
			groupingCol = gcg.IDCol
			fileCol = gcg.ImageCol
			bandCol = gcg.BandCol
			break
		}
	}
	if groupingCol == "" {
		return nil, errors.Errorf("no grouping col found in dataset")
	}
	if fileCol == "" {
		return nil, errors.Errorf("no file col found in dataset")
	}
	if bandCol == "" {
		return nil, errors.Errorf("no band col found in dataset")
	}

	filter := &api.FilterParams{}
	filter.AddFilter(&model.Filter{
		Key:        groupingCol,
		Type:       model.CategoricalFilter,
		Categories: groupKeys,
		Mode:       model.IncludeFilter,
	})
	filter.Variables = []string{fileCol, bandCol, groupingCol}

	// pull back all rows for a group id
	data, err := dataStorage.FetchData(ds.ID, ds.StorageName, filter, true, nil)
	if err != nil {
		return nil, err
	}

	// cycle through results to build the band mapping
	fileVariable, ok := data.Columns[fileCol]
	if !ok {
		return nil, errors.Errorf("no file column found in stored data")
	}
	bandVariable, ok := data.Columns[bandCol]
	if !ok {
		return nil, errors.Errorf("no band column found in stored data")
	}
	groupVariable, ok := data.Columns[groupingCol]
	if !ok {
		return nil, errors.Errorf("no group column found in stored data")
	}
	fileColumn := fileVariable.Index
	bandColumn := bandVariable.Index
	groupColumn := groupVariable.Index

	mapping := map[string]map[string]string{}
	for _, r := range data.Values {
		groupKey := r[groupColumn].Value.(string)
		if mapping[groupKey] == nil {
			mapping[groupKey] = map[string]string{}
		}
		// the mapping expects bXX but the database only stores XX
		bandKey := fmt.Sprintf("b%s", r[bandColumn].Value.(string))
		mapping[groupKey][bandKey] = r[fileColumn].Value.(string)
	}

	return mapping, nil
}
