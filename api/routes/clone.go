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
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util"
)

// CloningHandler generates a route handler that enables cloning
// of a dataset in the data storage and metadata storage.
func CloningHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, config env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")

		metaStorage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		dataStorage, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		ds, err := metaStorage.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		if ds.Clone {
			handleError(w, errors.New("Cannot make a clone of a clone"))
			return
		}
		datasetClone, err := task.GetUniqueOutputFolder(fmt.Sprintf("%s_clone", dataset), env.GetAugmentedPath())
		if err != nil {
			handleError(w, err)
			return
		}
		folderExisting := env.ResolvePath(ds.Source, ds.Folder)
		folderClone := env.ResolvePath(metadata.Augmented, datasetClone)
		storageNameClone, err := dataStorage.GetStorageName(datasetClone)
		if err != nil {
			handleError(w, err)
			return
		}

		err = metaStorage.CloneDataset(dataset, datasetClone, storageNameClone, datasetClone)
		if err != nil {
			handleError(w, err)
			return
		}

		err = dataStorage.CloneDataset(dataset, ds.StorageName, datasetClone, storageNameClone)
		if err != nil {
			handleError(w, err)
			return
		}

		// TEMP FIX: COPY EXISTING DATASET FOLDER FOR NEW DATASET AND UPDATE THE ID
		err = util.Copy(folderExisting, folderClone)
		if err != nil {
			handleError(w, err)
			return
		}
		schemaPath := path.Join(folderClone, compute.D3MDataSchema)
		meta, err := metadata.LoadMetadataFromOriginalSchema(schemaPath, false)
		if err != nil {
			handleError(w, err)
			return
		}
		meta.ID = datasetClone
		writer := serialization.GetStorage(meta.GetMainDataResource().ResPath)
		err = writer.WriteMetadata(schemaPath, meta, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, map[string]interface{}{"success": true, "clonedDatasetName": meta.ID})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal clustering result into JSON"))
			return
		}
	}
}

// CloningResultsHandler generates a route handler that enables cloning
// of a result + dataset in the data storage and metadata storage, creating
// a new dataset based on results.
func CloningResultsHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, solutionCtor api.SolutionStorageCtor, config env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		newDatasetName, err := url.PathUnescape(pat.Param(r, "dataset-name"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape dataset name"))
			return
		}
		predictionRequestID, err := url.PathUnescape(pat.Param(r, "produce-request-id"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape produce request id"))
			return
		}
		metaStorage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		dataStorage, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		solutionStorage, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// get the features from the request
		prediction, err := solutionStorage.FetchPrediction(predictionRequestID)
		if err != nil {
			handleError(w, err)
			return
		}
		request, err := solutionStorage.FetchRequestByFittedSolutionID(prediction.FittedSolutionID)
		if err != nil {
			handleError(w, err)
			return
		}
		features := []string{}
		for _, f := range request.Features {
			features = append(features, f.FeatureName)
		}

		// get needed request info
		pred, err := solutionStorage.FetchPredictionResultByProduceRequestID(predictionRequestID)
		if err != nil {
			handleError(w, err)
			return
		}
		predictionDS, err := metaStorage.FetchDataset(prediction.Dataset, true, true, true)
		if err != nil {
			handleError(w, err)
			return
		}

		// extract the data from the database (result + base)
		data, err := dataStorage.FetchResultDataset(prediction.Dataset, predictionDS.StorageName, newDatasetName, features, pred.ResultURI)
		if err != nil {
			handleError(w, err)
			return
		}

		// read the prediction DS metadata from disk for the new dataset
		predictionDSDatasetPath := env.ResolvePath(metadata.Augmented, path.Join(predictionDS.Folder, compute.D3MDataSchema))
		metaDisk, err := metadata.LoadMetadataFromOriginalSchema(predictionDSDatasetPath, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// store the dataset to disk
		outputPath := env.ResolvePath(metadata.Augmented, newDatasetName)
		writer := serialization.GetStorage(metaDisk.GetMainDataResource().ResPath)

		rawDS := &api.RawDataset{
			ID:              metaDisk.ID,
			Name:            metaDisk.Name,
			Metadata:        metaDisk,
			Data:            data,
			DefinitiveTypes: true,
		}
		err = writer.WriteDataset(outputPath, rawDS)
		if err != nil {
			handleError(w, err)
			return
		}

		// store new dataset metadata
		err = metaStorage.IngestDataset(metadata.Augmented, metaDisk)
		if err != nil {
			handleError(w, err)
			return
		}

		// ingest to postgres from disk
		cloneSchemaPath := path.Join(outputPath, compute.D3MDataSchema)
		err = task.IngestPostgres(cloneSchemaPath, cloneSchemaPath, metadata.Augmented, nil, task.NewConfig(config), false, false, true)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, map[string]interface{}{"success": true, "newDatasetID": metaDisk.ID})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal clustering result into JSON"))
			return
		}
	}
}
