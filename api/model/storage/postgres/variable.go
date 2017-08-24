package postgres

import (
	"fmt"
	"math"
	"strconv"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

func (s *Storage) getHistogramAggQuery(extrema *model.Extrema) (string, string) {
	// compute the bucket interval for the histogram
	interval := (extrema.Max - extrema.Min) / model.MaxNumBuckets
	if extrema.Type != model.FloatType {
		// smallest bin for integers is 1
		interval = math.Max(1, interval)
	}

	// get histogram agg name & query string.
	histogramAggName := model.HistogramAggPrefix + extrema.Name
	histogramQueryString := fmt.Sprintf("(%s / %f) * %f", extrema.Name, interval, interval)

	return histogramAggName, histogramQueryString
}

func (s *Storage) parseNumericHistogram(rows *pgx.Rows, extrema *model.Extrema) (*model.Histogram, error) {
	// get histogram agg name
	histogramAggName := model.HistogramAggPrefix + extrema.Name

	// Parse bucket results.
	buckets := make([]*model.Bucket, 0)
	for rows.Next() {
		var bucketValue float64
		var bucketCount int64
		err := rows.Scan(&bucketValue, &bucketCount)
		if err != nil {
			return nil, errors.Errorf("no %s histogram aggregation found", histogramAggName)
		}

		var key string
		if extrema.Type == model.FloatType {
			key = fmt.Sprintf("%f", bucketValue)
		} else {
			key = strconv.Itoa(int(bucketValue))
		}
		buckets = append(buckets, &model.Bucket{
			Key:   key,
			Count: bucketCount,
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

func (s *Storage) parseCategoricalHistogram(rows *pgx.Rows, variable *model.Variable) (*model.Histogram, error) {
	// get terms agg name
	termsAggName := model.TermsAggPrefix + variable.Name

	// Parse bucket results.
	buckets := make([]*model.Bucket, 0)
	if rows != nil {
		for rows.Next() {
			var term string
			var bucketCount int64
			err := rows.Scan(&term, &bucketCount)
			if err != nil {
				return nil, errors.Errorf("no %s histogram aggregation found", termsAggName)
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
		Type:    "categorical",
		Buckets: buckets,
	}, nil
}

func (s *Storage) parseExtrema(row *pgx.Row, variable *model.Variable) (*model.Extrema, error) {
	var minValue *float64
	var maxValue *float64
	if row != nil {
		err := row.Scan(&minValue, &maxValue)
		if err != nil {
			return nil, errors.Errorf("no min/max aggregation found")
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
	queryPart := fmt.Sprintf("MIN(%s) AS %s, MAX(%s) AS %s", variable.Name, minAggName, variable.Name, maxAggName)
	// add aggregations
	return queryPart
}

func (s *Storage) fetchExtrema(dataset string, variable *model.Variable) (*model.Extrema, error) {
	// add min / max aggregation
	aggQuery := s.getMinMaxAggsQuery(variable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s;", aggQuery, dataset)

	// execute the postgres query
	// NOTE: We may want to use the refular Query operation since QueryRow
	// hides db exceptions.
	res := s.client.QueryRow(queryString)

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
	histogramName, histogramQuery := s.getHistogramAggQuery(extrema)

	// Create the complete query string.
	query := fmt.Sprintf("SELECT (%s) AS %s, COUNT(*) AS count FROM %s GROUP BY %s ORDER BY %s;", histogramQuery, histogramName, dataset, histogramQuery, histogramName)

	// execute the postgres query
	res, err := s.client.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	return s.parseNumericHistogram(res, extrema)
}

func (s *Storage) fetchCategoricalHistogram(dataset string, variable *model.Variable) (*model.Histogram, error) {
	// Get count by category.
	query := fmt.Sprintf("SELECT %s, COUNT(*) AS count FROM %s GROUP BY %s;", variable.Name, dataset, variable.Name)

	// execute the postgres query
	res, err := s.client.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
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
