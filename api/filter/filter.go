package filter

import (
	"gopkg.in/olivere/elastic.v3"
)

// Filter represents a filter that can be applied to a dataset.
type Filter interface {
	Query(string) (elastic.Query, error)
	Parse([]string) error
}
