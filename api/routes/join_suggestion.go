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
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"

	compute "github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/model/storage/datamart"
)

// JoinSuggestionHandler generates a route handler that facilitates a search of
// dataset join suggestions. The search parameter is optional
// it contains the search terms if set, and if unset, flags that a list of all
// datasets should be returned.  The full list will be contain names only,
// descriptions and variable lists will not be included.
func JoinSuggestionHandler(esCtor model.MetadataStorageCtor, metaCtors map[string]model.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var datasets []*model.Dataset
		datasetsMap := make(map[string][]*model.Dataset)
		// get dataset name
		dataset := pat.Param(r, "dataset")

		// check for search terms
		terms, err := url.QueryUnescape(r.URL.Query().Get("search"))
		if err != nil {
			handleError(w, errors.Wrap(err, "Malformed datasets query"))
			return
		}

		// pull the dataset info for matching
		storage, err := esCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		res, err := storage.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		for provenance, ctor := range metaCtors {
			// get metadata client
			storage, err := ctor()
			if err != nil {
				handleError(w, err)
				return
			}

			// use a timeout in case the search hangs
			results := make(chan []*model.Dataset, 1)
			errors := make(chan error, 1)
			var datasetsPart []*model.Dataset
			var baseDataset *model.Dataset
			// provide base dataset for ISI and NYU datamart
			isDatamart := provenance == datamart.ProvenanceISI || provenance == datamart.ProvenanceNYU
			if isDatamart {
				baseDataset = res
			}
			go loadDatasets(storage, terms, baseDataset, results, errors)
			select {
			case res := <-results:
				datasetsPart = res
			case err = <-errors:
				//handleError(w, err)
				log.Warnf("error querying dataset: %v", err)
				continue
			case <-time.After(joinSuggestionSearchTimeout * time.Second):
				log.Warnf("timeout querying dataset from %s", provenance)
				datasetsPart = make([]*model.Dataset, 0)
			}

			// render dataset description as HTML
			for _, dataset := range datasetsPart {
				dataset.Description = renderMarkdown(dataset.Description)
			}
			datasetsMap[provenance] = datasetsPart
		}

		datamartDatasets := append(datasetsMap[datamart.ProvenanceISI], datasetsMap[datamart.ProvenanceNYU]...)
		datasets = filterDatasets(res, datamartDatasets, filterSuggestions)

		localDatasets := make(map[string]*model.Dataset)
		for provenance, datasets := range datasetsMap {
			if provenance == datamart.ProvenanceISI || provenance == datamart.ProvenanceNYU {
				continue
			}
			for _, dataset := range datasets {
				localDatasets[dataset.ID] = dataset
			}
		}

		// If a dataset already exists in the local, use the local dataset augmented with join suggestions from the corresponding datamart dataset
		// Note: there could be multiple nyu datamart result with same dataset id with diffrent join suggestions/score
		for i := 0; i < len(datasets); i++ {
			dataset := datasets[i]
			if locDataset, ok := localDatasets[dataset.ID]; ok {
				// make a copy of local dataset
				localDataset := *locDataset
				localDataset.JoinScore = datasets[i].JoinScore
				localDataset.JoinSuggestions = datasets[i].JoinSuggestions

				// Column names are normalized while dataset is being ingest
				// So we retreive normalized column names from the local dataset
				// using original names from the datamart dataset
				for _, suggestion := range localDataset.JoinSuggestions {
					for j, colName := range suggestion.JoinColumns {
						colNameTokens := strings.Split(colName, ", ")
						var localColNameTokens []string
						for _, token := range colNameTokens {
							// displayName holds the column name of the original datamart dataset
							localColNameTokens = append(localColNameTokens, getColKeyByDisplayName(localDataset, token))
						}
						localColName := strings.Join(localColNameTokens, ", ")
						suggestion.JoinColumns[j] = localColName
					}
				}
				// Some imported local datasets are missing description. In that case, add a description
				// (I guess this happens because a downloaded dataset from datamart which is being imported and ingested to the local system,
				// sometimes comes with datasetDoc.json file with no description in it)
				if localDataset.Description == "" {
					localDataset.Description = datasets[i].Description
				}
				datasets[i] = &localDataset
			}
		}

		// sort by join score and name
		sort.Slice(datasets, func(i, j int) bool {
			if datasets[i].JoinScore == datasets[j].JoinScore {
				return datasets[i].Name < datasets[j].Name
			}
			return datasets[i].JoinScore > datasets[j].JoinScore
		})

		err = handleJSON(w, DatasetsResult{
			Datasets: datasets,
		})

		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}

func getColKeyByDisplayName(dataset model.Dataset, colDisplayName string) string {
	for _, variable := range dataset.Variables {
		if variable.DisplayName == colDisplayName {
			return variable.Key
		}
	}
	return compute.NormalizeVariableName(colDisplayName) // fallback
}

func filterDatasets(queryDataset *model.Dataset, datasets []*model.Dataset,
	predicate func(sourceDataset *model.Dataset, joinDataset *model.Dataset) bool) []*model.Dataset {
	result := []*model.Dataset{}
	for _, dataset := range datasets {
		if predicate(queryDataset, dataset) {
			result = append(result, dataset)
		}
	}
	return result
}

func filterSuggestions(sourceDataset *model.Dataset, joinDataset *model.Dataset) bool {
	// filter out datasets that have already been joined
	if strings.Contains(sourceDataset.ID, joinDataset.ID) {
		return false
	}

	// filter out join datasets with no suggestions
	return joinDataset.JoinSuggestions != nil && len(joinDataset.JoinSuggestions) > 0
}
