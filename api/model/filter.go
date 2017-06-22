package model

import (
	"sort"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/util/json"
	log "github.com/unchartedsoftware/plog"
	elastic "gopkg.in/olivere/elastic.v3"
)

// FilteredData provides the metadata and raw data values that match a supplied
// input filter.
type FilteredData struct {
	Name     string          `json:"name"`
	Metadata []*Variable     `json:"metadata"`
	Values   [][]interface{} `json:"values"`
}

// VariableRange defines the min/max value for a variable filter.
type VariableRange struct {
	Variable
	Min float64
	Max float64
}

// VariableCategories defines the set of allowed categories for a categorical
// variable filter.
type VariableCategories struct {
	Variable
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

func parseResults(searchResults *elastic.SearchResult) (*FilteredData, error) {
	var data FilteredData

	for idx, hit := range searchResults.Hits.Hits {
		// parse hit into JSON
		src, err := json.Unmarshal(*hit.Source)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse data")
		}

		variables, ok := json.Map(src)
		if !ok {
			return nil, errors.Wrap(err, "failed to parse data")
		}

		// On the first time through, parse out name/type info and store that in a header.  We also
		// store the name/type tuples in a map for quick lookup
		if idx == 0 {
			data.Name = hit.Index
			for key, variable := range variables {
				varType, ok := json.String(variable, VariableTypeField)
				if !ok {
					return nil, errors.Errorf("failed to extract type info for %s during metadata creation", key)
				}
				data.Metadata = append(data.Metadata, &Variable{Name: key, Type: varType})
			}
			// sort to impose consistent ordering
			sort.SliceStable(data.Metadata, func(i, j int) bool {
				return data.Metadata[i].Name < data.Metadata[j].Name
			})
		}

		// Create a temporary metadata -> index map.  Required because the variable data for each hit returned
		//  from ES is an unordered key/value list.
		metadataIndex := make(map[string]int, len(data.Metadata))
		for idx, value := range data.Metadata {
			metadataIndex[value.Name] = idx
		}

		// extract data for all variables
		values := make([]interface{}, len(data.Metadata))
		for key, variable := range variables {
			index := metadataIndex[key]
			result, ok := json.Interface(variable, VariableValueField)
			if !ok {
				log.Errorf("%+v", err)
			}
			values[index] = result
		}
		// add the row to the variable data
		data.Values = append(data.Values, values)
	}
	return &data, nil
}

// FetchFilteredData creates an ES query to fetch a set of documents.  Applies filters to restrict the
// results to a user selected set of fields, with documents further filtered based on allowed ranges and
// categories.
func FetchFilteredData(client *elastic.Client, dataset string, filterParams *FilterParams) (*FilteredData, error) {
	// construct an ES query that fetches documents from the dataset with the supplied variable filters applied
	query := elastic.NewBoolQuery()
	var keys []string
	for _, variable := range filterParams.Ranged {
		query = query.Filter(elastic.NewRangeQuery(variable.Name + ".value").Gte(variable.Min).Lte(variable.Max))
		keys = append(keys, variable.Name)
	}
	for _, variable := range filterParams.Categorical {
		query = query.Filter(elastic.NewTermsQuery(variable.Name+".value", variable.Categories))
		keys = append(keys, variable.Name)
	}
	for _, variableName := range filterParams.None {
		keys = append(keys, variableName)
	}

	fetchContext := elastic.NewFetchSourceContext(true).Include(keys...)

	// execute the ES query
	res, err := client.Search().
		Query(query).
		Index(dataset).
		Size(filterParams.Size).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do()
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch filtered data query failed")
	}

	// parse the result
	return parseResults(res)
}
