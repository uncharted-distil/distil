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
	catResultLimit = 100
)

func (s *Storage) getHistogramAggQuery(extrema *model.Extrema) (string, string, string) {
	interval := extrema.GetBucketInterval()

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", model.HistogramAggPrefix, extrema.Name)
	bucketQueryString := fmt.Sprintf("width_bucket(\"%s\", %g, %g, %d) - 1",
		extrema.Name, extrema.Min, extrema.Max, extrema.GetBucketCount())
	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, extrema.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (s *Storage) parseNumericHistogram(varType string, rows *pgx.Rows, extrema *model.Extrema) (*model.Histogram, error) {
	// get histogram agg name
	histogramAggName := model.HistogramAggPrefix + extrema.Name

	// Parse bucket results.
	interval := extrema.GetBucketInterval()

	buckets := make([]*model.Bucket, extrema.GetBucketCount())
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

		// Since the max can match the limit, an extra bucket may exist.
		// Add the value to the second to last bucket.
		if bucket < int64(len(buckets)) {
			buckets[bucket].Count = bucketCount
		} else {
			buckets[len(buckets)-1].Count += bucketCount
		}
	}
	// assign histogram attributes
	return &model.Histogram{
		Name:    extrema.Name,
		Type:    model.NumericalType,
		VarType: varType,
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
	min := int64(math.MaxInt32)
	max := int64(-math.MaxInt32)

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
			if bucketCount < min {
				min = bucketCount
			}
			if bucketCount > max {
				max = bucketCount
			}
		}
	}

	// assign histogram attributes
	return &model.Histogram{
		Name:    variable.Name,
		Type:    model.CategoricalType,
		VarType: variable.Type,
		Buckets: buckets,
		Extrema: &model.Extrema{
			Min: float64(min),
			Max: float64(max),
		},
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
			err := rows.Scan(&targetTerm, &predictedTerm, &bucketCount)
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
	min := int64(math.MaxInt32)
	max := int64(-math.MaxInt32)

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
		if bucket.Count < min {
			min = bucket.Count
		}
		if bucket.Count > max {
			max = bucket.Count
		}
	}
	// assign histogram attributes
	return &model.Histogram{
		Name:    variable.Name,
		VarType: variable.Type,
		Type:    model.CategoricalType,
		Buckets: buckets,
		Extrema: &model.Extrema{
			Min: float64(min),
			Max: float64(max),
		},
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

func (s *Storage) fetchExtremaByURI(dataset string, resultURI string, variable *model.Variable) (*model.Extrema, error) {
	// add min / max aggregation
	aggQuery := s.getMinMaxAggsQuery(variable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $1;", aggQuery, dataset, s.getResultTable(dataset), d3mIndexFieldName)

	// execute the postgres query
	// NOTE: We may want to use the regular Query operation since QueryRow
	// hides db exceptions.
	res, err := s.client.Query(queryString, resultURI)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseExtrema(res, variable)
}

// FetchExtremaByURI return extrema of a variable in a result set.
func (s *Storage) FetchExtremaByURI(dataset string, resultURI string, index string, varName string) (*model.Extrema, error) {

	variable, err := s.metadata.FetchVariable(dataset, index, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}
	return s.fetchExtremaByURI(dataset, resultURI, variable)
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

	return s.parseNumericHistogram(variable.Type, res, extrema)
}

func (s *Storage) fetchNumericalHistogramByResult(dataset string, variable *model.Variable, resultURI string, extrema *model.Extrema) (*model.Histogram, error) {
	// need the extrema to calculate the histogram interval
	var err error
	if extrema == nil {
		extrema, err = s.fetchExtremaByURI(dataset, resultURI, variable)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch variable extrema for summary")
		}
	} else {
		extrema.Name = variable.Name
		extrema.Type = variable.Type
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := s.getHistogramAggQuery(extrema)

	// Create the complete query string.
	query := fmt.Sprintf("SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $1 GROUP BY %s ORDER BY %s;",
		bucketQuery, histogramQuery, histogramName, dataset,
		s.getResultTable(dataset), d3mIndexFieldName, bucketQuery, histogramName)

	// execute the postgres query
	res, err := s.client.Query(query, resultURI)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseNumericHistogram(variable.Type, res, extrema)
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

func (s *Storage) fetchCategoricalHistogramByResult(dataset string, variable *model.Variable, resultURI string) (*model.Histogram, error) {
	// Get count by category.
	query := fmt.Sprintf("SELECT data.\"%s\", COUNT(*) AS count FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $1 GROUP BY \"%s\" ORDER BY count desc, \"%s\" LIMIT %d;", variable.Name, dataset, s.getResultTable(dataset),
		d3mIndexFieldName, variable.Name, variable.Name, catResultLimit)

	// execute the postgres query
	res, err := s.client.Query(query, resultURI)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseCategoricalHistogram(res, variable)
}

func (s *Storage) fetchSummaryData(dataset string, index string, varName string, resultURI string, extrema *model.Extrema) (*model.Histogram, error) {
	// need description of the variables to request aggregation against.
	variable, err := s.metadata.FetchVariable(dataset, index, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}

	var histogram *model.Histogram

	if model.IsNumerical(variable.Type) {
		// fetch numeric histograms
		if resultURI == "" {
			histogram, err = s.fetchNumericalHistogram(dataset, variable)
		} else {
			histogram, err = s.fetchNumericalHistogramByResult(dataset, variable, resultURI, extrema)
		}
		if err != nil {
			return nil, errors.Wrap(err, "unable to get numerical histogram")
		}
	} else if model.IsCategorical(variable.Type) {
		// fetch categorical histograms
		if resultURI == "" {
			histogram, err = s.fetchCategoricalHistogram(dataset, variable)
		} else {
			histogram, err = s.fetchCategoricalHistogramByResult(dataset, variable, resultURI)
		}
		if err != nil {
			return nil, errors.Wrap(err, "unable to get categorical histogram")
		}
	} else {
		return nil, errors.Errorf("variable %s of type %s does not support summary", variable.Name, variable.Type)
	}

	// get number of rows
	numRows, err := s.FetchNumRows(dataset, nil)
	if err != nil {
		return nil, err
	}
	histogram.NumRows = numRows

	// add dataset
	histogram.Dataset = dataset

	return histogram, err
}

// FetchSummary returns the summary for the provided dataset and variable.
func (s *Storage) FetchSummary(dataset string, index string, varName string) (*model.Histogram, error) {
	return s.fetchSummaryData(dataset, index, varName, "", nil)
}

// FetchSummaryByResult returns the summary for the provided dataset
// and variable for data that is part of the result set.
func (s *Storage) FetchSummaryByResult(dataset string, index string, varName string, resultURI string, extrema *model.Extrema) (*model.Histogram, error) {
	return s.fetchSummaryData(dataset, index, varName, resultURI, extrema)
}
