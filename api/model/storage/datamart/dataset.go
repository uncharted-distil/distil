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

package datamart

import (
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
)

const (
	// ProvenanceNYU for NYU datamart
	ProvenanceNYU = "NYU"
	// ProvenanceISI for ISI datamart
	ProvenanceISI = "ISI"
)

// SearchQuery contains the basic properties to query.
type SearchQuery struct {
	Dataset *SearchQueryDatasetProperties `json:"dataset,omitempty"`
}

// SearchQueryDatasetProperties represents queryin on metadata.
type SearchQueryDatasetProperties struct {
	About       string   `json:"about,omitempty"`
	Description []string `json:"description,omitempty"`
	Name        []string `json:"name,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

// ImportDataset makes the dataset available for ingest and returns
// the URI to use for ingest.
func (s *Storage) ImportDataset(id string, uri string) (string, error) {
	return s.download(s, id, uri)
}

// CloneDataset is not supported (ES datasets are already ingested).
func (s *Storage) CloneDataset(dataset string, datasetNew string, storageNameNew string, folderNew string) error {
	return errors.Errorf("Not implemented")
}

// UpdateDataset updates a document consisting of the metadata to the datamart.
func (s *Storage) UpdateDataset(dataset *api.Dataset) error {
	return errors.Errorf("Not implemented")
}

// IngestDataset adds a document consisting of the metadata to the datamart.
func (s *Storage) IngestDataset(datasetSource metadata.DatasetSource, meta *model.Metadata) error {
	return errors.Errorf("Not implemented")
}

// FetchDatasets returns all datasets in the provided index.
func (s *Storage) FetchDatasets(includeIndex bool, includeMeta bool, includeSystemData bool) ([]*api.Dataset, error) {
	// use default string in search to get complete list
	return s.SearchDatasets("", nil, includeIndex, includeMeta, includeSystemData)
}

// DeleteDataset deletes a dataset from the datamart.
func (s *Storage) DeleteDataset(dataset string, softDelete bool) error {
	return errors.Errorf("Not implemented")
}

// DatasetExists returns true if a dataset exists.
func (s *Storage) DatasetExists(dataset string) (bool, error) {
	return false, errors.Errorf("Not implemented")
}

// FetchDataset returns a dataset in the provided index.
func (s *Storage) FetchDataset(datasetName string, includeIndex bool, includeMeta bool, includeSystemData bool) (*api.Dataset, error) {
	return nil, errors.Errorf("Not implemented")
}

// SearchDatasets returns the datasets that match the search criteria in the
// provided index.
func (s *Storage) SearchDatasets(terms string, baseDataset *api.Dataset, includeIndex bool, includeMeta bool, includeSystemData bool) ([]*api.Dataset, error) {
	if terms == "" && baseDataset == nil {
		return make([]*api.Dataset, 0), nil
	}
	return s.searchREST(terms, baseDataset)
}

// SetDataType is not supported by the datamart.
func (s *Storage) SetDataType(dataset string, varName string, varType string) error {
	return errors.Errorf("Not supported")
}

// SetExtrema is not supported by the datamart.
func (s *Storage) SetExtrema(dataset string, varName string, extrema *api.Extrema) error {
	return errors.Errorf("Not supported")
}

// AddVariable is not supported by the datamart.
func (s *Storage) AddVariable(dataset string, varName string, varDisplayName string, varType string, varRole []string) error {
	return errors.Errorf("Not supported")
}

// DeleteVariable is not supported by the datamart.
func (s *Storage) DeleteVariable(dataset string, varName string) error {
	return errors.Errorf("Not supported")
}

// AddGroupedVariable adds a variable grouping.
func (s *Storage) AddGroupedVariable(dataset string, varName string, varDisplayName string, varType string, varRole []string, grouping model.BaseGrouping) error {
	return errors.Errorf("Not supported")
}

// RemoveGroupedVariable removes a variable grouping.
func (s *Storage) RemoveGroupedVariable(datasetName string, grouping model.BaseGrouping) error {
	return errors.Errorf("Not supported")
}

func (s *Storage) searchREST(searchText string, baseDataset *api.Dataset) ([]*api.Dataset, error) {
	terms := strings.Fields(searchText)

	searchQueryDatasetProperties := &SearchQueryDatasetProperties{}
	if len(terms) > 0 {
		searchQueryDatasetProperties = &SearchQueryDatasetProperties{
			//About: searchText,
			//Name:        terms,
			//Description: terms,
			Keywords: []string{searchText},
		}
	}
	// get complete URI for the endpoint
	query := &SearchQuery{
		Dataset: searchQueryDatasetProperties,
	}

	// figure out the data path if a dataset is provided
	dataPath := ""
	if baseDataset != nil {
		datasetPath := env.ResolvePath(baseDataset.Source, baseDataset.Folder)
		meta, err := metadata.LoadMetadataFromOriginalSchema(path.Join(datasetPath, "datasetDoc.json"), true)
		if err != nil {
			return nil, errors.Wrap(err, "unable to load metadatat")
		}

		dr := meta.GetMainDataResource()
		dataPath = path.Join(datasetPath, dr.ResPath)
	}

	env.LogDatamartActionGlobal(s.searchFunction, "DATA_PREPARATION", "DATA_SEARCH")
	responseRaw, err := s.search(s, query, dataPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to post datamart search request")
	}

	datasets, err := s.parse(responseRaw, baseDataset)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse datamart search response")
	}

	return datasets, nil
}
