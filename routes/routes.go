// Package routes provides route handlers for the Distil server.
package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	elastic "gopkg.in/olivere/elastic.v2"

	"strings"

	"github.com/jeffail/gabs"
	log "github.com/unchartedsoftware/plog"
	"goji.io/pat"
)

// Route function type that all Goji route handlers must adhere to
type Route func(w http.ResponseWriter, r *http.Request)

// EchoHandler generates a route a simple echo route handler for testing purposes
func EchoHandler() Route {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Processing echo request")
		fmt.Fprintf(w, "Distil - %s", pat.Param(r, "echo"))
	}
}

// DatasetsHandler generates a dataset listing route handler associated with the caller supplied
// ES endpoint.
func DatasetsHandler(client *elastic.Client) Route {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Processing dataset request")
		boolQuery := elastic.NewBoolQuery().
			Must(elastic.NewMatchQuery("rawData", "false"))

		// execute the ES query
		searchResult, err := client.Search().
			Query(boolQuery).
			Index("data-redacted").
			Fields("_id").
			Pretty(true).
			Do()

		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

// VariablesHandler generates a variable listing route handler associated with the caller supplied
// ES endpoint
func VariablesHandler(client *elastic.Client) Route {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Processing variables request for %s", pat.Param(r, "dataset"))
		datasetID := pat.Param(r, "dataset") + "_dataset"

		boolQuery := elastic.NewBoolQuery().
			Must(elastic.NewMatchQuery("_id", datasetID))

		fetchContext := elastic.NewFetchSourceContext(true)
		fetchContext.Include("trainData.trainData")

		// execute the ES query
		searchResult, err := client.Search().
			Query(boolQuery).
			Index("data-redacted").
			Pretty(true).
			FetchSource(true).
			FetchSourceContext(fetchContext).
			Do()

		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			children, err := resultJSON.Path("trainData.trainData").Children()
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
