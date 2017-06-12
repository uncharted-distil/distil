// Package routes provides route handlers for the Distil server.  Convenient
// as a single file for now, but will probably need to be broken out by route
// as more are added.
package routes

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	elastic "gopkg.in/olivere/elastic.v2"

	"strings"

	"net/url"

	"github.com/jeffail/gabs"
	"github.com/pkg/errors"
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

func handleServerError(err error, w http.ResponseWriter) {
	log.Error(errors.Cause(err))
	http.Error(w, errors.Cause(err).Error(), http.StatusInternalServerError)
}

type varDesc struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type variableList struct {
	Variables []varDesc `json:"variables"`
}

type dataset struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	variableList
}

type datasetList struct {
	Datasets []dataset `json:"datasets"`
}

func parseVariables(searchHit *elastic.SearchHit) (*variableList, error) {
	var result variableList

	resultJSON, err := gabs.ParseJSON(*searchHit.Source)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse search result")
	}
	children, err := resultJSON.Path("variables").Children()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse variables from search result result")
	}
	for _, varData := range children {
		d := varData.Data().(map[string]interface{})
		result.Variables = append(result.Variables, varDesc{
			Name: d["varName"].(string),
			Type: d["varType"].(string),
		})
	}

	return &result, nil
}

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

func fetchVariables(client *elastic.Client, dataset string) ([]byte, error) {
	boolQuery := elastic.NewBoolQuery().
		Must(elastic.NewMatchQuery("_id", dataset))

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
		return nil, errors.Wrap(err, "ElasticSearch variable fetch query failed")
	}
	if len(searchResult.Hits.Hits) != 1 {
		return nil, errors.New("ElasticSearch variable fetch query returned > 1 results")
	}

	// Extract output into JSON ready structs
	variables, err := parseVariables(searchResult.Hits.Hits[0])
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse search result JSON")
	}

	// Marshall output into JSON
	js, err := json.Marshal(variables)
	if err != nil {
		return nil, errors.Wrap(err, "Unable marshal result into JSON")
	}

	return js, err
}

// VariablesHandler generates a variable listing route handler associated with the caller supplied
// ES endpoint.  The handler returns a list of name/type tuples for the given dataset.
func VariablesHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Processing variables request for %s", pat.Param(r, "dataset"))
		datasetID := pat.Param(r, "dataset") + "_dataset"

		js, err := fetchVariables(client, datasetID)
		if err != nil {
			handleServerError(err, w)
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
			handleServerError(err, w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(result)
	}
}

type bucketEntry struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

type histogram struct {
	Name    string        `json:"name"`
	Buckets []bucketEntry `json:"buckets"`
}

type histogramList struct {
	Histograms []histogram `json:"histograms"`
}

func histogramVariable(varName string, varType string) bool {
	return varName != "d3mIndex" && (varType == "integer" || varType == "float")
}

func parseRangeAggregation(name string, aggMsg *json.RawMessage) (float64, error) {
	// extract the min / max for each variable
	json, err := aggMsg.MarshalJSON()
	if err != nil {
		return math.NaN(), errors.Wrapf(err, "Failed to marshall range data for histogram %s", name)
	}
	aggJSON, err := gabs.ParseJSON(json)
	if err != nil {
		return math.NaN(), errors.Wrapf(err, "Failed to parse range data for histogram %s", name)
	}
	return aggJSON.Path("value").Data().(float64), nil
}

