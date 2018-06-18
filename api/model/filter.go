package model

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"

	"github.com/unchartedsoftware/distil/api/util/json"
)

const (
	// DefaultFilterSize represents the default filter search size.
	DefaultFilterSize = 100
	// FilterSizeLimit represents the largest filter size.
	FilterSizeLimit = 1000
	// CategoricalFilter represents a categorical filter type.
	CategoricalFilter = "categorical"
	// NumericalFilter represents a numerical filter type.
	NumericalFilter = "numerical"
	// FeatureFilter represents a categorical filter type.
	FeatureFilter = "feature"
	// RowFilter represents a numerical filter type.
	RowFilter = "row"
	// IncludeFilter represents an inclusive filter mode.
	IncludeFilter = "include"
	// ExcludeFilter represents an exclusive filter mode.
	ExcludeFilter = "exclude"
)

// FilterParams defines the set of numeric range and categorical filters. Variables
// with no range or category filters are also allowed.
type FilterParams struct {
	Size      int       `json:"size"`
	Filters   []*Filter `json:"filters"`
	Variables []string  `json:"variables"`
}

func stringSliceEqual(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Merge merges another set of filter params into this set, expanding all
// properties.
func (f *FilterParams) Merge(other *FilterParams) {
	// take greater of sizes
	if other.Size > f.Size {
		f.Size = other.Size
	}
	for _, filter := range other.Filters {
		found := false
		for _, currentFilter := range f.Filters {
			if filter.Name == currentFilter.Name &&
				filter.Min == currentFilter.Min &&
				filter.Max == currentFilter.Max &&
				stringSliceEqual(filter.Categories, currentFilter.Categories) {
				found = true
				break
			}
		}
		if !found {
			f.Filters = append(f.Filters, filter)
		}
	}
	for _, variable := range other.Variables {
		found := false
		for _, currentVariable := range f.Variables {
			if variable == currentVariable {
				found = true
				break
			}
		}
		if !found {
			f.Variables = append(f.Variables, variable)
		}
	}
}

// FilteredData provides the metadata and raw data values that match a supplied
// input filter.
type FilteredData struct {
	Name    string          `json:"name"`
	NumRows int             `json:"numRows"`
	Columns []string        `json:"columns"`
	Types   []string        `json:"types"`
	Values  [][]interface{} `json:"values"`
}

// Filter defines a variable filter.
type Filter struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Mode       string   `json:"mode"`
	Min        *float64 `json:"min"`
	Max        *float64 `json:"max"`
	Categories []string `json:"categories"`
	D3mIndices []string `json:"d3mIndices"`
}

// NewNumericalFilter instantiates a numerical filter.
func NewNumericalFilter(name string, mode string, min float64, max float64) *Filter {
	return &Filter{
		Name: name,
		Type: NumericalFilter,
		Mode: mode,
		Min:  &min,
		Max:  &max,
	}
}

// NewCategoricalFilter instantiates a categorical filter.
func NewCategoricalFilter(name string, mode string, categories []string) *Filter {
	sort.Strings(categories)
	return &Filter{
		Name:       name,
		Type:       CategoricalFilter,
		Mode:       mode,
		Categories: categories,
	}
}

// NewRowFilter instantiates a row filter.
func NewRowFilter(mode string, d3mIndices []string) *Filter {
	return &Filter{
		Type:       RowFilter,
		Mode:       mode,
		D3mIndices: d3mIndices,
	}
}

// GetFilterVariables builds the filtered list of fields based on the filtering parameters.
func GetFilterVariables(filterVariables []string, variables []*Variable) []*Variable {
	variableLookup := make(map[string]*Variable)
	for _, v := range variables {
		variableLookup[v.Name] = v
	}

	filtered := make([]*Variable, 0)
	for _, variable := range filterVariables {
		filtered = append(filtered, variableLookup[variable])
		// check for metadata var type
		if HasMetadataVar(variableLookup[variable].Type) {
			metadataVarName := fmt.Sprintf("%s%s", MetadataVarPrefix, variable)
			metadataVar, ok := variableLookup[metadataVarName]
			if ok {
				filtered = append(filtered, metadataVar)
			}
		}
	}

	return filtered
}

// ParseFilterParamsFromJSON parses filter parameters out of a map[string]interface{}
func ParseFilterParamsFromJSON(params map[string]interface{}) (*FilterParams, error) {
	filterParams := &FilterParams{
		Size: json.IntDefault(params, DefaultFilterSize, "size"),
	}

	filters, ok := json.Array(params, "filters")
	if ok {
		for _, filter := range filters {

			// type
			typ, ok := json.String(filter, "type")
			if !ok {
				return nil, errors.Errorf("no `type` provided for filter")
			}

			// mode
			mode, ok := json.String(filter, "mode")
			if !ok {
				return nil, errors.Errorf("no `mode` provided for filter")
			}

			// numeric
			if typ == NumericalFilter {
				name, ok := json.String(filter, "name")
				if !ok {
					return nil, errors.Errorf("no `name` provided for filter")
				}
				min, ok := json.Float(filter, "min")
				if !ok {
					return nil, errors.Errorf("no `min` provided for filter")
				}
				max, ok := json.Float(filter, "max")
				if !ok {
					return nil, errors.Errorf("no `max` provided for filter")
				}
				filterParams.Filters = append(filterParams.Filters, NewNumericalFilter(name, mode, min, max))
			}

			// categorical
			if typ == CategoricalFilter {
				name, ok := json.String(filter, "name")
				if !ok {
					return nil, errors.Errorf("no `name` provided for filter")
				}
				categories, ok := json.StringArray(filter, "categories")
				if !ok {
					return nil, errors.Errorf("no `categories` provided for filter")
				}
				filterParams.Filters = append(filterParams.Filters, NewCategoricalFilter(name, mode, categories))
			}

			// row
			if typ == RowFilter {
				indices, ok := json.StringArray(filter, "d3mIndices")
				if !ok {
					return nil, errors.Errorf("no `d3mIndices` provided for filter")
				}
				filterParams.Filters = append(filterParams.Filters, NewRowFilter(mode, indices))
			}
		}
	}

	variables, ok := json.StringArray(params, "variables")
	if ok {
		filterParams.Variables = variables
	}

	sort.SliceStable(filterParams.Filters, func(i, j int) bool {
		return filterParams.Filters[i].Name < filterParams.Filters[j].Name
	})

	return filterParams, nil
}
