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

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util/json"
)

// ImportHandler imports a dataset to the local file system and then ingests it.
func ImportHandler(dataCtor api.DataStorageCtor, datamartCtors map[string]api.MetadataStorageCtor, fileMetaCtor api.MetadataStorageCtor, esMetaCtor api.MetadataStorageCtor, config *task.IngestTaskConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetID := pat.Param(r, "datasetID")
		source := metadata.DatasetSource(pat.Param(r, "source"))
		provenance := pat.Param(r, "provenance")

		var origins []*model.DatasetOrigin
		if source == metadata.Augmented && provenance != "local" {
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
					handleError(w, errors.Wrap(err, "unable marshal dataset origins from JSON to struct"))
					return
				}
			}
		}

		meta, err := createMetadataStorageForSource(source, provenance, datamartCtors, fileMetaCtor, esMetaCtor)
		if err != nil {
			handleError(w, err)
			return
		}

		// import the dataset to the local filesystem.
		uri := env.ResolvePath(source, datasetID)

		ingestConfig := *config

		_, err = meta.ImportDataset(datasetID, uri)
		if err != nil {
			handleError(w, err)
			return
		}

		// ingest the imported dataset
		ingestSteps := &task.IngestSteps{ClassificationOverwrite: false}
		datasetID, err = task.IngestDataset(source, dataCtor, esMetaCtor, datasetID, origins, api.DatasetTypeModelling, &ingestConfig, ingestSteps)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data and sent the response back
		err = handleJSON(w, map[string]interface{}{"dataset": datasetID, "result": "ingested"})
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

func getDatasetOrigins(esStorage api.MetadataStorage, dataset string) ([]*model.DatasetOrigin, error) {
	ds, err := esStorage.FetchDataset(dataset, true, true)
	if err != nil {
		return nil, err
	}

	if ds.JoinSuggestions == nil {
		return make([]*model.DatasetOrigin, 0), nil
	}

	origins := make([]*model.DatasetOrigin, len(ds.JoinSuggestions))
	for i, js := range ds.JoinSuggestions {
		origins[i] = &model.DatasetOrigin{
			SearchResult:  js.DatasetOrigin.SearchResult,
			Provenance:    js.DatasetOrigin.Provenance,
			SourceDataset: dataset,
		}
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
