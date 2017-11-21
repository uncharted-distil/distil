package model

const (
	// DefaultFilterSize represents the default filter search size
	DefaultFilterSize = 100
	// CategoricalFilter represents a categorical filter type
	CategoricalFilter = "categorical"
	// NumericalFilter represents a numerical filter type.
	NumericalFilter = "numerical"
)

// FilteredData provides the metadata and raw data values that match a supplied
// input filter.
type FilteredData struct {
	Name    string          `json:"name"`
	Columns []string        `json:"columns"`
	Types   []string        `json:"types"`
	Values  [][]interface{} `json:"values"`
}

// VariableRange defines the min/max value for a variable filter.
type VariableRange struct {
	Name string
	Min  float64
	Max  float64
}

// VariableCategories defines the set of allowed categories for a categorical
// variable filter.
type VariableCategories struct {
	Name       string
	Categories []string
}

// FilterParams defines the set of numeric range and categorical filters.  Variables
// with no range or category filters are also allowed.
type FilterParams struct {
	Size        int
	Ranged      []VariableRange
	Categorical []VariableCategories
	None        []string
}

// GetFieldList builds the filtered list of fields based on the filtering parameters.
func GetFieldList(filterParams *FilterParams, variables []*Variable, inclusive bool) []*Variable {
	variableLookup := make(map[string]*Variable)
	for _, v := range variables {
		variableLookup[v.Name] = v
	}

	fieldList := make([]*Variable, 0)
	if inclusive {
		// if inclusive, include all fields except specifically excluded fields
		excludedFields := make(map[string]bool)
		for _, f := range filterParams.None {
			excludedFields[f] = true
		}

		for _, v := range variables {
			if !excludedFields[v.Name] {
				fieldList = append(fieldList, v)
			}
		}
	} else {
		// if exclusive, exclude all fields except specifically included fields
		for _, f := range filterParams.Ranged {
			fieldList = append(fieldList, variableLookup[f.Name])
		}
		for _, f := range filterParams.Categorical {
			fieldList = append(fieldList, variableLookup[f.Name])
		}
		for _, f := range filterParams.None {
			fieldList = append(fieldList, variableLookup[f])
		}
	}

	return fieldList
}
