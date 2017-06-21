package filter

import (
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v3"
)

// Category represents a categorical filter defined by a set of allowed terms
// categories for a categorical variable.
type Category struct {
	Categories []string
}

// Parse populates the filter from the query parameter arguments.
func (c *Category) Parse(params []string) error {
	if len(params) == 0 {
		return errors.New("missing categorical filter params, expected {type},{category_1},{category_2},...,{category_n}")
	}
	c.Categories = params
	return nil
}

// Query returns the relevant elasticsearch query for the filter.
func (c *Category) Query(field string) (elastic.Query, error) {
	return elastic.NewTermsQuery(field, c.Categories), nil
}
