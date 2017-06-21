package filter

import (
	"gopkg.in/olivere/elastic.v3"
)

// Empty represents an empty filter for a specific variable.
type Empty struct {
}

// Parse populates the filter from the query parameter arguments.
func (e *Empty) Parse(params []string) error {
	return nil
}

// Query returns the relevant elasticsearch query for the filter.
func (e *Empty) Query(field string) (elastic.Query, error) {
	return nil, nil
}
