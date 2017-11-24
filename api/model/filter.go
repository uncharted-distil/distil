package model

import (
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
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
	// EmptyFilter represents an empty filter type.
	EmptyFilter = "empty"
)

// FilteredData provides the metadata and raw data values that match a supplied
// input filter.
type FilteredData struct {
	Name    string          `json:"name"`
	Columns []string        `json:"columns"`
	Types   []string        `json:"types"`
	Values  [][]interface{} `json:"values"`
}

// VariableFilter defines a variable filter.
type VariableFilter struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Min        *float64 `json:"min"`
	Max        *float64 `json:"max"`
	Categories []string `json:"categories"`
}

// NewNumericalFilter instantiates a numerical filter.
func NewNumericalFilter(name string, min float64, max float64) *VariableFilter {
	return &VariableFilter{
		Name: name,
		Type: NumericalFilter,
		Min:  &min,
		Max:  &max,
	}
}

// NewCategoricalFilter instantiates a categorical filter.
func NewCategoricalFilter(name string, categories []string) *VariableFilter {
	sort.Strings(categories)
	return &VariableFilter{
		Name:       name,
		Type:       CategoricalFilter,
		Categories: categories,
	}
}

// NewEmptyFilter instantiates an empty filter.
func NewEmptyFilter(name string) *VariableFilter {
	return &VariableFilter{
		Name: name,
		Type: EmptyFilter,
	}
}

// FilterParams defines the set of numeric range and categorical filters.  Variables
// with no range or category filters are also allowed.
type FilterParams struct {
	Size    int
	Filters []*VariableFilter
}

// GetFilterVariables builds the filtered list of fields based on the filtering parameters.
func GetFilterVariables(filterParams *FilterParams, variables []*Variable, inclusive bool) []*Variable {
	variableLookup := make(map[string]*Variable)
	for _, v := range variables {
		variableLookup[v.Name] = v
	}

	filtered := make([]*Variable, 0)
	if inclusive {
		// if inclusive, include all fields except specifically excluded fields
		excludedFields := make(map[string]bool)
		for _, f := range filterParams.Filters {
			if f.Type == EmptyFilter {
				excludedFields[f.Name] = true
			}
		}
		for _, v := range variables {
			if !excludedFields[v.Name] {
				filtered = append(filtered, v)
			}
		}
	} else {
		// if exclusive, exclude all fields except specifically included fields
		for _, f := range filterParams.Filters {
			filtered = append(filtered, variableLookup[f.Name])
		}
	}

	return filtered
}

// ParseFilterParamsURL parses filter parameters out of a url.Values object.
func ParseFilterParamsURL(values url.Values) (*FilterParams, error) {
	// parses a search parameter string formatteed as:
	//
	// ?size=10&someIntField=integer,0,100&someCategoryFieldName=category,catA,catB,catF
	//
	filterParams := &FilterParams{
		Size: DefaultFilterSize,
	}

	for key, value := range values {
		// parse out the requested search size using the default in error cases and the
		// min of requested size and limit otherwise
		if key == "size" {
			if len(value) != 1 {
				return nil, errors.Errorf("expected single integer value for parameter [%s, %v]", key, value)
			}
			size, err := strconv.Atoi(value[0])
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse int from [%s, %v]", key, value)
			}
			if size < FilterSizeLimit {
				filterParams.Size = size
			} else {
				filterParams.Size = FilterSizeLimit
			}
		} else if value != nil && len(value) > 0 && value[0] != "" {
			// the are assumed to be variable range/cateogry parameters.

			// tokenize using a comma
			varParams := strings.Split(value[0], ",")
			filterType := varParams[0]
			if filterType == NumericalFilter {
				// floats and ints should have type, min, max as args
				if len(varParams) != 3 {
					return nil, errors.Errorf("expected {type},{min},{max} from [s%s, %v]", key, value)
				}
				min, err := strconv.ParseFloat(varParams[1], 64)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to parse range min from [%s, %v]", key, value)
				}
				max, err := strconv.ParseFloat(varParams[2], 64)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to parse range max from [%s, %v]", key, value)
				}
				filterParams.Filters = append(filterParams.Filters, NewNumericalFilter(key, min, max))
			} else if filterType == CategoricalFilter {
				// categorical/ordinal should have type,category, category,...,category as args
				if len(varParams) < 2 {
					return nil, errors.Errorf("expected {type},{category_1},{category_2},...,{category_n} from [%s, %v]", key, value)
				}
				filterParams.Filters = append(filterParams.Filters, NewCategoricalFilter(key, varParams[1:]))
			} else {
				return nil, errors.Errorf("unhandled parameter type from [%s, %v]", key, value)
			}
		} else {
			// if we just receive a parameter key that is not 'size' we treat it as a variable flag with not
			// associated range / category feature.
			filterParams.Filters = append(filterParams.Filters, NewEmptyFilter(key))
		}
	}

	sort.SliceStable(filterParams.Filters, func(i, j int) bool {
		return filterParams.Filters[i].Name < filterParams.Filters[j].Name
	})

	return filterParams, nil
}

// ParseFilterParamsJSON parses filter parameters out of a json.RawMessage object.
func ParseFilterParamsJSON(raw json.RawMessage) (*FilterParams, error) {
	// filter params for subsequent store query
	filterParams := &FilterParams{
		Size: DefaultFilterSize,
	}

	// unmarshall from params porition of message
	err := json.Unmarshal(raw, &filterParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse filter params")
	}

	for _, filter := range filterParams.Filters {
		// parse out filter parameters
		if filter.Type == "" {
			return nil, errors.Errorf("missing filter type")
		}

		switch filter.Type {
		case NumericalFilter:
			// numeric
			if filter.Min == nil ||
				filter.Max == nil {
				return nil, errors.New("numerical filter missing min/max value")
			}

		case CategoricalFilter:
			// categorical
			if filter.Categories == nil {
				return nil, errors.New("categorical filter missing categories set")
			}
			sort.Strings(filter.Categories)

		}
	}

	sort.SliceStable(filterParams.Filters, func(i, j int) bool {
		return filterParams.Filters[i].Name < filterParams.Filters[j].Name
	})

	return filterParams, nil
}
