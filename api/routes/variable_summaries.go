package routes

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"goji.io/pat"
	"gopkg.in/olivere/elastic.v3"

	"github.com/unchartedsoftware/distil/api/util/json"
)

// Extrema represents the extrema for a single variable.
type Extrema struct {
	Name string
	Min  float64
	Max  float64
}

// Bucket represents a single histogram bucket.
type Bucket struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// Histogram represents a single variable histogram.
type Histogram struct {
	Name    string   `json:"name"`
	Buckets []Bucket `json:"buckets"`
}

// SummaryResult represents a summary response for a variable.
type SummaryResult struct {
	Histograms []Histogram `json:"histograms"`
}

func isOrdinal(name string, typ string) bool {
	if name == "d3mIndex" {
		return false
	}
	return typ == "long" ||
		typ == "integer" ||
		typ == "short" ||
		typ == "byte" ||
		typ == "double" ||
		typ == "float" ||
		typ == "date"
}

func parseExtrema(res *elastic.SearchResult, variables []Variable) ([]Extrema, error) {
	// parse extrema
	var extremas []Extrema
	for _, variable := range variables {
		// get min / max agg names
		minAggName := fmt.Sprintf("min_%s", variable.Name)
		maxAggName := fmt.Sprintf("max_%s", variable.Name)
		// check min agg
		minAgg, ok := res.Aggregations.Min(minAggName)
		if !ok {
			continue
		}
		// check max agg
		maxAgg, ok := res.Aggregations.Max(maxAggName)
		if !ok {
			continue
		}
		// check values exist
		if minAgg.Value == nil || maxAgg.Value == nil {
			continue
		}
		// append to extrema
		extremas = append(extremas, Extrema{
			Name: variable.Name,
			Min:  *minAgg.Value,
			Max:  *maxAgg.Value,
		})
	}
	return extremas, nil
}

func fetchExtrema(client *elastic.Client, dataset string, variables []Variable) ([]Extrema, error) {
	// create a query that does min and max aggregations for each variable
	search := client.Search().
		Index(dataset).
		Size(0)
	// for each variable, create a min / max aggregation
	for _, variable := range variables {
		if isOrdinal(variable.Name, variable.Type) {
			// get field name
			field := fmt.Sprintf("%s.value", variable.Name)
			// get min / max agg names
			minAggName := fmt.Sprintf("min_%s", variable.Name)
			maxAggName := fmt.Sprintf("max_%s", variable.Name)
			// create aggregations
			minAgg := elastic.NewMinAggregation().Field(field)
			maxAgg := elastic.NewMaxAggregation().Field(field)
			// add aggregations
			search.
				Aggregation(minAggName, minAgg).
				Aggregation(maxAggName, maxAgg)
		}
	}
	// execute the search
	res, err := search.Do()
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute min/max aggregation query for summary generation")
	}
	return parseExtrema(res, variables)
}

func parseHistograms(res *elastic.SearchResult, extremas []Extrema) ([]Histogram, error) {
	// parse histograms
	var histograms []Histogram
	for _, extrema := range extremas {
		// get histogram agg
		agg, ok := res.Aggregations.Histogram(extrema.Name)
		if !ok {
			continue
		}
		// get histogram buckets
		var buckets []Bucket
		for _, bucket := range agg.Buckets {
			buckets = append(buckets, Bucket{
				Key:   strconv.Itoa(int(bucket.Key)),
				Count: bucket.DocCount,
			})
		}
		// create histogram
		histograms = append(histograms, Histogram{
			Name:    extrema.Name,
			Buckets: buckets,
		})
	}
	return histograms, nil
}

func fetchHistograms(client *elastic.Client, dataset string, extremas []Extrema) ([]Histogram, error) {
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	search := client.Search().
		Index(dataset).
		Size(0)
	// for each extreama, create a histogram aggregation
	for _, extrema := range extremas {
		name := extrema.Name
		// compute the bucket interval for the histogram
		// TODO: ES v5 supports float intervals for histograms. Need to
		// upgrade frm v2 and make thisuse floats.
		interval := int64(math.Floor((extrema.Max - extrema.Min) / 100))
		if interval < 1 {
			interval = 1
		}
		// create histogram agg
		histogramAgg := elastic.NewHistogramAggregation().
			Field(name + ".value").
			Interval(interval)
		// add histogram agg
		search.Aggregation(name, histogramAgg)
	}
	// execute the search
	res, err := search.Do()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variables summaries")
	}
	return parseHistograms(res, extremas)
}

func fetchSummaries(client *elastic.Client, index string, dataset string) ([]Histogram, error) {
	// need list of variables to request aggregation against.
	variables, err := fetchVariables(client, index, dataset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variables for summary generation")
	}
	// need the extrema of each var to calculate the histrogram interval
	extremas, err := fetchExtrema(client, dataset, variables)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable extrema for summary generation")
	}
	return fetchHistograms(client, dataset, extremas)
}

// VariableSummariesHandler generates a route handler that facilitates the
// creation and retrieval of summary information about the variables in a
// dataset.  Currently this consists of a histogram for each variable, but can
// be extended to support avg, std dev, percentiles etc.  in th future.
func VariableSummariesHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")
		// get dataset name
		dataset := pat.Param(r, "dataset")

		log.Infof("Processing variables summaries request for %s", dataset)

		// fetch summary histogram
		histograms, err := fetchSummaries(client, index, dataset)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall output into JSON
		bytes, err := json.Marshal(SummaryResult{
			Histograms: histograms,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}
