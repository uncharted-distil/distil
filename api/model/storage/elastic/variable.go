package elastic

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"gopkg.in/olivere/elastic.v5"
)

func (s *Storage) getNumericalVariables(variables []*model.Variable) []*model.Variable {
	var result []*model.Variable
	for _, variable := range variables {
		if model.IsNumerical(variable.Type) {
			result = append(result, variable)
		}
	}
	return result
}

func (s *Storage) getCategoricalVariables(variables []*model.Variable) []*model.Variable {
	var result []*model.Variable
	for _, variable := range variables {
		if model.IsCategorical(variable.Type) {
			result = append(result, variable)
		}
	}
	return result
}

func (s *Storage) parseExtrema(res *elastic.SearchResult, variable *model.Variable) (*model.Extrema, error) {
	// get min / max agg names
	minAggName := model.MinAggPrefix + variable.Name
	maxAggName := model.MaxAggPrefix + variable.Name
	// check min agg
	minAgg, ok := res.Aggregations.Min(minAggName)
	if !ok {
		return nil, errors.Errorf("no %s aggregation found", minAggName)
	}
	// check max agg
	maxAgg, ok := res.Aggregations.Max(maxAggName)
	if !ok {
		return nil, errors.Errorf("no %s aggregation found", maxAggName)
	}
	// check values exist
	if minAgg.Value == nil || maxAgg.Value == nil {
		return nil, errors.Errorf("no min / max aggregation values found")
	}
	// assign attributes
	return &model.Extrema{
		Name: variable.Name,
		Type: variable.Type,
		Min:  *minAgg.Value,
		Max:  *maxAgg.Value,
	}, nil
}

func (s *Storage) appendMinMaxAggs(search *elastic.SearchService, variable *model.Variable) *elastic.SearchService {
	// get field name
	field := variable.Name + "." + model.VariableValueField
	// get min / max agg names
	minAggName := model.MinAggPrefix + variable.Name
	maxAggName := model.MaxAggPrefix + variable.Name
	// create aggregations
	minAgg := elastic.NewMinAggregation().Field(field)
	maxAgg := elastic.NewMaxAggregation().Field(field)
	// add aggregations
	return search.
		Aggregation(minAggName, minAgg).
		Aggregation(maxAggName, maxAgg)
}

func (s *Storage) fetchExtrema(dataset string, variable *model.Variable) (*model.Extrema, error) {
	// create a query that does min and max aggregations for each variable
	search := s.client.Search().
		Index(dataset).
		Size(0)
	// add min / max aggregation
	s.appendMinMaxAggs(search, variable)
	// execute the search
	res, err := search.Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute min/max aggregation query for summary generation")
	}
	return s.parseExtrema(res, variable)
}

func (s *Storage) parseNumericHistogram(res *elastic.SearchResult, extrema *model.Extrema) (*model.Histogram, error) {
	// get histogram agg name
	histogramAggName := model.HistogramAggPrefix + extrema.Name
	// get histogram agg
	agg, ok := res.Aggregations.Histogram(histogramAggName)
	if !ok {
		return nil, errors.Errorf("no %s aggregation found", histogramAggName)
	}
	// get histogram buckets
	var buckets []*model.Bucket
	for _, bucket := range agg.Buckets {
		var key string
		if extrema.Type == model.FloatType {
			key = fmt.Sprintf("%f", bucket.Key)
		} else {
			key = strconv.Itoa(int(bucket.Key))
		}
		buckets = append(buckets, &model.Bucket{
			Key:   key,
			Count: bucket.DocCount,
		})
	}
	// assign histogram attributes
	return &model.Histogram{
		Name:    extrema.Name,
		Type:    "numerical",
		Extrema: extrema,
		Buckets: buckets,
	}, nil
}

func (s *Storage) appendHistogramAgg(search *elastic.SearchService, extrema *model.Extrema) *elastic.SearchService {
	// compute the bucket interval for the histogram
	interval := (extrema.Max - extrema.Min) / model.MaxNumBuckets
	if extrema.Type != model.FloatType {
		interval = math.Floor(interval)
		interval = math.Max(1, interval)
	}
	// get histogram agg name
	histogramAggName := model.HistogramAggPrefix + extrema.Name
	// create histogram agg
	histogramAgg := elastic.NewHistogramAggregation().
		Field(extrema.Name + "." + model.VariableValueField).
		Interval(interval)
	// add histogram agg
	return search.Aggregation(histogramAggName, histogramAgg)
}

func (s *Storage) fetchNumericalHistogram(dataset string, variable *model.Variable) (*model.Histogram, error) {
	// need the extrema to calculate the histogram interval
	extrema, err := s.fetchExtrema(dataset, variable)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable extrema for summary")
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	search := s.client.Search().
		Index(dataset).
		Size(0)
	// add histogram agg
	s.appendHistogramAgg(search, extrema)
	// execute the search
	res, err := search.Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variables summaries")
	}
	return s.parseNumericHistogram(res, extrema)
}

func (s *Storage) parseCategoricalHistogram(res *elastic.SearchResult, variable *model.Variable) (*model.Histogram, error) {
	// get terms agg name
	termsAggName := model.TermsAggPrefix + variable.Name
	// check terms agg
	terms, ok := res.Aggregations.Terms(termsAggName)
	if !ok {
		return nil, errors.Errorf("no %s aggregation found", termsAggName)
	}
	// get histogram buckets
	var buckets []*model.Bucket
	for _, bucket := range terms.Buckets {
		// check value exist
		buckets = append(buckets, &model.Bucket{
			Key:   bucket.KeyNumber.String(),
			Count: bucket.DocCount,
		})
	}
	// assign histogram attributes
	return &model.Histogram{
		Name:    variable.Name,
		Type:    "categorical",
		Buckets: buckets,
	}, nil
}

func (s *Storage) appendTermsAgg(search *elastic.SearchService, variable *model.Variable) *elastic.SearchService {
	// get field name
	field := variable.Name + "." + model.VariableValueField
	// get terms agg name
	termsAggName := model.TermsAggPrefix + variable.Name
	// create aggregation
	termsAgg := elastic.NewTermsAggregation().Field(field)
	// add aggregations
	return search.Aggregation(termsAggName, termsAgg)
}

func (s *Storage) fetchCategoricalHistogram(dataset string, variable *model.Variable) (*model.Histogram, error) {
	// create a query that does min and max aggregations for each variable
	search := s.client.Search().
		Index(dataset).
		Size(0)
	// add terms aggregation
	s.appendTermsAgg(search, variable)
	// execute the search
	res, err := search.Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to execute terms aggregation query for summary, %v", res))
	}
	return s.parseCategoricalHistogram(res, variable)
}

// FetchSummary returns the summary for the provided dataset and variable.
func (s *Storage) FetchSummary(variable *model.Variable, dataset string) (*model.Histogram, error) {
	if model.IsNumerical(variable.Type) {
		// fetch numeric histograms
		numeric, err := s.fetchNumericalHistogram(dataset, variable)
		if err != nil {
			return nil, err
		}
		return numeric, nil
	}
	if model.IsCategorical(variable.Type) {
		// fetch categorical histograms
		categorical, err := s.fetchCategoricalHistogram(dataset, variable)
		if err != nil {
			return nil, err
		}
		return categorical, nil
	}
	if model.IsText(variable.Type) {
		// fetch text analysis
		return nil, nil
	}
	return nil, errors.Errorf("variable %s of type %s does not support summary", variable.Name, variable.Type)
}
