// Package routes provides route handlers for the Distil server.
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	elastic "gopkg.in/olivere/elastic.v2"

	"strings"

	"net/url"

	"github.com/jeffail/gabs"
	log "github.com/unchartedsoftware/plog"
	"goji.io/pat"
)

const (
	datasetIndex = "datasets"
)

// EchoHandler generates a route a simple echo route handler for testing purposes
func EchoHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Processing echo request")
		fmt.Fprintf(w, "Distil - %s", pat.Param(r, "echo"))
	}
}

// FileHandler provides a static file lookup route using the OS file system
func FileHandler(rootDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(rootDir)).ServeHTTP(w, r)
	}
}

// VariablesHandler generates a variable listing route handler associated with the caller supplied
// ES endpoint
func VariablesHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Processing variables request for %s", pat.Param(r, "dataset"))
		datasetID := pat.Param(r, "dataset") + "_dataset"

		boolQuery := elastic.NewBoolQuery().
			Must(elastic.NewMatchQuery("_id", datasetID))

		fetchContext := elastic.NewFetchSourceContext(true)
		fetchContext.Include("variables")

		// execute the ES query
		searchResult, err := client.Search().
			Query(boolQuery).
			Index(datasetIndex).
			FetchSource(true).
			FetchSourceContext(fetchContext).
			Do()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err)
			return
		}

		// Extract output into JSON ready form
		type VarDesc struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}
		type Result struct {
			Variables []VarDesc `json:"variables"`
		}
		var result Result

		for _, hit := range searchResult.Hits.Hits {
			resultJSON, err := gabs.ParseJSON(*hit.Source)
			if err != nil {
				log.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			children, err := resultJSON.Path("variables").Children()
			if err != nil {
				log.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			for _, varData := range children {
				d := varData.Data().(map[string]interface{})
				result.Variables = append(result.Variables, VarDesc{
					Name: d["varName"].(string),
					Type: d["varType"].(string),
				})
			}
		}

		// Marshall output into JSON
		js, err := json.Marshal(result)
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func fetchDatasets(client *elastic.Client) ([]byte, error) {
	log.Info("Processing dataset fetch request")

	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include("_id", "description")

	// execute the ES query
	searchResult, err := client.Search().
		Index("datasets").
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Fields("_id").
		Do()

	if err != nil {
		return nil, err
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
		result.Datasets = append(result.Datasets, DatasetDesc{Name: datasetName})
	}

	// Marshall output into JSON
	js, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func searchDatasets(client *elastic.Client, terms string) ([]byte, error) {
	log.Infof("Processing datasets search request for %s", terms)
	query := elastic.NewMultiMatchQuery(terms, "_id", "description", "variables.varName").Analyzer("standard")

	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include("_id", "description", "variables.varName", "variables.varType")

	// execute the ES query
	searchResult, err := client.Search().
		Query(query).
		Index(datasetIndex).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do()

	if err != nil {
		return nil, err
	}

	// Structs for marshalling to JSON
	type Variable struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	type Dataset struct {
		Name        string     `json:"name"`
		Description string     `json:"description"`
		Variables   []Variable `json:"variables"`
	}

	type Results struct {
		Datasets []Dataset `json:"datasets"`
	}

	var result Results
	for _, hit := range searchResult.Hits.Hits {
		// parse hit into JSON
		resultJSON, err := gabs.ParseJSON(*hit.Source)
		if err != nil {
			return nil, err
		}

		// extract dataset name (ID is mirror of name)
		name := strings.TrimSuffix(hit.Id, "_dataset")

		// extract the description
		description, ok := resultJSON.Path("description").Data().(string)
		if !ok {
			log.Warnf("Description empty for %s", name)
		}

		// extract the variables list
		var variables []Variable
		children, err := resultJSON.Path("variables").Children()
		if len(children) == 0 || err != nil {
			log.Warnf("Variable list empty for %s", name)
		}
		for _, varData := range children {
			d := varData.Data().(map[string]interface{})
			variable := Variable{d["varName"].(string), d["varType"].(string)}
			variables = append(variables, variable)
		}

		// write everythign out to result struct
		result.Datasets = append(result.Datasets, Dataset{name, description, variables})
	}

	// Marshall results into JSON
	js, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return js, nil
}

// DatasetsHandler generates a route handler that facilitates a search of dataset descriptions
// and variable names, return matching dataset names as a result.
func DatasetsHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for search parameter
		terms, err := url.QueryUnescape(r.URL.Query().Get("search"))
		if err != nil {
			log.Error("Malformed datasets query")
			return
		}

		// if its present, forward a search, otherwise fetch all datasets
		var result []byte
		if terms != "" {
			result, err = searchDatasets(client, terms)
		} else {
			result, err = fetchDatasets(client)
		}

		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(result)
	}
}
