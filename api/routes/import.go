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
	"goji.io/pat"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-ingest/metadata"
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

		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		// get elasticsearch client
		esStorage, err := esMetaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// set the origin information
		var origins []*model.DatasetOrigin
		originalDatasetID, ok := params["originalDatasetID"].(string)
		joinedDatasetID, ok := params["joinedDatasetID"].(string)
		if ok {
			searchResultIndexF, ok := params["searchResultIndex"].(float64)
			if !ok {
				handleError(w, errors.Errorf("Search result index needed for joined dataset import"))
				return
			}
			searchResultIndex := int(searchResultIndexF)

			// get the joined dataset for the search result
			joinedDataset, err := esStorage.FetchDataset(joinedDatasetID, true, true)
			if err != nil {
				handleError(w, err)
				return
			}

			// add the joining origin to the source dataset joining
			origins, err = getDatasetOrigins(esStorage, originalDatasetID)
			if err != nil {
				handleError(w, err)
				return
			}
			origins = append(origins, &model.DatasetOrigin{
				SearchResult:  joinedDataset.JoinSuggestions[searchResultIndex].DatasetOrigin.SearchResult,
				Provenance:    joinedDataset.JoinSuggestions[searchResultIndex].DatasetOrigin.Provenance,
				SourceDataset: originalDatasetID,
			})
		}

		// multiple search results would be received if a single dataset has
		// multiple join suggestions
		searchResults, ok := json.Array(params, "searchResults")
		if ok {
			origins = api.ParseDatasetOriginsFromJSON(searchResults)
		}

		// update ingest config to use ingest URI.
		cfg, err := env.LoadConfig()
		if err != nil {
			handleError(w, err)
			return
		}

		meta, err := createMetadataStorageForSource(source, provenance, datamartCtors, fileMetaCtor, esMetaCtor)
		if err != nil {
			handleError(w, err)
			return
		}

		// import the dataset to the local filesystem.
		uri := env.ResolvePath(source, datasetID)

		ingestConfig := *config
		ingestConfig.SummaryEnabled = false

		_, err = meta.ImportDataset(datasetID, uri)
		if err != nil {
			handleError(w, err)
			return
		}

		// ingest the imported dataset
		err = task.IngestDataset(source, dataCtor, esMetaCtor, cfg.ESDatasetsIndex, datasetID, origins, &ingestConfig)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data and sent the response back
		err = handleJSON(w, map[string]interface{}{"result": "ingested"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
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
