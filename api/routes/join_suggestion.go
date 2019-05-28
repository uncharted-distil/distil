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
	"net/url"
	"time"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"goji.io/pat"

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

		res, err := storage.FetchDataset(dataset, false, false)
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
			if provenance == datamart.ProvenanceISI || provenance == datamart.ProvenanceNYU {
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

			datasets = append(datasets, datasetsPart...)
		}

		// imported datasets override non-imported datasets
		exists := make(map[string]*model.Dataset)
		for _, dataset := range datasets {
			if !hasSuggestions(dataset) {
				continue
			}

			existing, ok := exists[dataset.ID]
			if !ok {
				// we don't have it, add it
				exists[dataset.ID] = dataset
			} else {
				// we already have it, if it is `dataset`, replace it
				if existing.Provenance == datamart.ProvenanceNYU || existing.Provenance == datamart.ProvenanceISI {
					exists[dataset.ID] = dataset
				}
			}
		}

		var deconflicted []*model.Dataset
		for _, dataset := range exists {
			deconflicted = append(deconflicted, dataset)
		}

		// marshal data
		err = handleJSON(w, DatasetsResult{
			Datasets: deconflicted,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}

func hasSuggestions(dataset *model.Dataset) bool {
	return dataset.JoinSuggestions != nil && len(dataset.JoinSuggestions) > 0
}
