package routes

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"goji.io/pat"
	"gopkg.in/olivere/elastic.v2"

	"github.com/unchartedsoftware/distil/api/util/json"
)

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

// VariableSummariesHandler generates a route handler that facilitates the creation and retrieval
// of summary information about the variables in a datset.  Currently this consists of a histogram
// for each variable, but can be extended to support avg, std dev, percentiles etc.  in th future.
func VariableSummariesHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get dataset id
		datasetID := dataset + "_dataset"

		log.Infof("Processing variables summaries request for %s", dataset)

		// Need list of variables to request aggregation against.
		variables, err := fetchVariables(client, index, datasetID)
		if err != nil {
			handleServerError(errors.Wrap(err, "Failed to fetch variable list for summary generation"), w)
			return
		}

		// Create a query that does min and max aggregations for each variable
		search := client.Search().
			Index(dataset).
			Size(0)

		// for each variable, create a min / max aggregation
		for _, variable := range variables {
			if histogramVariable(variable.Name, variable.Type) {
				// get field name
				field := fmt.Sprintf("%s.value", name)
				// create aggs
				minAgg := elastic.NewMinAggregation().Field(field)
				maxAgg := elastic.NewMaxAggregation().Field(field)
				// create agg names
				minAggName := fmt.Sprintf("min__%s", name)
				maxAggName := fmt.Sprintf("max__%s", name)
				// add aggregations
				search.
					Aggregation(minAggName, minAgg).
					Aggregation(maxAggName, maxAgg)
			}
		}

		// execute the search
		res, err := search.Do()
		if err != nil {
			handleServerError(errors.Wrap(err, "Failed to execute min/max aggregation query for summary generation"), w)
			return
		}

		// For each returned aggregation, create a histogram aggregation.  Bucket size is derived from
		// the min/max and desired bucket count.
		search = client.Search().
			Index(dataset).
			Size(0)

		for _, variable := range variables {

			name := variable.Name
			minAggName := fmt.Sprintf("min__%s", name)
			maxAggName := fmt.Sprintf("max__%s", name)

			minAgg, ok := res.Aggregations.Min(minAggName)
			if !ok {
				continue
			}

			maxAgg, ok := searchResult.Aggregations.Max(maxAggName)
			if !ok {
				continue
			}

			if minAgg.Value == nil  && maxAgg.Value == nil {
				continue
			}

			// compute the bucket interval for the histogram
			// TODO: ES v5 supports float intervals for histograms. Need to
			// upgrade frm v2 and make thisuse floats.
			interval := int64(math.Floor((maxAgg.Value - minAgg.Value) / 100))
			if interval < 1 {
				interval = 1
			}

			// update the histogram aggregation request
			histogramAgg := elastic.NewHistogramAggregation().
				Field(name + ".value").
				Interval(interval)
			search.Aggregation(name, histogramAgg)
		}

		// Execute the search
		res, err = search.Do()
		if err != nil {
			handleServerError(errors.Wrap(err, "Failed to fetch histograms for variables summaries"), w)
			return
		}

		// Parse the results and store in structs for marshalling to JSON
		var result histogramList
		for name, aggregation := range res.Aggregations {

			// Pull the data for each aggregation out into JSON rep
			// json, err := aggregation.MarshalJSON()
			// if err != nil {
			// 	log.Warnf("%+v", errors.Wrapf(err, "Failed to marshal JSON entry for %s", name))
			// 	continue
			// }

			aggJSON, err := json.Unmarshal(aggregation)
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
