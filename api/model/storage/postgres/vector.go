package postgres

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// VectorField defines behaviour for any Vector type.
type VectorField struct {
	Storage  *Storage
	Dataset  string
	Variable *model.Variable
}

// NewVectorField creates a new field of the vector type. A vector field
// uses unnest to flatten the database array and then uses the underlying
// data type to get summaries.
func NewVectorField(storage *Storage, dataset string, variable *model.Variable) *VectorField {
	field := &VectorField{
		Storage:  storage,
		Dataset:  dataset,
		Variable: variable,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *VectorField) FetchSummaryData(resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	var underlyingField Field
	if f.isNumerical() {
		underlyingField = NewNumericalFieldSubSelect(f.Storage, f.Dataset, f.Variable, f.subSelect)
	} else {
		underlyingField = NewCategoricalFieldSubSelect(f.Storage, f.Dataset, f.Variable, f.subSelect)
	}

	return underlyingField.FetchSummaryData(resultURI, filterParams, extrema)
}

// FetchNumericalStats gets the variable's numerical summary info (mean, stddev).
func (f *VectorField) FetchNumericalStats(filterParams *model.FilterParams) (*NumericalStats, error) {
	// confirm that the underlying type is numerical
	if !f.isNumerical() {
		return nil, errors.Errorf("field '%s' is not a numerical vector", f.Variable.Key)
	}

	// use the underlying numerical field implementation
	field := NewNumericalFieldSubSelect(f.Storage, f.Dataset, f.Variable, f.subSelect)

	return field.FetchNumericalStats(filterParams)
}

// FetchNumericalStatsByResult gets the variable's numerical summary info (mean, stddev) for a result set.
func (f *VectorField) FetchNumericalStatsByResult(resultURI string, filterParams *model.FilterParams) (*NumericalStats, error) {
	// confirm that the underlying type is numerical
	if !f.isNumerical() {
		return nil, errors.Errorf("field '%s' is not a numerical vector", f.Variable.Key)
	}

	// use the underlying numerical field implementation
	field := NewNumericalFieldSubSelect(f.Storage, f.Dataset, f.Variable, f.subSelect)

	return field.FetchNumericalStatsByResult(resultURI, filterParams)
}

// FetchPredictedSummaryData pulls predicted data from the result table and builds
// the categorical histogram for the field.
func (f *VectorField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	return nil, errors.Errorf("vector field cannot be a target so no result will be pulled")
}

func (f *VectorField) isNumerical() bool {
	return model.IsNumerical(strings.Replace(f.Variable.Type, "Vector", "", -1))
}

func (f *VectorField) subSelect() string {
	return fmt.Sprintf("(SELECT \"%s\", unnest(\"%s\") as %s FROM %s)",
		model.D3MIndexFieldName, f.Variable.Key, f.Variable.Key, f.Dataset)
}
