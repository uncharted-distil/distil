package model

import (
	"math"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v3"
)

const (
	// MinAggPrefix is the prefix used for min aggregations.
	MinAggPrefix = "min_"
	// MaxAggPrefix is the prefix used for max aggregations.
	MaxAggPrefix = "max_"
	// TermsAggPrefix is the prefix used for terms aggregations.
	TermsAggPrefix = "terms_"
	// VariableValueField is the field which stores the variable value.
	VariableValueField = "value"
)

// Extrema represents the extrema for a single variable.
type Extrema struct {
	Name string  `json:"-"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

// Bucket represents a single histogram bucket.
type Bucket struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// Histogram represents a single variable histogram.
type Histogram struct {
	Name    string   `json:"name"`
	Extrema *Extrema `json:"extrema,omitempty"`
	Buckets []Bucket `json:"buckets"`
}

func isNumeric(name string, typ string) bool {
	if name == "d3mIndex" {
		return false
	}
	return typ == "integer" ||
		typ == "float" ||
		typ == "dateTime"
}

func isCategorical(name string, typ string) bool {
	if name == "d3mIndex" {
		return false
	}
	return typ == "categorical" ||
		typ == "ordinal" ||
		typ == "text"
}

func parseExtrema(res *elastic.SearchResult, variables []Variable) ([]Extrema, error) {
	// parse extrema
	var extremas []Extrema
	for _, variable := range variables {
		// get min / max agg names
		minAggName := MinAggPrefix + variable.Name
		maxAggName := MaxAggPrefix + variable.Name
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
		if isNumeric(variable.Name, variable.Type) {
			// get field name
			field := variable.Name + "." + VariableValueField
			// get min / max agg names
			minAggName := MinAggPrefix + variable.Name
			maxAggName := MaxAggPrefix + variable.Name
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

func parseNumericHistograms(res *elastic.SearchResult, extremas []Extrema) ([]Histogram, error) {
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
			Name: extrema.Name,
			Extrema: &Extrema{
				Min: extrema.Min,
				Max: extrema.Max,
			},
			Buckets: buckets,
		})
	}
	return histograms, nil
}

func fetchNumericalHistograms(client *elastic.Client, dataset string, extremas []Extrema) ([]Histogram, error) {
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
			Field(name + "." + VariableValueField).
			Interval(interval)
		// add histogram agg
		search.Aggregation(name, histogramAgg)
	}
	// execute the search
	res, err := search.Do()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variables summaries")
	}
	return parseNumericHistograms(res, extremas)
}

func parseCategoricalHistograms(res *elastic.SearchResult, variables []Variable) ([]Histogram, error) {
	// parse categories
	var histograms []Histogram
	for _, variable := range variables {
		// get terms agg name
		termsAggName := TermsAggPrefix + variable.Name
		// check terms agg
		terms, ok := res.Aggregations.Terms(termsAggName)
		if !ok {
			continue
		}
		// get histogram buckets
		var buckets []Bucket
		for _, bucket := range terms.Buckets {
			// check value exist
			buckets = append(buckets, Bucket{
				Key:   bucket.KeyNumber.String(),
				Count: bucket.DocCount,
			})
		}
		// create histogram
		histograms = append(histograms, Histogram{
			Name:    variable.Name,
			Buckets: buckets,
		})
	}
	return histograms, nil
}

func fetchCategoricalHistograms(client *elastic.Client, dataset string, variables []Variable) ([]Histogram, error) {
	// create a query that does min and max aggregations for each variable
	search := client.Search().
		Index(dataset).
		Size(0)
	// for each variable, create a min / max aggregation
	for _, variable := range variables {
		if isCategorical(variable.Name, variable.Type) {
			// get field name
			field := variable.Name + "." + VariableValueField
			// get terms agg name
			termsAggName := TermsAggPrefix + variable.Name
			// create aggregations
			termsAgg := elastic.NewTermsAggregation().Field(field)
			// add aggregations
			search.Aggregation(termsAggName, termsAgg)
		}
	}
	// execute the search
	res, err := search.Do()
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute terms aggregation query for summary")
	}
	return parseCategoricalHistograms(res, variables)
}

// FetchSummaries returns all the variable summaries for the provided index and
// dataset
func FetchSummaries(client *elastic.Client, index string, dataset string) ([]Histogram, error) {
	// need list of variables to request aggregation against.
	variables, err := FetchVariables(client, index, dataset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variables for summary")
	}
	// need the extrema of each var to calculate the histrogram interval
	extremas, err := fetchExtrema(client, dataset, variables)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable extrema for summary")
	}
	// fetch numeric histograms
	numeric, err := fetchNumericalHistograms(client, dataset, extremas)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch numerical histograms for summary")
	}
	// fetch categorical histograms
	categorical, err := fetchCategoricalHistograms(client, dataset, variables)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch categorical histograms for summary")
	}
	// merge
	return append(numeric, categorical...), nil
}
