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

package model

import (
	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/util/json"
)

// DatasetType is used to identify the type of dataset ingested.
type DatasetType string

const (
	// DatasetTypeModelling is a dataset used to build models.
	DatasetTypeModelling DatasetType = "modelling"
	// DatasetTypeInference is a dataset consumed by a model to infer predictions.
	DatasetTypeInference DatasetType = "inference"
)

// RawDataset contains basic information about the structure of the dataset as well
// as the raw learning data.
type RawDataset struct {
	ID              string
	Name            string
	Metadata        *model.Metadata
	Data            [][]string
	DefinitiveTypes bool
}

// Dataset represents a decsription of a dataset.
type Dataset struct {
	ID              string                 `json:"datasetID"`
	Name            string                 `json:"datasetName"`
	StorageName     string                 `json:"storageName"`
	Folder          string                 `json:"datasetFolder"`
	Description     string                 `json:"description"`
	Summary         string                 `json:"summary"`
	SummaryML       string                 `json:"summaryMachine"`
	Variables       []*model.Variable      `json:"variables"`
	NumRows         int64                  `json:"numRows"`
	NumBytes        int64                  `json:"numBytes"`
	Provenance      string                 `json:"provenance"`
	Source          metadata.DatasetSource `json:"source"`
	JoinSuggestions []*JoinSuggestion      `json:"datasetOrigins"`
	JoinScore       float64                `json:"joinScore"`
	Type            DatasetType            `json:"type"`
	LearningDataset string                 `json:"learningDataset"`
}

// QueriedDataset wraps dataset querying components into a single entity.
type QueriedDataset struct {
	Metadata *Dataset
	Data     *FilteredData
	Filters  *FilterParams
	IsTrain  bool
}

// JoinSuggestion specifies potential joins between datasets.
type JoinSuggestion struct {
	BaseDataset   string               `json:"baseDataset"`
	BaseColumns   []string             `json:"baseColumns"`
	JoinDataset   string               `json:"joinDataset"`
	JoinColumns   []string             `json:"joinColumns"`
	JoinScore     float64              `json:"joinScore"`
	DatasetOrigin *model.DatasetOrigin `json:"datasetOrigin"`
	Index         int                  `json:"index"`
}

// FetchDataset builds a QueriedDataset from the needed parameters.
func FetchDataset(dataset string, includeIndex bool, includeMeta bool, filterParams *FilterParams, storageMeta MetadataStorage, storageData DataStorage) (*QueriedDataset, error) {
	metadata, err := storageMeta.FetchDataset(dataset, false, true)
	if err != nil {
		return nil, err
	}

	data, err := storageData.FetchData(dataset, metadata.StorageName, filterParams, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch data")
	}

	return &QueriedDataset{
		Metadata: metadata,
		Data:     data,
		Filters:  filterParams,
	}, nil
}

// GetD3MIndexVariable returns the D3M index variable.
func (d *Dataset) GetD3MIndexVariable() *model.Variable {
	for _, v := range d.Variables {
		if v.Name == model.D3MIndexName {
			return v
		}
	}

	return nil
}

// UpdateExtremas updates the variable extremas based on the data stored.
func UpdateExtremas(dataset string, varName string, storageMeta MetadataStorage, storageData DataStorage) error {
	// get the metadata and then query the data storage for the latest values
	d, err := storageMeta.FetchDataset(dataset, false, false)
	if err != nil {
		return err
	}

	// find the variable
	var v *model.Variable
	for _, variable := range d.Variables {
		if variable.Name == varName {
			v = variable
			break
		}
	}

	// only care about datetime and numerical
	if model.IsDateTime(v.Type) || model.IsNumerical(v.Type) {
		// get the extrema
		extrema, err := storageData.FetchExtrema(d.StorageName, v)
		if err != nil {
			return err
		}

		// store the extrema to ES
		err = storageMeta.SetExtrema(dataset, varName, extrema)
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseDatasetOriginsFromJSON parses dataset origins from string maps.
func ParseDatasetOriginsFromJSON(originsJSON []map[string]interface{}) []*model.DatasetOrigin {

	origins := make([]*model.DatasetOrigin, len(originsJSON))

	for i, originJSON := range originsJSON {
		origins[i] = parseDatasetOriginFromJSON(originJSON)
	}

	return origins
}

func parseDatasetOriginFromJSON(originJSON map[string]interface{}) *model.DatasetOrigin {
	searchResult, ok := json.String(originJSON, "searchResult")
	if !ok {
		searchResult = ""
	}
	provenance, ok := json.String(originJSON, "provenance")
	if !ok {
		provenance = ""
	}
	sourceDataset, ok := json.String(originJSON, "sourceDataset")
	if !ok {
		sourceDataset = ""
	}

	return &model.DatasetOrigin{
		SearchResult:  searchResult,
		Provenance:    provenance,
		SourceDataset: sourceDataset,
	}
}