// VariableSummariesHandler generates a route handler that facilitates the creation and retrieval
// of summary information about the variables in a datset.  Currently this consists of a histogram
// for each variable, but can be extended to support avg, std dev, percentiles etc.  in th future.
func VariableSummariesHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Processing variables summaries request for %s", pat.Param(r, "dataset"))
		datasetName := pat.Param(r, "dataset")
		datasetID := datasetName + "_dataset"

		// Need list of variables to request aggregation against.
		variablesJSON, err := fetchVariables(client, datasetID)
		if err != nil {
			handleServerError(errors.Wrap(err, "Failed to fetch variable list for summary generation"), w)
			return
		}
		parsedVariables, err := gabs.ParseJSON(variablesJSON)
		if err != nil {
			handleServerError(errors.Wrap(err, "Failed to parse variable list for summary generation"), w)
			return
		}
		variables, err := parsedVariables.Path("variables").Children()
		if err != nil {
			handleServerError(errors.Wrap(err, "Failed to parse variable list for summary generation"), w)
			return
		}

		// Create a query that does min and max aggregations for each variable
		search := client.Search().
			Index(datasetName).
			Size(0)

		var variableNames []string
		for _, variable := range variables {
			// parse out the name and type
			name := variable.Path("name").Data().(string)
			varType := variable.Path("type").Data().(string)

			// for those that can have a histogram generated, create min and max aggregation
			if name != "" && histogramVariable(name, varType) {
				variableNames = append(variableNames, name)

				esFieldName := fmt.Sprintf("%s.value", name)

				minAggregation := elastic.NewMinAggregation().Field(esFieldName)
				maxAggregation := elastic.NewMaxAggregation().Field(esFieldName)

				minAggName := fmt.Sprintf("min__%s", name)
				maxAggName := fmt.Sprintf("max__%s", name)

				search = search.Aggregation(minAggName, minAggregation).
					Aggregation(maxAggName, maxAggregation)
			}
		}

		// Execute the search
		searchResult, err := search.Do()
		if err != nil {
			handleServerError(errors.Wrap(err, "Failed to execute min/max aggregation query for summary generation"), w)
			return
		}

		// For each returned aggregation, create a histogram aggregation.  Bucket size is derived from
		// the min/max and desired bucket count.
		search = client.Search().
			Index(datasetName).
			Size(0)

		for _, name := range variableNames {
			minAgg := searchResult.Aggregations["min__"+name]
			maxAgg := searchResult.Aggregations["max__"+name]

			minVal, err := parseRangeAggregation(name, minAgg)
			if err != nil {
				log.Error(errors.Cause(err))
				continue
			}
			maxVal, err := parseRangeAggregation(name, maxAgg)
			if err != nil {
				log.Error(errors.Cause(err))
				continue
			}

			// compute the bucket interval for the histogram
			// TODO: ES v 5 supports float intervals for histograms.  Need to upgrade frm v2 and make this
			// use floats.
			interval := int64(math.Floor((maxVal - minVal) / 100))
			if interval < 1 {
				interval = 1
			}

			// update the histogram aggregation request
			histogramAggregation := elastic.NewHistogramAggregation().Field(name + ".value").Interval(interval)
			search = search.Aggregation(name, histogramAggregation)
		}

		// Execute the search
		searchResult, err = search.Do()
		if err != nil {
			handleServerError(errors.Wrap(err, "Failed to fetch histograms for variables summaries"), w)
			return
		}

		// Parse the results and store in structs for marshalling to JSON
		var result histogramList
		for name, aggregation := range searchResult.Aggregations {

			// Pull the data for each aggregation out into JSON rep
			json, err := aggregation.MarshalJSON()
			if err != nil {
				log.Warnf("%+v", errors.Wrapf(err, "Failed to marshal JSON entry for %s", name))
				continue
			}
			aggJSON, err := gabs.ParseJSON(json)
			if err != nil {
				log.Warnf("%+v", errors.Wrapf(err, "Failed to parse JSON entry for %s", name))
				continue
			}

			buckets, err := aggJSON.Path("buckets").Children()
			if err != nil {
				log.Warnf("%+v", errors.Wrapf(err, "Failed to extract buckets from JSON entry %s", name))
				continue
			}

			// Convert the JSON into the struct hierarchy we want to return to the client
			var histogram histogram
			histogram.Name = name
			for _, bucket := range buckets {
				key, ok := bucket.Path("key").Data().(float64)
				if ok {
					count, ok := bucket.Path("doc_count").Data().(float64)
					if ok {
						strKey := strconv.FormatFloat(key, 'f', -1, 64)
						histogram.Buckets = append(histogram.Buckets, bucketEntry{strKey, int64(count)})
					}
				}
				if len(histogram.Buckets) == 0 {
					log.Warnf("Failed to find histogram data for %s", name)
				}
			}
			result.Histograms = append(result.Histograms, histogram)
		}

		// Marshall output into JSON
		js, err := json.Marshal(result)
		if err != nil {
			handleServerError(err, w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
