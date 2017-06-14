package routes

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"goji.io/pat"
	"gopkg.in/olivere/elastic.v2"

	"github.com/unchartedsoftware/distil/api/util/json"
)

// Dataset represents a decsription of a dataset.
type Dataset struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Variables   []Variable `json:"variables"`
}

// DatasetResult represents the result of a datasets response.
type DatasetResult struct {
	Datasets []Dataset `json:"datasets"`
}

func parseDatasets(res *elastic.SearchResult) ([]Dataset, error) {
	var datasets []Dataset
	for _, hit := range res.Hits.Hits {
		// parse hit into JSON
		src, err := json.Unmarshal(*hit.Source)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse dataset")
		}
		// extract dataset name (ID is mirror of name)
		name := strings.TrimSuffix(hit.Id, "_dataset")
		// extract the description
		description, ok := json.String(src, "description")
		if !ok {
			log.Warnf("Description empty for %s", name)
		}
		// extract the variables list
		variables, err := parseVariables(hit)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse dataset")
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

func fetchDatasets(client *elastic.Client, index string) ([]Dataset, error) {
	log.Info("Processing dataset fetch request")

	fetchContext := elastic.NewFetchSourceContext(true).
		Include("_id", "description", "variables.varName", "variables.varType")

	// execute the ES query
	res, err := client.Search().
		Index(index).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Fields("_id").
		Do()
	if err != nil {
		return nil, errors.Wrap(err, "Elastic Search dataset fetch query failed")
	}

	return parseDatasets(res)
}

func searchDatasets(client *elastic.Client, index string, terms string) ([]Dataset, error) {
	log.Infof("Processing datasets search request for %s", terms)

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
		return nil, errors.Wrap(err, "Elastic search dataset search query failed")
	}

	return parseDatasets(res)
}

// DatasetsHandler generates a route handler that facilitates a search of
// dataset descriptions and variable names, returning a name, description and
// variable list for any dataset that matches. The search parameter is optional
// it contains the search terms if set, and if unset, flags that a list of all
// datasets should be returned.  The full list will be contain names only,
// descriptions and variable lists will not be included.
func DatasetsHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")
		// check for search terms
		terms, err := url.QueryUnescape(r.URL.Query().Get("search"))
		if err != nil {
			log.Error("Malformed datasets query")
			return
		}
		// if its present, forward a search, otherwise fetch all datasets
		var datasets []Dataset
		if terms != "" {
			datasets, err = searchDatasets(client, index, terms)
		} else {
			datasets, err = fetchDatasets(client, index)
		}
		if err != nil {
			handleServerError(err, w)
			return
		}
		// marshall data
		bytes, err := json.Marshal(DatasetResult{
			Datasets: datasets,
		})
		if err != nil {
			handleServerError(errors.Wrap(err, "Unable marshal dataset result into JSON"), w)
			return
		}
		// send response
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}
