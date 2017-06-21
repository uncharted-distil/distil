package filter

import (
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v3"
)

// Range represents a range filter defined by a min and max value.
type Range struct {
	Min float64
	Max float64
}

// Parse populates the filter from the query parameter arguments.
func (r *Range) Parse(params []string) error {
	if len(params) != 2 {
		return errors.New("missing range filter params, expected {type},{min},{max}")
	}
	min, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse range filter {min}")
	}
	max, err := strconv.ParseFloat(params[1], 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse range filter {max}")
	}
	r.Min = min
	r.Max = max
	return nil
}

// Query returns the relevant elasticsearch query for the filter.
func (r *Range) Query(field string) (elastic.Query, error) {
	return elastic.NewRangeQuery(field).Gte(r.Min).Lte(r.Max), nil
}
