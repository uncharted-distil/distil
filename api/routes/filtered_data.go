package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/util/json"
	log "github.com/unchartedsoftware/plog"
	"goji.io/pat"
	elastic "gopkg.in/olivere/elastic.v3"
)

const (
	variableParseError = "error parsing %s as %v"
	defaultSearchSize  = 100
	searchSizeLimit    = 1000
)

type data struct {
	Name     string          `json:"name"`
	Metadata []Variable      `json:"metadata"`
	Values   [][]interface{} `json:"values"`
}

type variableRange struct {
	Variable
	min float64
	max float64
}

type variableCategorical struct {
	Variable
	categories []string
}

type searchParams struct {
	size        int
	ranged      []variableRange
	categorical []variableCategorical
	none        []string
}

func handleParamParseError(key string, value []string, err error) {
	log.Errorf("failure parsing search params [%s,%v] - %+v", key, value, err)
}

func parseSearchParams(r *http.Request) *searchParams {
	var searchParams searchParams
	searchParams.size = defaultSearchSize

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
				searchParams.size = size
			} else {
				searchParams.size = searchSizeLimit
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
				searchParams.ranged = append(searchParams.ranged, variableRange{Variable{key, varType}, min, max})
			case "ordinal", "categorical":
				if len(value) >= 2 {
					handleParamParseError(key, value, errors.New("expected type,category_1,category_2,...,category_n"))
					continue
				}
				searchParams.categorical = append(searchParams.categorical, variableCategorical{Variable{key, varType}, varParams[1:]})
			default:
				continue
			}
		} else {
			searchParams.none = append(searchParams.none, key)
		}
	}
	return &searchParams
}

func parseVariable(varType string, data interface{}) (interface{}, error) {
	var val interface{}
	var ok bool

	switch varType {
	case "float":
		val, ok = json.Float(data.(map[string]interface{}), "value")
		if !ok {
			return nil, errors.Errorf(variableParseError, data, varType)
		}
	case "integer", "ordinal":
		val, ok = json.Int(data.(map[string]interface{}), "value")
		if !ok {
			return nil, errors.Errorf(variableParseError, data, varType)
		}
	case "categorical":
		val, ok = json.String(data.(map[string]interface{}), "value")
		if !ok {
			return nil, errors.Errorf(variableParseError, data, varType)
		}
	default:
		return nil, errors.Errorf("unhandled var type %s for %v", data, varType)
	}
	return val, nil
}

func parseData(searchResults *elastic.SearchResult) (*data, error) {
	var data data

	for idx, hit := range searchResults.Hits.Hits {
		// parse hit into JSON
		src, err := json.Unmarshal(*hit.Source)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse data")
		}

		// On the first time through, parse out name/type info and store that in a header.  We also
		// store the name/type tuples in a map for quick lookup
		if idx == 0 {
			data.Name = hit.Index
			for key, value := range src {
				varType, ok := json.String(value.(map[string]interface{}), "schemaType")
				if !ok {
					return nil, errors.Errorf("failed to extract type info for %s during metadata creation", key)
				}
				variable := Variable{key, varType}
				data.Metadata = append(data.Metadata, variable)
			}
		}

		// Create a temporary metadata -> index map.  Required because the variable data for each hit returned
		//  from ES is unordered.
		var metadataIndex = make(map[string]int, len(data.Metadata))
		for idx, value := range data.Metadata {
			metadataIndex[value.Name] = idx
		}

		// extract data for all variables
		values := make([]interface{}, len(data.Metadata))
		for key, value := range src {
			index := metadataIndex[key]
			varType := data.Metadata[index].Type
			result, err := parseVariable(varType, value)
			if err != nil {
				log.Errorf("%+v", err)
			}
			values[index] = result
		}
		// add the row to the variable data
		data.Values = append(data.Values, values)
	}
	return &data, nil
}

// FilteredDataHandler creates a route that fetches filtered data from an elastic search instance.
func FilteredDataHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")

		log.Infof("Processing data request from dataset %s", dataset)

		// get variable names and ranges out of the params
		searchParams := parseSearchParams(r)

		// construct an ES query that fetches documents from the dataset with the supplied variable filters applied
		query := elastic.NewBoolQuery()
		var keys []string
		for _, variable := range searchParams.ranged {
			query = query.Filter(elastic.NewRangeQuery(variable.Name + ".value").Gte(variable.min).Lte(variable.max))
			keys = append(keys, variable.Name)
		}
		for _, variable := range searchParams.categorical {
			query = query.Filter(elastic.NewTermsQuery(variable.Name+".value", variable.categories))
			keys = append(keys, variable.Name)
		}
		for _, variableName := range searchParams.none {
			keys = append(keys, variableName)
		}

		fetchContext := elastic.NewFetchSourceContext(true).Include(keys...)

		// execute the ES query
		res, err := client.Search().
			Query(query).
			Index(dataset).
			Size(searchParams.size).
			FetchSource(true).
			FetchSourceContext(fetchContext).
			Do()
		if err != nil {
			log.Errorf("elasticsearch filtered data query failed - %+v", err)
		}

		// parse the result
		data, err := parseData(res)

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
