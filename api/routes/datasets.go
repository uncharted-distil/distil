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
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"github.com/russross/blackfriday"
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/model/storage/datamart"
	"github.com/uncharted-distil/distil/api/util"
)

const (
	searchTimeout               = 10
	joinSuggestionSearchTimeout = 600
)

// DatasetResult represents the result of a dataset response.
type DatasetResult struct {
	Dataset *model.Dataset `json:"dataset"`
}

// DatasetsResult represents the result of a datasets response.
type DatasetsResult struct {
	Datasets []*model.Dataset `json:"datasets"`
}

// DatasetHandler generates a route handler that returns a specified dataset summary.
func DatasetHandler(ctor model.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// get dataset name
		dataset := pat.Param(r, "dataset")

		// get metadata client
		storage, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}

		// get dataset summary
		res, err := storage.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data
		err = handleJSON(w, DatasetResult{
			Dataset: res,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to marshal dataset result into JSON"))
			return
		}
	}
}

// DatasetsHandler generates a route handler that facilitates a search of
// dataset descriptions and variable names, returning a name, description and
// variable list for any dataset that matches. The search parameter is optional
// it contains the search terms if set, and if unset, flags that a list of all
// datasets should be returned.  The full list will be contain names only,
// descriptions and variable lists will not be included.
func DatasetsHandler(metaCtors map[string]model.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var datasets []*model.Dataset
		// check for search terms
		terms, err := url.QueryUnescape(r.URL.Query().Get("search"))
		if err != nil {
			handleError(w, errors.Wrap(err, "Malformed datasets query"))
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
			go loadDatasets(storage, terms, nil, results, errors)
			select {
			case res := <-results:
				datasetsPart = res
			case err = <-errors:
				//handleError(w, err)
				log.Warnf("error querying dataset: %v", err)
				continue
			case <-time.After(searchTimeout * time.Second):
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
			handleError(w, errors.Wrap(err, "unable to marshal dataset results into JSON"))
			return
		}
	}
}

// AvailableDatasetsHandler generates a route handle that will return the list of
// files & folders that can be imported.
func AvailableDatasetsHandler(metaCtor model.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// restrict to augmented folder
		rootFolder := env.GetAugmentedPath()
		files, err := ioutil.ReadDir(rootFolder)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to read augmented folder contents"))
			return
		}

		// get the existing dataset folders
		meta, err := metaCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to create metadata storage"))
			return
		}

		datasets, err := meta.FetchDatasets(false, false)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to fetch existing datasets"))
			return
		}

		existingFolders := []string{}
		for _, ds := range datasets {
			existingFolders = append(existingFolders, ds.Folder)
		}

		// return the path with the name of the available file.
		type AvailableFile struct {
			Name string `json:"name"`
			Path string `json:"path"` // TODO: this is a security issue.
		}

		// folders could be datasets that are already imported
		available := []AvailableFile{}
		for _, f := range files {
			if f.IsDir() {
				if isAvailableForImport(path.Join(rootFolder, f.Name()), existingFolders) {
					available = append(available, AvailableFile{f.Name(), rootFolder})
				}
			} else {
				if !strings.HasPrefix(f.Name(), ".") { // Hide dot files
					available = append(available, AvailableFile{f.Name(), rootFolder})
				}
			}
		}

		// marshal data
		err = handleJSON(w, map[string][]AvailableFile{
			"availableDatasets": available,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to marshal dataset results into JSON"))
			return
		}
	}
}

func isAvailableForImport(folderPath string, existingDatasets []string) bool {
	// if it is in D3M format, it could already be ingested
	if util.IsDatasetDir(folderPath) {
		for _, ds := range existingDatasets {
			if path.Base(folderPath) == ds {
				return false
			}
		}
	}

	return true
}

func loadDatasets(storage model.MetadataStorage, terms string, baseDataset *model.Dataset, results chan []*model.Dataset, errors chan error) {
	// if its present, forward a search, otherwise fetch all datasets
	var datasetsPart []*model.Dataset
	var err error

	if terms != "" || baseDataset != nil {
		datasetsPart, err = storage.SearchDatasets(terms, baseDataset, false, false)

		// combine the suggestions by dataset since the datamarts may break them up
		joinedSuggestions := make(map[string]*model.Dataset)
		for _, ds := range datasetsPart {
			existingDataset := joinedSuggestions[ds.Name]
			if existingDataset == nil {
				joinedSuggestions[ds.Name] = ds
			} else {
				// merge the suggestions while keeping the highest score
				if ds.JoinScore > existingDataset.JoinScore {
					existingDataset.JoinScore = ds.JoinScore
				}
				for _, js := range ds.JoinSuggestions {
					js.Index = js.Index + len(existingDataset.JoinSuggestions)
				}
				existingDataset.JoinSuggestions = append(existingDataset.JoinSuggestions, ds.JoinSuggestions...)
			}
		}

		// write out to datasetsPart
		datasetsPart = make([]*model.Dataset, 0)
		for _, ds := range joinedSuggestions {
			datasetsPart = append(datasetsPart, ds)
		}
	} else {
		datasetsPart, err = storage.FetchDatasets(false, false)
	}

	if err != nil {
		errors <- err
	} else {
		results <- datasetsPart
	}
}

func renderMarkdown(markdown string) string {
	// process the markdown into HTML
	unsafe := blackfriday.Run([]byte(markdown))
	// just to be safe, sanatize the HTML
	return string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
}
