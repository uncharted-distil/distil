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
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
)

// CloningHandler generates a route handler that enables cloning
// of a dataset in the data storage and metadata storage.
func CloningHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, config env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")

		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		// get variable names and ranges out of the params
		var filterParams *api.FilterParams
		if len(params) > 0 {
			filterParams, err = api.ParseFilterParamsFromJSON(params)
			if err != nil {
				handleError(w, err)
				return
			}
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
		folderClone := env.ResolvePath(metadata.Augmented, datasetClone)

		// replace any grouped variables in filter params with the group's
		expandedFilterParams, err := api.ExpandFilterParams(dataset, filterParams, false, metaStorage)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to expand filter params"))
			return
		}

		err = task.CloneDataset(dataset, datasetClone, folderClone, metaStorage, dataStorage, expandedFilterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, map[string]interface{}{"success": true, "clonedDatasetName": datasetClone})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal clustering result into JSON"))
			return
		}
	}
}

type cloningParams struct {
	sourceDataset     string
	predictionDataset string
	targetName        string
	features          []string
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
		includeDatasetFeatures := false
		if params["includeDatasetFeatures"] != nil {
			includeDatasetFeatures = params["includeDatasetFeatures"].(bool)
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
		parsedParams, err := getCloningParams(predictionRequestID, metaStorage, solutionStorage, includeDatasetFeatures)
		if err != nil {
			handleError(w, err)
			return
		}

		// get needed request info
		pred, err := solutionStorage.FetchPredictionResultByProduceRequestID(predictionRequestID)
		if err != nil {
			handleError(w, err)
			return
		}

		newDatasetID, err := task.CreateDatasetFromResult(newDatasetName, parsedParams.predictionDataset,
			parsedParams.sourceDataset, parsedParams.features, parsedParams.targetName, pred.ResultURI, metaStorage, dataStorage, config)
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

func getCloningParams(predictionRequestID string, metaStorage api.MetadataStorage, solutionStorage api.SolutionStorage, includeDatasetFeatures bool) (*cloningParams, error) {
	// get the features from the request
	prediction, err := solutionStorage.FetchPrediction(predictionRequestID)
	if err != nil {
		return nil, err
	}
	request, err := solutionStorage.FetchRequestByFittedSolutionID(prediction.FittedSolutionID)
	if err != nil {
		return nil, err
	}
	targetName := ""
	features := []string{model.D3MIndexFieldName}
	for _, f := range request.Features {
		if !includeDatasetFeatures && f.FeatureType != model.FeatureTypeTarget {
			features = append(features, f.FeatureName)
		} else {
			targetName = f.FeatureName
		}
	}

	// read the source metadata for typing information
	req, err := solutionStorage.FetchRequestByFittedSolutionID(prediction.FittedSolutionID)
	if err != nil {
		return nil, err
	}

	if includeDatasetFeatures {
		// just select every feature in the source dataset
		ds, err := metaStorage.FetchDataset(req.Dataset, true, false, false)
		if err != nil {
			return nil, err
		}

		features = []string{}
		for _, v := range ds.Variables {
			// the target is not a feature!
			// augmented fields should not be included
			if v.Key != targetName && !v.HasRole(model.VarDistilRoleAugmented) {
				features = append(features, v.Key)
			}
		}
	}

	parsedParams := &cloningParams{
		predictionDataset: prediction.Dataset,
		sourceDataset:     req.Dataset,
		targetName:        targetName,
		features:          features,
	}

	return parsedParams, nil
}
