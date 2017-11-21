package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// ParseFilterParams parses filter parameters out of a request object
func ParseFilterParams(r *http.Request) (*model.FilterParams, error) {
	// parses a search parameter string formatteed as ?size=10&someIntField=integer,0,100&someCategoryFieldName=category,catA,catB,catF
	var filterParams model.FilterParams
	filterParams.Size = model.DefaultFilterSize

	for key, value := range r.URL.Query() {
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
			if size < searchSizeLimit {
				filterParams.Size = size
			} else {
				filterParams.Size = searchSizeLimit
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
				filterParams.Ranged = append(filterParams.Ranged,
					model.VariableRange{
						Min:  min,
						Max:  max,
						Name: key,
					})
			} else if filterType == CategoricalFilter {
				// categorical/ordinal should have type,category, category,...,category as args
				if len(varParams) < 2 {
					return nil, errors.Errorf("expected {type},{category_1},{category_2},...,{category_n} from [%s, %v]", key, value)
				}
				filterParams.Categorical = append(filterParams.Categorical,
					model.VariableCategories{
						Name:       key,
						Categories: varParams[1:],
					})
			} else {
				return nil, errors.Errorf("unhandled parameter type from [%s, %v]", key, value)
			}
		} else {
			// if we just receive a parameter key that is not 'size' we treat it as a variable flag with not
			// associated range / category feature.
			filterParams.None = append(filterParams.None, key)
		}
	}
	return &filterParams, nil
}
