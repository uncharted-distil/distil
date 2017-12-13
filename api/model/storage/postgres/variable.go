package postgres

import (
	"fmt"
	"math"
	"strconv"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

const (
	catResultLimit = 10
)

func (s *Storage) calculateInterval(extrema *model.Extrema) float64 {
	// compute the bucket interval for the histogram
	interval := (extrema.Max - extrema.Min) / model.MaxNumBuckets
	if !model.IsFloatingPoint(extrema.Type) {
		interval = math.Floor(interval)
		interval = math.Max(1, interval)
	}
	return interval
}

func (s *Storage) getHistogramAggQuery(extrema *model.Extrema) (string, string, string) {
	interval := s.calculateInterval(extrema)

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", model.HistogramAggPrefix, extrema.Name)
	bucketQueryString := fmt.Sprintf("width_bucket(\"%s\", %g, %g, %d) -1",
		extrema.Name, extrema.Min, extrema.Max, model.MaxNumBuckets-1)
	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, extrema.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (s *Storage) parseNumericHistogram(rows *pgx.Rows, extrema *model.Extrema) (*model.Histogram, error) {
	// get histogram agg name
	histogramAggName := model.HistogramAggPrefix + extrema.Name

	// Parse bucket results.
	interval := s.calculateInterval(extrema)

	buckets := make([]*model.Bucket, model.MaxNumBuckets)
	key := extrema.Min
	for i := 0; i < len(buckets); i++ {
		keyString := ""
		if model.IsFloatingPoint(extrema.Type) {
			keyString = fmt.Sprintf("%f", key)
		} else {
			keyString = strconv.Itoa(int(key))
		}

		buckets[i] = &model.Bucket{
			Key:   keyString,
			Count: 0,
		}

		key = key + interval
	}
	for rows.Next() {
		var bucketValue float64
		var bucketCount int64
		var bucket int64
		err := rows.Scan(&bucket, &bucketValue, &bucketCount)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", histogramAggName))
		}
		buckets[bucket].Count = bucketCount
	}
	// assign histogram attributes
	return &model.Histogram{
		Name:    extrema.Name,
		Type:    "numerical",
		Extrema: extrema,
		Buckets: buckets,
	}, nil
}

func (s *Storage) parseCategoricalHistogram(rows *pgx.Rows, variable *model.Variable) (*model.Histogram, error) {
	termsAggName := model.TermsAggPrefix + variable.Name

	// parse as either one dimension or two dimension category histogram.  This could be collapsed down into a
	// single function.
	dimension := len(rows.FieldDescriptions()) - 1
	if dimension == 1 {
		return parseUnivariateCategoricalHistogram(rows, variable, termsAggName)
	} else if dimension == 2 {
		return parseBivariateCategoricalHistogram(rows, variable, termsAggName)
	} else {
		return nil, errors.Errorf("Unhandled dimension of %d for histogram %s", dimension, termsAggName)
	}
}

func parseUnivariateCategoricalHistogram(rows *pgx.Rows, variable *model.Variable, termsAggName string) (*model.Histogram, error) {
	// Parse bucket results.
	buckets := make([]*model.Bucket, 0)
	if rows != nil {
		for rows.Next() {
			var term string
			var bucketCount int64
			err := rows.Scan(&term, &bucketCount)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", termsAggName))
			}

			buckets = append(buckets, &model.Bucket{
				Key:   term,
				Count: bucketCount,
			})
		}
	}

	// assign histogram attributes
	return &model.Histogram{
		Name:    variable.Name,
		Type:    model.CategoricalType,
		Buckets: buckets,
	}, nil
}

