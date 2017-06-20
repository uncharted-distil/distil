package model

import (
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v3"

	"github.com/unchartedsoftware/distil/api/util/json"
)

const (
	// DatasetSuffix is the suffix for the dataset entry when stored in
	// elasticsearch.
	DatasetSuffix = "_dataset"
)

// Dataset represents a decsription of a dataset.
type Dataset struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Variables   []Variable `json:"variables"`
}

func parseDatasets(res *elastic.SearchResult) ([]Dataset, error) {
	var datasets []Dataset
	for _, hit := range res.Hits.Hits {
		// parse hit into JSON
		src, err := json.Unmarshal(*hit.Source)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse dataset")
		}
		// extract dataset name (ID is mirror of name)
		name := strings.TrimSuffix(hit.Id, DatasetSuffix)
		// extract the description
		description, ok := json.String(src, "description")
		if !ok {
			description = ""
		}
		// extract the variables list
		variables, err := parseVariables(hit)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse dataset")
		}
		// write everythign out to result struct
		datasets = append(datasets, Dataset{
			Name:        name,
			Description: description,
			Variables:   variables,
		})
	}
	return datasets, nil
}

// FetchDatasets returns all datasets in the provided index.
func FetchDatasets(client *elastic.Client, index string) ([]Dataset, error) {
	fetchContext := elastic.NewFetchSourceContext(true).
		Include("_id", "description", "variables.varName", "variables.varType")

	// execute the ES query
	res, err := client.Search().
		Index(index).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do()
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset fetch query failed")
	}

	return parseDatasets(res)
}

// SearchDatasets returns the datasets that match the search criteria in the
// provided index.
func SearchDatasets(client *elastic.Client, index string, terms string) ([]Dataset, error) {
	query := elastic.NewMultiMatchQuery(terms, "_id", "description", "variables.varName").
		Analyzer("standard")

	fetchContext := elastic.NewFetchSourceContext(true).
		Include("_id", "description", "variables.varName", "variables.varType")

	// execute the ES query
	res, err := client.Search().
		Query(query).
		Index(index).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do()
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset search query failed")
	}

	return parseDatasets(res)
}
