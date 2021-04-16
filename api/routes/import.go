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
		datasetIDSource := pat.Param(r, "datasetID")
		sourceParsed := metadata.DatasetSource(pat.Param(r, "source"))
		provenance := pat.Param(r, "provenance")
		isSampling := true // Flag to sample imported dataset
		sourceLearningDataset := ""

		ingestSteps := &task.IngestSteps{
			ClassificationOverwrite: false,
			VerifyMetadata:          true,
			FallbackMerged:          true,
			CheckMatch:              true,
			SkipFeaturization:       false,
		}
		ingestParams := &task.IngestParams{
			Source:   sourceParsed,
			DataCtor: dataCtor,
			MetaCtor: esMetaCtor,
			ID:       datasetIDSource,
			Type:     api.DatasetTypeModelling,
		}
		if (ingestParams.Source == metadata.Augmented || ingestParams.Source == metadata.Public) && provenance == "local" {
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

			// one of path or joined dataset is needed
			if params["joinedDataset"] == nil && params["path"] == nil {
				missingParamErr(w, "path or joinedDataset")
				return
			}

			var originalDataset map[string]interface{}
			var joinedDataset map[string]interface{}
			if params["joinedDataset"] != nil {
				if params["originalDataset"] == nil {
					missingParamErr(w, "originalDataset")
					return
				}

				// set the origin information
				var okOriginal, okJoined bool
				originalDataset, okOriginal = params["originalDataset"].(map[string]interface{})
				joinedDataset, okJoined = params["joinedDataset"].(map[string]interface{})
				if okOriginal && okJoined {
					// make sure only one of the datasets has a prefeaturized version
					originalLearningDataset := originalDataset["learningDataset"].(string)
					joinedLearningDataset := joinedDataset["learningDataset"].(string)
					if originalLearningDataset != "" && joinedLearningDataset != "" {
						handleError(w, errors.Errorf("joining datasets that both have learning datasets is not supported"))
						return
					} else if originalLearningDataset != "" {
						ingestSteps.SkipFeaturization = true
						sourceLearningDataset = originalLearningDataset
					} else if joinedLearningDataset != "" {
						ingestSteps.SkipFeaturization = true
						sourceLearningDataset = joinedLearningDataset
					}

					// combine the origin and joined dateset into an array of structs
					origins, err := getOriginsFromMaps(originalDataset, joinedDataset)
					if err != nil {
						handleError(w, errors.Wrap(err, "unable to marshal dataset origins from JSON to struct"))
						return
					}
					ingestParams.Origins = origins
				}
			}

			if params["path"] != nil {
				datasetPathRaw := params["path"].(string)
				log.Infof("Creating dataset '%s' from '%s'", ingestParams.ID, datasetPathRaw)
				creationResult, err := createDataset(datasetPathRaw, ingestParams.ID, config)
				if err != nil {
					handleError(w, errors.Wrap(err, "unable to create raw dataset"))
					return
				}
				ingestParams.ID = creationResult.name
				ingestParams.Path = creationResult.path
				ingestParams.RawGroupings = creationResult.groups
				ingestParams.IndexFields = creationResult.indexFields
				ingestSteps.VerifyMetadata = false

				if originalDataset != nil {
					// if no groups were created, copy them from the passed in datasets
					if len(ingestParams.RawGroupings) == 0 {
						log.Infof("copying groupings from source datasets")
						groups, err := getDatasetGroups(originalDataset)
						if err != nil {
							handleError(w, errors.Wrap(err, "unable to get original dataset groups"))
							return
						}
						groupsJoin, err := getDatasetGroups(joinedDataset)
						if err != nil {
							handleError(w, errors.Wrap(err, "unable to get joining dataset groups"))
							return
						}
						ingestParams.RawGroupings = append(groups, groupsJoin...)
					}

					// set the definitive types based on the currently stored metadata
					metaStore, err := esMetaCtor()
					if err != nil {
						handleError(w, errors.Wrap(err, "unable to create metadata storage"))
						return
					}
					definitiveVars := append(getVariablesDefault(originalDataset["id"].(string), metaStore), getVariablesDefault(joinedDataset["id"].(string), metaStore)...)
					ingestParams.DefinitiveTypes = api.MapVariables(definitiveVars, func(variable *model.Variable) string { return variable.Key })
				} else {
					ingestParams.DefinitiveTypes = api.MapVariables(creationResult.definitiveVars, func(variable *model.Variable) string { return variable.Key })
				}
				log.Infof("Created dataset '%s' from local source '%s'", ingestParams.ID, ingestParams.Path)
			}
		}

		// If the source is Public, the dataset has been imported in the augmented folder,
		// from now on, the ES and database ingestion are done from the augmented folder files.
		if ingestParams.Source == metadata.Public {
			ingestParams.Source = metadata.Augmented
		}

		meta, err := createMetadataStorageForSource(ingestParams.Source, provenance, datamartCtors, fileMetaCtor, esMetaCtor)
		if err != nil {
			handleError(w, err)
			return
		}

		// import the dataset to the local filesystem.
		if ingestParams.Path == "" {
			ingestParams.Path = env.ResolvePath(ingestParams.Source, ingestParams.ID)
		}
		log.Infof("Importing dataset '%s' from '%s'", ingestParams.ID, ingestParams.Path)
		_, err = meta.ImportDataset(ingestParams.ID, ingestParams.Path)
		if err != nil {
			handleError(w, err)
			return
		}

		// ingest the imported dataset
		ingestConfig := task.NewConfig(*config)

		// Check if the imported dataset should be sampled
		if !isSampling {
			ingestConfig.SampleRowLimit = math.MaxInt32 // Maximum int value.
		}

		log.Infof("Ingesting dataset '%s'", ingestParams.Path)
		ingestResult, err := task.IngestDataset(ingestParams, ingestConfig, ingestSteps)
		if err != nil {
			handleError(w, err)
			return
		}

		// if there is a source learning dataset, then sync it properly with the newly imported dataset
		// this will occur when the import is of a joined dataset
		err = syncPrefeaturizedDataset(ingestResult.DatasetID, ingestParams.Path, sourceLearningDataset, esMetaCtor)
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
	name           string
	path           string
	groups         []map[string]interface{}
	indexFields    []string
	definitiveVars []*model.Variable
}

