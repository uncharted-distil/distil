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
	"net/http"
	"path"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-ingest/pkg/metadata"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
)

// PredictionsHandler receives a file and produces results using the specified
// fitted solution id
func PredictionsHandler(outputPath string, dataStorageCtor api.DataStorageCtor, solutionStorageCtor api.SolutionStorageCtor,
	metaStorageCtor api.MetadataStorageCtor, config *env.Config, ingestConfig *task.IngestTaskConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		fittedSolutionID := pat.Param(r, "fitted-solution-id")

		// read the file from the request
		data, err := receiveFile(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to receive file from request"))
			return
		}
		log.Infof("received data to use for predictions for dataset %s solution %s", dataset, fittedSolutionID)

		solutionStorage, err := solutionStorageCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to initialize solution storage"))
			return
		}
		metaStorage, err := metaStorageCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to initialize metadata storage"))
			return
		}
		dataStorage, err := dataStorageCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to initialize data storage"))
			return
		}

		// get the source dataset from the fitted solution ID
		req, err := solutionStorage.FetchRequestByFittedSolutionID(fittedSolutionID)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to fetch request using fitted solution id"))
			return
		}

		// read the metadata of the original dataset
		datasetES, err := metaStorage.FetchDataset(req.Dataset, false, false)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to fetch dataset from es"))
			return
		}
		schemaPath := path.Join(env.ResolvePath(datasetES.Source, datasetES.Folder), compute.D3MDataSchema)
		meta, err := metadata.LoadMetadataFromOriginalSchema(schemaPath)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to load metadata from source dataset schema doc"))
			return
		}

		err = task.Predict(meta, dataset, fittedSolutionID, data, outputPath, config.ESDatasetsIndex, getTarget(req), metaStorage, dataStorage, solutionStorage, ingestConfig)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to generate predictions"))
			return
		}

		// marshal data and sent the response back
		err = handleJSON(w, map[string]interface{}{"result": "done"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}

func getTarget(request *api.Request) string {
	for _, f := range request.Features {
		if f.FeatureType == "target" {
			return f.FeatureName
		}
	}

	return ""
}
