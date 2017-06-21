package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/filter"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util/json"
)

const (
	defaultFilterSize = 100
	filterSizeLimit   = 1000
)

func parseFilterSize(values []string) (int, error) {
	// parse out the requested search size using the default in error cases
	// and the min of requested size and limit otherwise
	if len(values) != 1 {
		return 0, errors.New("failure parsing filter size, expected size={size}")
	}
	size, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, errors.Errorf("failed to parse size %v, value must be an integer", values[0])
	}
	if size > filterSizeLimit {
		return filterSizeLimit, nil
	}
	return size, nil
}

func isEmptyFilter(values []string) bool {
	// no values, empty filter
	return values == nil || len(values) == 0 || values[0] == ""
}

func parseVariableFilter(varName string, values []string) (filter.Filter, error) {
	// split query param values
	split := strings.Split(values[0], ",")

	// ensure we have at least the variable type and one parameter
	if len(split) < 2 {
		return nil, errors.Errorf("failure parsing filter params %s=%v, expected {variable}={type},{arguments...}", varName, values)
	}

	// get variable type and the filter params
	varType := split[0]
	params := split[1:]

	switch varType {
	case "integer", "float":
		// range filter
		filter := &filter.Range{}
		err := filter.Parse(params)
		if err != nil {
			return nil, err
		}
		return filter, nil

	case "ordinal", "categorical":
		// category filter
		filter := &filter.Category{}
		err := filter.Parse(params)
		if err != nil {
			return nil, err
		}
		return filter, nil

	default:
		return nil, errors.Errorf("failure parsing filter params %s=%v, unsupported {type} of %v, expected {variable}={type},{arguments...}", varName, values, varType)
	}
}

func parseFilterSet(r *http.Request) (*filter.Set, error) {
	set := filter.NewSet(defaultFilterSize)
	for key, values := range r.URL.Query() {
		if key == "size" {
			// filter size
			size, err := parseFilterSize(values)
			if err != nil {
				return nil, err
			}
			set.Size = size

		} else if isEmptyFilter(values) {
			// add empty filter
			set.Filters[key] = &filter.Empty{}

		} else {
			// add variable filter
			filter, err := parseVariableFilter(key, values)
			if err != nil {
				return nil, err
			}
			set.Filters[key] = filter
		}
	}
	return set, nil
}

// FilteredDataHandler creates a route that fetches filtered data from an elastic search instance.
func FilteredDataHandler(ctor elastic.ClientCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// get dataset name
		dataset := pat.Param(r, "dataset")

		// get variable names and ranges out of the params
		set, err := parseFilterSet(r)
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
		data, err := model.FetchFilteredData(client, dataset, set)
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