func createDataset(datasetPath string, datasetName string, config *env.Config) (*datasetCreationResult, error) {
	// create the raw dataset
	log.Infof("Creating raw dataset '%s' from '%s'", datasetName, datasetPath)
	ds, groups, indexFields, err := createRawDataset(datasetPath, datasetName)
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
		name:           datasetName,
		path:           formattedPath,
		groups:         groups,
		indexFields:    indexFields,
		definitiveVars: ds.GetDefinitiveTypes(),
	}, nil
}

func createRawDataset(datasetPath string, datasetName string) (task.DatasetConstructor, []map[string]interface{}, []string, error) {
	if util.IsArchiveFile(datasetPath) {
		// expand the archive
		expandedInfo, err := dataset.ExpandZipDataset(datasetPath, datasetName)
		if err != nil {
			return nil, nil, nil, err
		}

		datasetPath = expandedInfo.ExtractedFilePath
	}

	// create the dataset constructors for downstream processing
	var ds task.DatasetConstructor
	var err error
	var groups []map[string]interface{}
	var indexFields []string
	if util.IsDatasetDir(datasetPath) {
		ds, err = dataset.NewD3MDataset(datasetName, datasetPath)
	} else if rawDatasetIsTabular(datasetPath) {
		ds, err = createTableDataset(datasetPath, datasetName)
	} else {
		// check to see what type of files it contains
		var fileType string
		fileType, err = dataset.CheckFileType(datasetPath)
		if err != nil {
			return nil, nil, nil, err
		}
		if fileType == "png" || fileType == "jpeg" || fileType == "jpg" {
			ds, err = createMediaDataset(datasetName, fileType, datasetPath)
		} else if fileType == "tif" {
			ds, err = createRemoteSensingDataset(datasetName, fileType, datasetPath)
			groups = []map[string]interface{}{dataset.CreateSatelliteGrouping(), dataset.CreateGeoBoundsGrouping()}
			indexFields = dataset.GetSatelliteIndexFields()
		} else if fileType == "txt" {
			ds, err = createTextDataset(datasetName, datasetPath)
		} else {
			err = errors.Errorf("unsupported archived file type %s", fileType)
		}
	}
	if err != nil {
		return nil, nil, nil, err
	}

	return ds, groups, indexFields, nil
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

func getDatasetGroups(dsRaw map[string]interface{}) ([]map[string]interface{}, error) {
	// cycle through variables, pulling groups out as they are found
	groups := []map[string]interface{}{}
	for _, v := range dsRaw["variables"].([]interface{}) {
		rawVariables := v.(map[string]interface{})
		if rawVariables["grouping"] != nil {
			groups = append(groups, rawVariables["grouping"].(map[string]interface{}))
		}
	}

	return groups, nil
}

func getVariablesDefault(datasetID string, metaStorage api.MetadataStorage) []*model.Variable {
	ds, err := metaStorage.FetchDataset(datasetID, true, true, true)
	if err != nil {
		log.Infof("unable to fetch variables so defaulting to empty list")
		return []*model.Variable{}
	}

	return ds.Variables
}

func syncPrefeaturizedDataset(datasetID string, sourceDataset string, sourceLearningDataset string, metaCtor api.MetadataStorageCtor) error {
	if sourceLearningDataset == "" {
		return nil
	}

	metaStorage, err := metaCtor()
	if err != nil {
		return err
	}

	ds, err := metaStorage.FetchDataset(datasetID, true, true, true)
	if err != nil {
		return err
	}

	// TODO: CHECK UNIQUENESS!!!!
	joinedLearningDataset := task.CreateFeaturizedDatasetID(datasetID)
	joinedLearningDataset = env.ResolvePath(ds.Source, joinedLearningDataset)

	dsDisk, err := task.CopyDiskDataset(sourceLearningDataset, joinedLearningDataset, ds.ID, ds.StorageName)
	if err != nil {
		return err
	}

	// sync the dataset on disk
	// read the unfeaturized dataset from disk
	dsDiskSource, err := api.LoadDiskDatasetFromFolder(sourceDataset)
	if err != nil {
		return err
	}

	// update the featurized dataset using the unfeaturized data
	err = dsDisk.UpdateOnDisk(ds, dsDiskSource.Dataset.Data, true)
	if err != nil {
		return err
	}

	// update the metadata in ES
	ds.LearningDataset = joinedLearningDataset
	err = metaStorage.UpdateDataset(ds)
	if err != nil {
		return err
	}

	return nil
}
