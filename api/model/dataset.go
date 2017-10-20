package model

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/util/json"
	"gopkg.in/olivere/elastic.v5"
)

const (
	// DatasetSuffix is the suffix for the dataset entry when stored in
	// elasticsearch.
	DatasetSuffix = "_dataset"
	metadataType  = "metadata"
)

// Dataset represents a decsription of a dataset.
type Dataset struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Summary     string      `json:"summary"`
	Variables   []*Variable `json:"variables"`
	NumRows     int64       `json:"numRows"`
	NumBytes    int64       `json:"numBytes"`
}

func parseDatasets(client *elastic.Client, res *elastic.SearchResult) ([]*Dataset, error) {
	var datasets []*Dataset
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
		// extract the summary
		summary, ok := json.String(src, "summary")
		if !ok {
			summary = ""
		}
		// extract the number of rows
		numRows, ok := json.Int(src, "numRows")
		if !ok {
			summary = ""
		}
		// extract the number of bytes
		numBytes, ok := json.Int(src, "numBytes")
		if !ok {
			summary = ""
		}
		// extract the variables list
		variables, err := parseVariables(hit)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse dataset")
		}
		// write everythign out to result struct
		datasets = append(datasets, &Dataset{
			Name:        name,
			Description: description,
			Summary:     summary,
			NumRows:     int64(numRows),
			NumBytes:    int64(numBytes),
			Variables:   variables,
		})
	}
	return datasets, nil
}

// FetchDatasets returns all datasets in the provided index.
func FetchDatasets(client *elastic.Client, index string) ([]*Dataset, error) {
	// execute the ES query
	res, err := client.Search().
		Index(index).
		FetchSource(true).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset fetch query failed")
	}
	return parseDatasets(client, res)
}

// SearchDatasets returns the datasets that match the search criteria in the
// provided index.
func SearchDatasets(client *elastic.Client, index string, terms string) ([]*Dataset, error) {
	query := elastic.NewMultiMatchQuery(terms, "_id", "description", "variables.varName").
		Analyzer("standard")
	// execute the ES query
	res, err := client.Search().
		Query(query).
		Index(index).
		FetchSource(true).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset search query failed")
	}
	return parseDatasets(client, res)
}

// SetDataType updates the data type of the field in ES. NOTE: Not implemented!
func SetDataType(client *elastic.Client, dataset string, index string, field string, fieldType string) error {
	// Fetch all existing variables
	vars, err := FetchVariables(client, index, dataset)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch existing variable")
	}

	// Update only the variable we care about
	for _, v := range vars {
		if v.Name == field {
			v.Type = fieldType
		}
	}

	source := map[string]interface{}{
		Variables: vars,
	}

	// push the document into the metadata index
	_, err = client.Update().
		Index(index).
		Type(metadataType).
		Id(dataset + DatasetSuffix).
		Doc(source).
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to add document to index `%s`", index)
	}
	return nil
}
