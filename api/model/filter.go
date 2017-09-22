package model

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

// FetchFilteredData creates a query to fetch a set of documents.  Applies filters to restrict the
// results to a user selected set of fields, with documents further filtered based on allowed ranges and
// categories.
func FetchFilteredData(storage Storage, dataset string, index string, filterParams *FilterParams) (*FilteredData, error) {
	return storage.FetchData(dataset, index, filterParams)
}
