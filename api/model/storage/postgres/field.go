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
	"fmt"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// Field defines behaviour for a database field type.
type Field interface {
	FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool) (*api.VariableSummary, error)
	FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.VariableSummary, error)
	GetStorage() *Storage
	GetStorageName() string
	GetKey() string
	GetLabel() string
	GetType() string
}

// BasicField provides access to baseline field data
type BasicField struct {
	Storage     *Storage
	StorageName string
	Key         string
	Label       string
	Type        string
}

// GetStorage returns the storage associated with the field
func (b *BasicField) GetStorage() *Storage {
	return b.Storage
}

// GetStorageName returns the storage name of the table
func (b *BasicField) GetStorageName() string {
	return b.StorageName
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
// cluster column ID to ensure that it is used in downstream queries.
func (b *BasicField) updateClusterHighlight(filterParams *api.FilterParams) error {
	if !filterParams.Empty() && filterParams.Highlight != nil {
		if b.hasClusterData(filterParams.Highlight.Key) {
			filterParams.Highlight.Key = clusteringColName(filterParams.Highlight.Key)
		}
	}
	return nil
}
func (b *BasicField) hasClusterData(variableName string) bool {
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = '%s' AND column_name = '%s');",
		b.StorageName, variableName)
	res, err := b.Storage.client.Query(query)
	if err != nil {
		errors.Wrap(err, "failed to query cluster column status")
		return false
	}
	if res != nil {
		defer res.Close()
	}
	for res.Next() {
		var foundCol bool
		err = res.Scan(&foundCol)
		return err == nil && foundCol
	}
	return false
}

func clusteringColName(variableName string) string {
	return fmt.Sprintf("%s%s", model.ClusterVarPrefix, variableName)
}
