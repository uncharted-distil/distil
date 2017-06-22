package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	elastic_api "github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util/json"
	"goji.io/pat"
)

const (
	defaultSearchSize = 100
	searchSizeLimit   = 1000
)

func parseFilterParams(r *http.Request) (*model.FilterParams, error) {
	// parses a search parameter string formatteed as ?size=10&someIntField=integer,0,100&someCategoryFieldName=category,catA,catB,catF
	var filterParams model.FilterParams
	filterParams.Size = defaultSearchSize

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
			varType := varParams[0]
			switch varType {
			case "integer", "float":
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
					model.VariableRange{Min: min, Max: max, Variable: model.Variable{Name: key, Type: varType}})
			case "ordinal", "categorical":
				// categorical/ordinal should have type,category, category,...,category as args
				if len(varParams) < 2 {
					return nil, errors.Errorf("expected {type},{category_1},{category_2},...,{category_n} from [%s, %v]", key, value)
				}
				filterParams.Categorical = append(filterParams.Categorical,
					model.VariableCategories{Variable: model.Variable{Name: key, Type: varType}, Categories: varParams[1:]})
			default:
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

// FilteredDataHandler creates a route that fetches filtered data from an elastic search instance.
func FilteredDataHandler(ctor elastic_api.ClientCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")

		// get variable names and ranges out of the params
		filterParams, err := parseFilterParams(r)
		if err != nil {
			handleError(w, err)
			return
		}

		// get elasticsearch client
		client, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch filtered data based on the supplied search parameters
		data, err := model.FetchFilteredData(client, dataset, filterParams)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}

		// marshall output into JSON
		bytes, err := json.Marshal(data)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}
