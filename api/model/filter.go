package model

import (
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

// GetFilterVariables builds the filtered list of fields based on the filtering parameters.
func GetFilterVariables(filterParams *FilterParams, variables []*Variable) []*Variable {
	variableLookup := make(map[string]*Variable)
	for _, v := range variables {
		variableLookup[v.Name] = v
	}

	filtered := make([]*Variable, 0)
	for _, variable := range filterParams.Variables {
		filtered = append(filtered, variableLookup[variable])
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

			// name
			name, ok := json.String(filter, "name")
			if !ok {
				return nil, errors.Errorf("no `name` provided for filter")
			}

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
				categories, ok := json.StringArray(filter, "categories")
				if !ok {
					return nil, errors.Errorf("no `categories` provided for filter")
				}
				filterParams.Filters = append(filterParams.Filters, NewCategoricalFilter(name, mode, categories))
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