func parseBivariateCategoricalHistogram(rows *pgx.Rows, variable *model.Variable, termsAggName string) (*model.Histogram, error) {
	// extract the counts
	countMap := map[string]map[string]int64{}
	if rows != nil {
		for rows.Next() {
			var predictedTerm string
			var targetTerm string
			var bucketCount int64
			err := rows.Scan(&predictedTerm, &targetTerm, &bucketCount)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", termsAggName))
			}
			if len(countMap[predictedTerm]) == 0 {
				countMap[predictedTerm] = map[string]int64{}
			}
			countMap[predictedTerm][targetTerm] = bucketCount
		}
	}

	// convert the extracted counts into buckets suitable for serialization
	buckets := make([]*model.Bucket, 0)
	for predictedKey, targetCounts := range countMap {
		bucket := model.Bucket{
			Key:     predictedKey,
			Count:   0,
			Buckets: []*model.Bucket{},
		}
		for targetKey, count := range targetCounts {
			targetBucket := model.Bucket{
				Key:   targetKey,
				Count: count,
			}
			bucket.Count = bucket.Count + count
			bucket.Buckets = append(bucket.Buckets, &targetBucket)
		}
		buckets = append(buckets, &bucket)
	}

	// assign histogram attributes
	return &model.Histogram{
		Name:    variable.Name,
		Type:    model.CategoricalType,
		Buckets: buckets,
	}, nil
}

func (s *Storage) parseExtrema(row *pgx.Rows, variable *model.Variable) (*model.Extrema, error) {
	var minValue *float64
	var maxValue *float64
	if row != nil {
		// Expect one row of data.
		row.Next()
		err := row.Scan(&minValue, &maxValue)
		if err != nil {
			return nil, errors.Wrap(err, "no min / max aggregation found")
		}
	}
	// check values exist
	if minValue == nil || maxValue == nil {
		return nil, errors.Errorf("no min / max aggregation values found")
	}
	// assign attributes
	return &model.Extrema{
		Name: variable.Name,
		Type: variable.Type,
		Min:  *minValue,
		Max:  *maxValue,
	}, nil
}

func (s *Storage) getMinMaxAggsQuery(variable *model.Variable) string {
	// get min / max agg names
	minAggName := model.MinAggPrefix + variable.Name
	maxAggName := model.MaxAggPrefix + variable.Name

	// create aggregations
	queryPart := fmt.Sprintf("MIN(\"%s\") AS \"%s\", MAX(\"%s\") AS \"%s\"", variable.Name, minAggName, variable.Name, maxAggName)
	// add aggregations
	return queryPart
}

func (s *Storage) fetchExtrema(dataset string, variable *model.Variable) (*model.Extrema, error) {
	// add min / max aggregation
	aggQuery := s.getMinMaxAggsQuery(variable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s;", aggQuery, dataset)

	// execute the postgres query
	// NOTE: We may want to use the regular Query operation since QueryRow
	// hides db exceptions.
	res, err := s.client.Query(queryString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseExtrema(res, variable)
}

func (s *Storage) fetchNumericalHistogram(dataset string, variable *model.Variable) (*model.Histogram, error) {
	// need the extrema to calculate the histogram interval
	extrema, err := s.fetchExtrema(dataset, variable)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable extrema for summary")
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := s.getHistogramAggQuery(extrema)

	// Create the complete query string.
	query := fmt.Sprintf("SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count FROM %s GROUP BY %s ORDER BY %s;",
		bucketQuery, histogramQuery, histogramName, dataset, bucketQuery, histogramName)

	// execute the postgres query
	res, err := s.client.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseNumericHistogram(res, extrema)
}

func (s *Storage) fetchCategoricalHistogram(dataset string, variable *model.Variable) (*model.Histogram, error) {
	// Get count by category.
	query := fmt.Sprintf("SELECT \"%s\", COUNT(*) AS count FROM %s GROUP BY \"%s\" ORDER BY count desc, \"%s\" LIMIT %d;", variable.Name, dataset, variable.Name, variable.Name, catResultLimit)

	// execute the postgres query
	res, err := s.client.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseCategoricalHistogram(res, variable)
}

// FetchSummary returns the summary for the provided dataset and variable.
func (s *Storage) FetchSummary(dataset string, index string, varName string) (*model.Histogram, error) {
	// need description of the variables to request aggregation against.
	variable, err := s.metadata.FetchVariable(dataset, index, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}

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
