package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	elastic_api "github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util/json"
	log "github.com/unchartedsoftware/plog"
	"goji.io/pat"
)

const (
	defaultSearchSize  = 100
	searchSizeLimit    = 1000
)

func handleParamParseError(key string, value []string, err error) {
	log.Errorf("failure parsing search params [%s,%v] - %+v", key, value, err)
}

func parseFilterParams(r *http.Request) *model.FilterParams {
	var filterParams model.FilterParams
	filterParams.Size = defaultSearchSize

	for key, value := range r.URL.Query() {
		// parse out the requested search size using the default in error cases and the
		// min of requested size and limit otherwise
		if key == "size" {
			if len(value) != 1 {
				handleParamParseError(key, value, errors.New("expected single value for size"))
			}
			size, err := strconv.Atoi(value[0])
			if err != nil {
				handleParamParseError(key, value, errors.Wrap(err, "failed to parse size"))
				continue
			}
			if size < searchSizeLimit {
				filterParams.Size = size
			} else {
				filterParams.Size = searchSizeLimit
			}

		} else if value != nil && len(value) > 0 && value[0] != "" {
			// split the value on a comma
			varParams := strings.Split(value[0], ",")
			varType := varParams[0]
			switch varType {
			case "integer", "float":
				if len(varParams) != 3 {
					handleParamParseError(key, value, errors.New("expected type,min,max"))
					continue
				}
				min, err := strconv.ParseFloat(varParams[1], 64)
				if err != nil {
					handleParamParseError(key, value, errors.Wrap(err, "failed to parse min"))
					continue
				}
				max, err := strconv.ParseFloat(varParams[2], 64)
				if err != nil {
					handleParamParseError(key, value, errors.Wrap(err, "failed to parse max"))
					continue
				}
				filterParams.Ranged = append(filterParams.Ranged,
					model.VariableRange{Min: min, Max: max, Variable: model.Variable{Name: key, Type: varType}})
			case "ordinal", "categorical":
				if len(value) >= 2 {
					handleParamParseError(key, value, errors.New("expected type,category_1,category_2,...,category_n"))
					continue
				}
				filterParams.Categorical = append(filterParams.Categorical,
					model.VariableCategories{Variable: model.Variable{Name: key, Type: varType}, Categories: varParams[1:]})
			default:
				continue
			}
		} else {
			filterParams.None = append(filterParams.None, key)
		}
	}
	return &filterParams
}

// FilteredDataHandler creates a route that fetches filtered data from an elastic search instance.
func FilteredDataHandler(ctor elastic_api.ClientCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")

		// get variable names and ranges out of the params
		filterParams := parseFilterParams(r)

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
