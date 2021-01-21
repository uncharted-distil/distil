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
	"github.com/uncharted-distil/distil-compute/model"
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
		predictionRequestID, err := url.PathUnescape(pat.Param(r, "produce-request-id"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape produce request id"))
			return
		}

		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to parse post parameters"))
			return
		}
		if params == nil {
			missingParamErr(w, "parameters")
			return
		}
		if params["datasetName"] == nil {
			missingParamErr(w, "datasetName")
			return
		}
		newDatasetName := params["datasetName"].(string)

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
		targetName := ""
		features := []string{model.D3MIndexFieldName}
		for _, f := range request.Features {
			if f.FeatureType != model.FeatureTypeTarget {
				features = append(features, f.FeatureName)
			} else {
				targetName = f.FeatureName
			}
		}

		// get needed request info
		pred, err := solutionStorage.FetchPredictionResultByProduceRequestID(predictionRequestID)
		if err != nil {
			handleError(w, err)
			return
		}

		// read the source metadata for typing information
		req, err := solutionStorage.FetchRequestByFittedSolutionID(prediction.FittedSolutionID)
		if err != nil {
			handleError(w, err)
			return
		}

		newDatasetID, err := task.CreateDatasetFromResult(newDatasetName, prediction.Dataset, req.Dataset, features, targetName, pred.ResultURI, metaStorage, dataStorage, config)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, map[string]interface{}{"success": true, "newDatasetID": newDatasetID})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal clustering result into JSON"))
			return
		}
	}
}
