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

package file

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/model"
)

const (
	// Provenance for file
	Provenance = "file"
)

// ImportDataset makes the dataset available for ingest and returns
// the URI to use for ingest.
func (s *Storage) ImportDataset(id string, uri string) (string, error) {
	// dataset is already on local file system and accessible for ingest
	return uri, nil
}

// IngestDataset adds a document consisting of the metadata to the file system.
func (s *Storage) IngestDataset(datasetSource metadata.DatasetSource, meta *model.Metadata) error {
	return errors.Errorf("Not implemented")
}

// CloneDataset is not supported (ES datasets are already ingested).
func (s *Storage) CloneDataset(dataset string, datasetNew string, storageNameNew string) error {
	return errors.Errorf("Not implemented")
}

// UpdateDataset updates a document consisting of the metadata to the file system.
func (s *Storage) UpdateDataset(dataset *api.Dataset) error {
	return errors.Errorf("Not implemented")
}

// DeleteDataset deletes a dataset from the file system.
func (s *Storage) DeleteDataset(dataset string) error {
	return errors.Errorf("Not implemented")
}

// FetchDatasets returns all datasets in the provided index.
func (s *Storage) FetchDatasets(includeIndex bool, includeMeta bool) ([]*api.Dataset, error) {
	// use default string in search to get complete list
	return s.SearchDatasets("", nil, includeIndex, includeMeta)
}

// FetchDataset returns a dataset in the provided index.
func (s *Storage) FetchDataset(datasetName string, includeIndex bool, includeMeta bool) (*api.Dataset, error) {
	return nil, errors.Errorf("Not implemented")
}

// SearchDatasets returns the datasets that match the search criteria in the
// provided index.
func (s *Storage) SearchDatasets(terms string, baseDataset *api.Dataset, includeIndex bool, includeMeta bool) ([]*api.Dataset, error) {
	rawSets, err := s.searchFolders(strings.Fields(terms))
	if err != nil {
		return nil, err
	}

	return s.parseDatasets(rawSets)
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
func (s *Storage) AddVariable(dataset string, varName string, varDisplayName string, varType string, varRole string) error {
	return errors.Errorf("Not supported")
}

// DeleteVariable is not supported by the datamart.
func (s *Storage) DeleteVariable(dataset string, varName string) error {
	return errors.Errorf("Not supported")
}

// AddGroupedVariable adds a variable grouping.
func (s *Storage) AddGroupedVariable(dataset string, varName string, varDisplayName string, varType string, varRole string, grouping model.BaseGrouping) error {
	return errors.Errorf("Not supported")
}

// RemoveGroupedVariable removes a variable grouping.
func (s *Storage) RemoveGroupedVariable(datasetName string, grouping model.BaseGrouping) error {
	return errors.Errorf("Not supported")
}

func (s *Storage) parseDatasets(raw []*model.Metadata) ([]*api.Dataset, error) {
	datasets := make([]*api.Dataset, 0)

	for _, meta := range raw {
		// merge all variables into a single set
		// TODO: figure out how we handle multiple data resources!
		vars := make([]*model.Variable, 0)
		for _, dr := range meta.DataResources {
			vars = append(vars, dr.Variables...)
		}
		datasets = append(datasets, &api.Dataset{
			Name:        meta.Name,
			Description: meta.Description,
			Folder:      meta.DatasetFolder,
			Summary:     meta.Summary,
			SummaryML:   meta.SummaryMachine,
			NumRows:     int64(meta.NumRows),
			NumBytes:    int64(meta.NumBytes),
			Variables:   vars,
			Provenance:  Provenance,
		})
	}

	return datasets, nil
}

func (s *Storage) searchFolders(terms []string) ([]*model.Metadata, error) {
	// cycle through each folder
	folders, err := ioutil.ReadDir(s.folder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read datamart directory")
	}

	matches := make([]*model.Metadata, 0)
	for _, info := range folders {
		if !info.IsDir() {
			if info.Name()[0] == '.' {
				// we ignore any files prefixed with `.`, ex. `.gitkeep`
				continue
			} else {
				return nil, errors.Errorf("'%s' is not a directory and is not prefixed with `.` but is in the datamart directory", info.Name())
			}
		}

		// load the metadata
		schemaFilename := path.Join(s.folder, info.Name(), compute.D3MDataSchema)
		meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFilename, true)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read metadata")
		}

		// check if match
		if datasetMatches(meta, terms) {
			matches = append(matches, meta)
		}
	}

	return matches, nil
}

func datasetMatches(meta *model.Metadata, terms []string) bool {
	// search the columns & description
	if matches(meta.Description, terms) || matches(meta.Summary, terms) ||
		matches(meta.Name, terms) {
		return true
	}

	for _, dr := range meta.DataResources {
		for _, f := range dr.Variables {
			if matches(f.Name, terms) {
				return true
			}
		}
	}

	return false
}

func matches(text string, terms []string) bool {
	//TODO: probably want to weigh matches in some way (more terms matched = better?)
	for _, t := range terms {
		if strings.Contains(text, t) {
			return true
		}
	}

	// if no terms provided, assume match
	return len(terms) == 0
}
