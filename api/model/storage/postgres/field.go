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

package postgres

import (
	"fmt"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

const (
	baseTableAlias = "data"
)

// Field defines behaviour for a database field type.
type Field interface {
	FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error)
	FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error)
	GetStorage() *Storage
	GetDatasetStorageName() string
	GetDatasetName() string
	GetKey() string
	GetLabel() string
	GetType() string
	fetchExtremaStorage() (*api.Extrema, error)
	fetchExtremaByURI(resultURI string) (*api.Extrema, error)
}

// TimelineField defines the behaviour of a field which can be used as a timeline.
type TimelineField interface {
	Field
	fetchHistogram(filterParams *api.FilterParams, numBuckets int) (*api.Histogram, error)
	fetchHistogramWithJoins(filterParams *api.FilterParams, numBuckets int, joins []*joinDefinition, wheres []string, params []interface{}) (*api.Histogram, error)
}

// BasicField provides access to baseline field data
type BasicField struct {
	Storage            *Storage
	DatasetStorageName string
	DatasetName        string
	Key                string
	Label              string
	Type               string
	Count              string
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

func (b *BasicField) fetchExtremaStorage() (*api.Extrema, error) {
	return &api.Extrema{}, nil
}

func (b *BasicField) fetchExtremaByURI(resultURI string) (*api.Extrema, error) {
	return &api.Extrema{}, nil
}

func createJoinStatements(joins []*joinDefinition) string {
	joinSQL := ""
	for _, j := range joins {
		joinSQL = fmt.Sprintf("%s INNER JOIN %s AS %s ON %s.\"%s\" = %s.\"%s\"",
			joinSQL, j.joinTableName, j.joinAlias, j.baseAlias, j.baseColumn, j.joinAlias, j.joinColumn)
	}

	return joinSQL
}

// Checks to see if the highlighted variable has cluster data.  If so, the highlight key will be switched to the
// cluster column ID to ensure that it is used in downstream queries.  This is necessary when dealing with the timerseries
// compound facet, which will display cluster info when available.
func updateClusterFilters(metadataStorage api.MetadataStorage, dataset string, filterParams *api.FilterParams, mode api.SummaryMode) error {
	if filterParams != nil && !filterParams.IsEmpty(false) {
		for _, s := range filterParams.Filters {
			for _, f := range s.FeatureFilters {
				for _, h := range f.List {
					if err := updateClusterFilter(metadataStorage, dataset, filterParams.DataMode, h); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func updateClusterFilter(metadataStorage api.MetadataStorage, dataset string, dataMode api.DataMode, filter *model.Filter) error {
	varExists, err := metadataStorage.DoesVariableExist(dataset, filter.Key)
	if err != nil {
		return err
	}
	if !varExists {
		return nil
	}

	variable, err := metadataStorage.FetchVariable(dataset, filter.Key)
	if err != nil {
		return err
	}

	api.UpdateFilterKey(metadataStorage, dataset, dataMode, filter, variable)

	return nil
}

func (b BasicField) updateClusterHighlight(filterParams *api.FilterParams, mode api.SummaryMode) error {
	return updateClusterFilters(b.GetStorage().metadata, b.GetDatasetName(), filterParams, mode)
}

func getCountSQL(count string) string {
	if count == "" {
		count = "*"
	} else {
		count = fmt.Sprintf("DISTINCT \"%s\"", count)
	}

	return count
}
