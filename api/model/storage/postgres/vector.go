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
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

const (
	unnestedSuffix = "_unnested"
)

// VectorField defines behaviour for any Vector type.
type VectorField struct {
	Storage     *Storage
	StorageName string
	Key         string
	Label       string
	Type        string
	Unnested    string
}

// NewVectorField creates a new field of the vector type. A vector field
// uses unnest to flatten the database array and then uses the underlying
// data type to get summaries.
func NewVectorField(storage *Storage, storageName string, key string, label string, typ string) *VectorField {
	field := &VectorField{
		Storage:     storage,
		StorageName: storageName,
		Key:         key + unnestedSuffix,
		Label:       label,
		Type:        typ,
		Unnested:    key,
	}
	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *VectorField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	var underlyingField Field
	if f.isNumerical() {
		underlyingField = NewNumericalFieldSubSelect(f.Storage, f.StorageName, f.Key, f.Label, f.Type, f.subSelect)
	} else {
		underlyingField = NewCategoricalFieldSubSelect(f.Storage, f.StorageName, f.Key, f.Label, f.Type, f.subSelect)
	}

	histo, err := underlyingField.FetchSummaryData(resultURI, filterParams, extrema)
	if err != nil {
		return nil, err
	}
	histo.Key = f.Unnested
	return histo, nil
}

// FetchNumericalStats gets the variable's numerical summary info (mean, stddev).
func (f *VectorField) FetchNumericalStats(filterParams *api.FilterParams) (*NumericalStats, error) {
	// confirm that the underlying type is numerical
	if !f.isNumerical() {
		return nil, errors.Errorf("field '%s' is not a numerical vector", f.Key)
	}

	// use the underlying numerical field implementation
	field := NewNumericalFieldSubSelect(f.Storage, f.StorageName, f.Key, f.Label, f.Type, f.subSelect)

	return field.FetchNumericalStats(filterParams)
}

// FetchNumericalStatsByResult gets the variable's numerical summary info (mean, stddev) for a result set.
func (f *VectorField) FetchNumericalStatsByResult(resultURI string, filterParams *api.FilterParams) (*NumericalStats, error) {
	// confirm that the underlying type is numerical
	if !f.isNumerical() {
		return nil, errors.Errorf("field '%s' is not a numerical vector", f.Key)
	}

	// use the underlying numerical field implementation
	field := NewNumericalFieldSubSelect(f.Storage, f.StorageName, f.Key, f.Label, f.Type, f.subSelect)

	return field.FetchNumericalStatsByResult(resultURI, filterParams)
}

// FetchPredictedSummaryData pulls predicted data from the result table and builds
// the categorical histogram for the field.
func (f *VectorField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	return nil, errors.Errorf("vector field cannot be a target so no result will be pulled")
}

func (f *VectorField) isNumerical() bool {
	return model.IsNumerical(strings.Replace(f.Type, "Vector", "", -1))
}

func (f *VectorField) subSelect() string {
	return fmt.Sprintf("(SELECT \"%s\", unnest(\"%s\") as %s FROM %s)",
		model.D3MIndexFieldName, f.Unnested, f.Key, f.StorageName)
}
