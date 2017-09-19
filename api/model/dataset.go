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

func fetchDatasetSummary(client *elastic.Client, dataset string) (int64, int64, error) {
	// get stats about the index
	stats, err := client.IndexStats(dataset).Do(context.Background())
	if err != nil {
		return 0, 0, errors.Errorf("Error occurred while querying index stats for `%s`: %v",
			dataset,
			err)
	}
	// don't access by index name, it won't work if this is an alias to an
	// index. Since we are doing a query for a specific index already, there
	// should be only one index in the response.
	if len(stats.Indices) < 1 {
		return 0, 0, errors.Errorf("Index `%s` does not exist", dataset)
	}
	// grab the first index in the map (there should only be one)
	var indexStats *elastic.IndexStats
	for _, value := range stats.Indices {
		indexStats = value
		break
	}
	// get number of documents
	numDocs := int64(0)
	// ensure no nil pointers
	if indexStats.Primaries != nil &&
		indexStats.Primaries.Docs != nil {
		numDocs = indexStats.Primaries.Docs.Count
	}
	// get the btye size
	byteSize := int64(0)
	// ensure no nil pointers
	if indexStats.Primaries != nil &&
		indexStats.Primaries.Store != nil {
		byteSize = indexStats.Primaries.Store.SizeInBytes
	}
	return numDocs, byteSize, nil
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
			Variables:   variables,
		})
	}
	// get index stats
	for _, dataset := range datasets {
		numRows, numBytes, err := fetchDatasetSummary(client, dataset.Name)
		if err != nil {
			return nil, errors.Wrap(err, "elasticsearch dataset index stats failed")
		}
		dataset.NumRows = numRows
		dataset.NumBytes = numBytes
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
