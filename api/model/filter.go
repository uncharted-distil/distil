package model

import (
	"sort"

	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v3"

	"github.com/unchartedsoftware/distil/api/filter"
	"github.com/unchartedsoftware/distil/api/util/json"
)

// FilteredData provides the metadata and raw data values that match a supplied
// input filter.
type FilteredData struct {
	Name     string          `json:"name"`
	Metadata []*Variable     `json:"metadata"`
	Values   [][]interface{} `json:"values"`
}

func parseFilteredData(res *elastic.SearchResult) (*FilteredData, error) {
	data := &FilteredData{}

	for idx, hit := range res.Hits.Hits {
		// parse hit into JSON
		src, err := json.Unmarshal(*hit.Source)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse data")
		}

		// parse source into variables map
		variables, ok := json.Map(src)
		if !ok {
			return nil, errors.Wrap(err, "failed to parse data")
		}

		// On the first time through, parse out name/type info and store that in a header.  We also
		// store the name/type tuples in a map for quick lookup
		if idx == 0 {
			data.Name = hit.Index
			for key, variable := range variables {
				varType, ok := json.String(variable, VariableSchemaTypeField)
				if !ok {
					return nil, errors.Errorf("failed to extract type info for %s during metadata creation", key)
				}
				// append variable metadata
				data.Metadata = append(data.Metadata, &Variable{
					Name: key,
					Type: varType,
				})

			}
			// sort by name for deterministic results
			sort.SliceStable(data.Metadata, func(i, j int) bool {
				return data.Metadata[i].Name < data.Metadata[j].Name
			})
		}

		// create a temporary metadata -> index map. Required because the
		// variable data for each hit returned from ES is unordered.
		metadataIndex := make(map[string]int, len(data.Metadata))
		for idx, value := range data.Metadata {
			metadataIndex[value.Name] = idx
		}

		// extract data for all variables
		values := make([]interface{}, len(data.Metadata))
		for key, variable := range variables {
			value, ok := json.Interface(variable, VariableValueField)
			if !ok {
				continue
			}
			index := metadataIndex[key]
			values[index] = value
		}
		// add the row to the variable data
		data.Values = append(data.Values, values)
	}
	return data, nil
}

// FetchFilteredData creates an ES query to fetch a set of documents. Applies
// filters to restrict the results to a user selected set of fields, with
// documents further filtered based on allowed ranges and categories.
func FetchFilteredData(client *elastic.Client, dataset string, set *filter.Set) (*FilteredData, error) {
	// execute the filter set elasticsearch query
	res, err := set.Search(client, dataset)
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch filtered data query failed")
	}

	// parse the result
	return parseFilteredData(res)
}
