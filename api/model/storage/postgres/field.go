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

package postgres

import (
	api "github.com/uncharted-distil/distil/api/model"
)

// Field defines behaviour for a database field type.
type Field interface {
	FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool) (*api.VariableSummary, error)
	FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.VariableSummary, error)
	GetStorage() *Storage
	GetDatasetStorageName() string
	GetDatasetName() string
	GetKey() string
	GetLabel() string
	GetType() string
}

// BasicField provides access to baseline field data
type BasicField struct {
	Storage            *Storage
	DatasetStorageName string
	DatasetName        string
	Key                string
	Label              string
	Type               string
}

// GetStorage returns the storage associated with the field
func (b *BasicField) GetStorage() *Storage {
	return b.Storage
}

// GetDatasetStorageName returns the name used for the dataset table (PG imposes a number of limits on table names)
func (b *BasicField) GetDatasetStorageName() string {
	return b.DatasetStorageName
}

// GetDatasetName returns the name of the dataset
func (b *BasicField) GetDatasetName() string {
	return b.DatasetName
}

// GetKey returns the unique field key (name)
func (b *BasicField) GetKey() string {
	return b.Key
}

// GetLabel returns the field's label.  May not be unique, shouldn't be used to identify the field (use Key)
func (b *BasicField) GetLabel() string {
	return b.Label
}

// GetType returns the internal Distil type of the field.
func (b *BasicField) GetType() string {
	return b.Type
}

// Checks to see if the highlighted variable has cluster data.  If so, the highlight key will be switched to the
// cluster column ID to ensure that it is used in downstream queries.  This necessary when dealing with the timerseries
// compound facet, which will display cluster info when available.
func (b *BasicField) updateClusterHighlight(filterParams *api.FilterParams) error {
	if !filterParams.Empty() && filterParams.Highlight != nil {
		varExists, err := b.GetStorage().metadata.DoesVariableExist(b.GetDatasetName(), filterParams.Highlight.Key)
		if err != nil {
			return err
		}
		if !varExists {
			return nil
		}

		variable, err := b.GetStorage().metadata.FetchVariable(b.GetDatasetName(), filterParams.Highlight.Key)
		if err != nil {
			return err
		}

		if variable.Grouping != nil {
			if variable.Grouping.Properties.ClusterCol != "" &&
				api.HasClusterData(b.GetDatasetName(), variable.Grouping.Properties.ClusterCol, b.GetStorage().metadata) {
				filterParams.Highlight.Key = variable.Grouping.Properties.ClusterCol
				return nil
			}
			filterParams.Highlight.Key = variable.Grouping.IDCol
		}
	}
	return nil
}
