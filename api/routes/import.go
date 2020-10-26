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

package routes

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"path"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/dataset"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util"
	"github.com/uncharted-distil/distil/api/util/json"
)

// ImportHandler imports a dataset to the local file system and then ingests it.
func ImportHandler(dataCtor api.DataStorageCtor, datamartCtors map[string]api.MetadataStorageCtor,
	fileMetaCtor api.MetadataStorageCtor, esMetaCtor api.MetadataStorageCtor,
	config *env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetID := pat.Param(r, "datasetID")
		source := metadata.DatasetSource(pat.Param(r, "source"))
		provenance := pat.Param(r, "provenance")
		isSampling := true // Flag to sample imported dataset

		var origins []*model.DatasetOrigin
		var rawGroupings []map[string]interface{}
		if source == metadata.Augmented && provenance == "local" {
			// parse POST params
			params, err := getPostParameters(r)
			if err != nil {
				handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
				return
			}

			if params == nil {
				missingParamErr(w, "parameters")
				return
			}

			// Check if we want to sample the dataset
			if params["nosample"] != nil {
				isSampling = false
			}

			if params["joinedDataset"] != nil {
				if params["originalDataset"] == nil {
					missingParamErr(w, "originalDataset")
					return
				}

				if params["joinedDataset"] == nil {
					missingParamErr(w, "joinedDataset")
					return
				}

				// set the origin information
				originalDataset, okOriginal := params["originalDataset"].(map[string]interface{})
				joinedDataset, okJoined := params["joinedDataset"].(map[string]interface{})
				if okOriginal && okJoined {
					// combine the origin and joined dateset into an array of structs
					origins, err = getOriginsFromMaps(originalDataset, joinedDataset)
					if err != nil {
						handleError(w, errors.Wrap(err, "unable to marshal dataset origins from JSON to struct"))
						return
					}
				}
			} else {
				if params["path"] == nil {
					missingParamErr(w, "path")
					return
				}

				datasetPath := params["path"].(string)
				log.Infof("Creating dataset '%s' from '%s'", datasetID, datasetPath)
				creationResult, err := createDataset(datasetPath, datasetID, config)
				if err != nil {
					handleError(w, errors.Wrap(err, "unable to create raw dataset"))
					return
				}
				datasetID = creationResult.name
				rawGroupings = creationResult.groups
				log.Infof("Created dataset '%s' from local source '%s'", datasetID, datasetPath)
			}
		}

		meta, err := createMetadataStorageForSource(source, provenance, datamartCtors, fileMetaCtor, esMetaCtor)
		if err != nil {
			handleError(w, err)
			return
		}

		// import the dataset to the local filesystem.
		uri := env.ResolvePath(source, datasetID)
		log.Infof("Importing dataset '%s' from '%s'", datasetID, uri)
		_, err = meta.ImportDataset(datasetID, uri)
		if err != nil {
			handleError(w, err)
			return
		}

		// ingest the imported dataset
		ingestSteps := &task.IngestSteps{ClassificationOverwrite: false, RawGroupings: rawGroupings}
		ingestConfig := task.NewConfig(*config)

		// Check if the imported dataset should be sampled
		if !isSampling {
			ingestConfig.SampleRowLimit = math.MaxInt32 // Maximum int value.
		}

		log.Infof("Ingesting dataset '%s'", uri)
		ingestResult, err := task.IngestDataset(source, dataCtor, esMetaCtor, datasetID, origins, api.DatasetTypeModelling, ingestConfig, ingestSteps)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data and sent the response back
		err = handleJSON(w, map[string]interface{}{
			"dataset":  ingestResult.DatasetID,
			"sampled":  ingestResult.Sampled,
			"rowCount": ingestResult.RowCount,
			"result":   "ingested"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}

func getOriginsFromMaps(originalDataset map[string]interface{}, joinedDataset map[string]interface{}) ([]*model.DatasetOrigin, error) {
	joinSuggestions := make([]interface{}, 0)
	if originalDataset["joinSuggestion"] != nil {
		joinSuggestions = append(joinSuggestions, originalDataset["joinSuggestion"].([]interface{})...)
	}
	if joinedDataset["joinSuggestion"] != nil {
		joinSuggestions = append(joinSuggestions, joinedDataset["joinSuggestion"].([]interface{})...)
	}

	origins := make([]*model.DatasetOrigin, len(joinSuggestions))
	for i, js := range joinSuggestions {
		targetOriginModel := model.DatasetOrigin{}
		targetJoin := js.(map[string]interface{})
		targetJoinOrigin := targetJoin["datasetOrigin"].(map[string]interface{})
		err := json.MapToStruct(&targetOriginModel, targetJoinOrigin)
		if err != nil {
			return nil, err
		}
		origins[i] = &targetOriginModel
	}

	return origins, nil
}

func createMetadataStorageForSource(datasetSource metadata.DatasetSource, provenance string,
	datamartCtors map[string]api.MetadataStorageCtor,
	fileMetaCtor api.MetadataStorageCtor, esMetaCtor api.MetadataStorageCtor) (api.MetadataStorage, error) {
	if datasetSource == metadata.Contrib {
		return datamartCtors[provenance]()
	}
	if datasetSource == metadata.Seed {
		return esMetaCtor()
	}
	if datasetSource == metadata.Augmented {
		return fileMetaCtor()
	}
	return nil, fmt.Errorf("unrecognized source `%v`", datasetSource)
}

type datasetCreationResult struct {
	name   string
	path   string
	groups []map[string]interface{}
}

func createDataset(datasetPath string, datasetName string, config *env.Config) (*datasetCreationResult, error) {
	// check if already a dataset
	if util.IsDatasetDir(datasetPath) {
		return &datasetCreationResult{
			name: datasetName,
			path: datasetPath,
		}, nil
	}

	// create the raw dataset
	log.Infof("Creating raw dataset '%s' from '%s'", datasetName, datasetPath)
	ds, groups, err := createRawDataset(datasetPath, datasetName)
	if err != nil {
		return nil, err
	}

	// create the formatted d3m dataset
	outputPath := path.Join(config.D3MOutputDir, config.AugmentedSubFolder)
	log.Infof("Creating final dataset '%s' from '%s'", datasetName, outputPath)
	datasetName, formattedPath, err := task.CreateDataset(datasetName, ds, outputPath, config)
	if err != nil {
		return nil, err
	}

	return &datasetCreationResult{
		name:   datasetName,
		path:   formattedPath,
		groups: groups,
	}, nil
}

func createRawDataset(datasetPath string, datasetName string) (task.DatasetConstructor, []map[string]interface{}, error) {
	if util.IsArchiveFile(datasetPath) {
		// expand the archive
		expandedInfo, err := dataset.ExpandZipDataset(datasetPath, datasetName)
		if err != nil {
			return nil, nil, err
		}

		datasetPath = expandedInfo.ExtractedFilePath
	}

	// create the dataset constructors for downstream processing
	var ds task.DatasetConstructor
	var err error
	var multiBandImageGroup map[string]interface{}
	var geoBoundsGroup map[string]interface{}
	if util.IsDatasetDir(datasetPath) {
		ds, err = dataset.NewD3MDataset(datasetName, datasetPath)
	} else if rawDatasetIsTabular(datasetPath) {
		ds, err = createTableDataset(datasetPath, datasetName)
	} else {
		// check to see what type of files it contains
		var fileType string
		fileType, err = dataset.CheckFileType(datasetPath)
		if err != nil {
			return nil, nil, err
		}
		if fileType == "png" || fileType == "jpeg" || fileType == "jpg" {
			ds, err = createMediaDataset(datasetName, fileType, datasetPath)
		} else if fileType == "tif" {
			ds, err = createRemoteSensingDataset(datasetName, fileType, datasetPath)
			multiBandImageGroup = dataset.CreateSatelliteGrouping()
			geoBoundsGroup = dataset.CreateGeoBoundsGrouping()
		} else if fileType == "txt" {
			ds, err = createTextDataset(datasetName, datasetPath)
		} else {
			err = errors.Errorf("unsupported archived file type %s", fileType)
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return ds, []map[string]interface{}{multiBandImageGroup, geoBoundsGroup}, nil
}

func rawDatasetIsTabular(datasetPath string) bool {
	// check if datasetPath is a folder
	if !util.FileExists(datasetPath) || util.IsDirectory(datasetPath) {
		return false
	}

	// check if it is a csv file
	return path.Ext(datasetPath) == ".csv"
}

func createTableDataset(datasetPath string, datasetName string) (task.DatasetConstructor, error) {
	data, err := ioutil.ReadFile(datasetPath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read raw tabular data")
	}

	ds, err := dataset.NewTableDataset(datasetName, data, true)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func createTextDataset(datasetName string, extractedFilePath string) (task.DatasetConstructor, error) {
	ds, err := dataset.NewMediaDatasetFromExpanded(datasetName, "txt", "txt", "", extractedFilePath)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func createMediaDataset(datasetName string, imageType string, extractedFilePath string) (task.DatasetConstructor, error) {
	ds, err := dataset.NewMediaDatasetFromExpanded(datasetName, imageType, "jpeg", "", extractedFilePath)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func createRemoteSensingDataset(datasetName string, imageType string, extractedFilePath string) (task.DatasetConstructor, error) {
	ds, err := dataset.NewSatelliteDatasetFromExpanded(datasetName, imageType, "", extractedFilePath)
	if err != nil {
		return nil, err
	}

	return ds, nil
}
