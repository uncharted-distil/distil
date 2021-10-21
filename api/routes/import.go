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
	"math"
	"net/http"
	"path"
	"strings"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/task/importer"
	"github.com/uncharted-distil/distil/api/util"
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
		datasetDescription := ""
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

		esMetaStorage, err := esMetaCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to initialize metadata storage connection"))
			return
		}
		ingestParamsOriginal := &task.IngestParams{
			Source:   sourceParsed,
			DataCtor: dataCtor,
			MetaCtor: esMetaCtor,
			ID:       datasetIDSource,
			Type:     api.DatasetTypeModelling,
		}

		imp := getImporter(provenance, params, esMetaStorage, config)
		err = imp.Initialize(params, ingestParamsOriginal)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to initialize import"))
			return
		}

		ingestSteps, ingestParams, err := imp.PrepareImport()
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to prepare import"))
			return
		}

		if sourceParsed == metadata.Public {
			sourceParsed = metadata.Augmented
		}
		ingestParams.Source = sourceParsed
		ingestParams.DataCtor = dataCtor
		ingestParams.MetaCtor = esMetaCtor
		ingestParams.ID = datasetIDSource
		ingestParams.Type = api.DatasetTypeModelling

		meta, err := createMetadataStorageForSource(ingestParams.Source, provenance, datamartCtors, fileMetaCtor, esMetaCtor)
		if err != nil {
			handleError(w, err)
			return
		}

		log.Infof("Importing dataset '%s' from '%s'", ingestParams.ID, ingestParams.Path)
		dsPath, err := meta.ImportDataset(ingestParams.ID, ingestParams.Path)
		if err != nil {
			handleError(w, err)
			return
		}
		// update dataset description
		if datasetDescription != "" {
			ds, err := api.LoadDiskDatasetFromFolder(dsPath)
			if err != nil {
				handleError(w, err)
				return
			}
			ds.Dataset.Metadata.Description = datasetDescription
			err = ds.SaveDataset()
			if err != nil {
				handleError(w, err)
				return
			}
		}
		// ingest the imported dataset
		ingestConfig := task.NewConfig(*config)

		// Check if the imported dataset should be sampled
		if !isSampling {
			ingestConfig.SampleRowLimit = math.MaxInt32 // Maximum int value.
		}

		err = moveResources(ingestParams.GetSchemaDocPath())
		if err != nil {
			handleError(w, err)
			return
		}
		log.Infof("Ingesting dataset '%s'", ingestParams.Path)
		ingestResult, err := task.IngestDataset(ingestParams, ingestConfig, ingestSteps)
		if err != nil {
			handleError(w, err)
			return
		}

		err = imp.CleanupImport(ingestResult)
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

func getImporter(provenance string, params map[string]interface{}, esMetaStorage api.MetadataStorage, config *env.Config) importer.Importer {
	if provenance != "local" {
		return importer.NewDatamart()
	}

	if params["joinedDataset"] != nil {
		return importer.NewJoined(esMetaStorage, config)
	}

	return importer.NewLocal(config)
}

func moveResources(schemaDoc string) error {
	log.Infof("checking to see if any data resources in the dataset found at '%s' need to be moved to the resource folder", schemaDoc)
	// read the dataset from disk
	dsFolder := path.Dir(schemaDoc)
	dsDisk, err := api.LoadDiskDatasetFromFolder(dsFolder)
	if err != nil {
		return err
	}

	// any resources not in the resource folder should be moved there
	mainDR := dsDisk.Dataset.Metadata.GetMainDataResource()
	updated := false
	for _, dr := range dsDisk.Dataset.Metadata.DataResources {
		// main data resource should stay in the dataset folder
		if dr != mainDR {
			// move the resource over to the resource folder
			if !util.IsInDirectory(env.GetResourcePath(), dr.ResPath) {
				destinationPathFull := strings.Replace(dr.ResPath, path.Dir(dsFolder), env.GetResourcePath(), 1)
				destinationPath := util.GetUniqueFolder(path.Dir(destinationPathFull))
				destinationPath = path.Join(destinationPath, path.Base(destinationPathFull))
				log.Infof("moving data resource from '%s' to '%s'", dr.ResPath, destinationPath)
				err = util.Move(dr.ResPath, destinationPath)
				if err != nil {
					return err
				}

				log.Infof("updating data resource to point to new resource path")
				dr.ResPath = destinationPath
				updated = true
			}
		}
	}

	if updated {
		log.Infof("updating metadata on disk to point to the right resource path")
		err = dsDisk.SaveMetadata()
		if err != nil {
			return err
		}
	}

	log.Infof("all data resources now located in the proper folders")

	return nil
}
