package routes

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/jeffail/gabs"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"goji.io/pat"
	"gopkg.in/olivere/elastic.v2"
)

func parseDataset(searchResult *elastic.SearchResult) (*datasetList, error) {
	var result datasetList
	for _, hit := range searchResult.Hits.Hits {
		// parse hit into JSON
		resultJSON, err := gabs.ParseJSON(*hit.Source)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse dataset")
		}

		// extract dataset name (ID is mirror of name)
		name := strings.TrimSuffix(hit.Id, "_dataset")

		// extract the description
		description, ok := resultJSON.Path("description").Data().(string)
		if !ok {
			log.Warnf("Description empty for %s", name)
		}

		// extract the variables list
		variables, err := parseVariables(hit)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse dataset")
		}

		// write everythign out to result struct
		result.Datasets = append(result.Datasets, dataset{name, description, *variables})
	}
	return &result, nil
}

func fetchDatasets(client *elastic.Client, index string) ([]byte, error) {
	log.Info("Processing dataset fetch request")

	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include("_id", "description")

	// execute the ES query
	searchResult, err := client.Search().
		Index(index).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Fields("_id").
		Do()

	if err != nil {
		return nil, errors.Wrap(err, "Elastic Search dataset fetch query failed")
	}

	// Extract output into JSON ready form
	type DatasetDesc struct {
		Name string `json:"name"`
	}
	type Result struct {
		Datasets []DatasetDesc `json:"datasets"`
	}
	var result Result
	for _, hit := range searchResult.Hits.Hits {
		datasetName := strings.TrimSuffix(hit.Id, "_dataset")
		result.Datasets = append(result.Datasets, DatasetDesc{datasetName})
	}

	// Marshall output into JSON
	js, err := json.Marshal(result)
	if err != nil {
		return nil, errors.Wrap(err, "Unable marshal result into JSON")
	}
	return js, nil
}

func searchDatasets(client *elastic.Client, index string, terms string) ([]byte, error) {
	log.Infof("Processing datasets search request for %s", terms)
	query := elastic.NewMultiMatchQuery(terms, "_id", "description", "variables.varName").Analyzer("standard")

	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include("_id", "description", "variables.varName", "variables.varType")

	// execute the ES query
	searchResult, err := client.Search().
		Query(query).
		Index(index).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do()

	if err != nil {
		return nil, errors.Wrap(err, "Elastic search dataset search query failed")
	}

	result, err := parseDataset(searchResult)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse dataset search result")
	}

	// Marshall results into JSON
	js, err := json.Marshal(result)
	if err != nil {
		return nil, errors.Wrap(err, "Unable marshal dataset result into JSON")
	}
	return js, nil
}

// DatasetsHandler generates a route handler that facilitates a search of dataset descriptions
// and variable names, returning a name, description and variable list for any dataset that matches.
// The search parameter is optional - it contains the search terms if set, and if unset, flags that
// a list of all datasets should be returned.  The full list will be contain names only - descriptions
// and variable lists will not be included.
func DatasetsHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")

		// check for search parameter
		terms, err := url.QueryUnescape(r.URL.Query().Get("search"))
		if err != nil {
			log.Error("Malformed datasets query")
			return
		}

		// if its present, forward a search, otherwise fetch all datasets
		var result []byte
		if terms != "" {
			result, err = searchDatasets(client, index, terms)
		} else {
			result, err = fetchDatasets(client, index)
		}

		if err != nil {
			handleServerError(err, w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(result)
	}
}
