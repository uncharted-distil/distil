//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package postgres

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// TextField defines behaviour for the text field type.
type TextField struct {
	Storage     *Storage
	StorageName string
	Key         string
	Label       string
	Type        string
}

// NewTextField creates a new field for text types.
func NewTextField(storage *Storage, storageName string, key string, label string, typ string) *TextField {
	field := &TextField{
		Storage:     storage,
		StorageName: storageName,
		Key:         key,
		Label:       label,
		Type:        typ,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *TextField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error
	if resultURI == "" {
		baseline, err = f.fetchHistogram(nil, invert)
		if err != nil {
			return nil, err
		}
		if !filterParams.Empty() {
			filtered, err = f.fetchHistogram(filterParams, invert)
			if err != nil {
				return nil, err
			}
		}
	} else {
		baseline, err = f.fetchHistogramByResult(resultURI, nil)
		if err != nil {
			return nil, err
		}
		if !filterParams.Empty() {
			filtered, err = f.fetchHistogramByResult(resultURI, filterParams)
			if err != nil {
				return nil, err
			}
		}
	}

	return &api.VariableSummary{
		Label:    f.Label,
		Key:      f.Key,
		Type:     model.CategoricalType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
	}, nil
}

func (f *TextField) getTimeMinMaxAggsQuery(timeVar *model.Variable) string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + timeVar.Name
	maxAggName := api.MaxAggPrefix + timeVar.Name

	timeSelect := fmt.Sprintf("CAST(\"%s\" AS INTEGER)", timeVar.Name)
	if timeVar.Type == model.DateTimeType {
		timeSelect = fmt.Sprintf("CAST(extract(epoch from \"%s\") AS INTEGER)", timeVar.Name)
	}

	// create aggregations
	queryPart := fmt.Sprintf("MIN(%s) AS \"%s\", MAX(%s) AS \"%s\"",
		timeSelect, minAggName, timeSelect, maxAggName)
	// add aggregations
	return queryPart
}

func (f *TextField) fetchTimeExtrema(timeVar *model.Variable) (*api.Extrema, error) {

	// add min / max aggregation
	aggQuery := f.getTimeMinMaxAggsQuery(timeVar)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s;", aggQuery, f.StorageName)

	// execute the postgres query
	res, err := f.Storage.client.Query(queryString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseTimeExtrema(timeVar, res)
}

func (f *TextField) fetchTimeExtremaByResultURI(timeVar *model.Variable, resultURI string) (*api.Extrema, error) {

	// add min / max aggregation
	aggQuery := f.getTimeMinMaxAggsQuery(timeVar)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $1;",
		aggQuery, f.StorageName, f.Storage.getResultTable(f.StorageName), model.D3MIndexFieldName)

	// execute the postgres query
	res, err := f.Storage.client.Query(queryString, resultURI)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch time extrema by result URI for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseTimeExtrema(timeVar, res)
}

func (f *TextField) parseTimeExtrema(timeVar *model.Variable, rows *pgx.Rows) (*api.Extrema, error) {
	var minValue *int64
	var maxValue *int64
	if rows != nil {
		// Expect one row of data.
		exists := rows.Next()
		if !exists {
			return nil, fmt.Errorf("no rows in extrema query result")
		}
		err := rows.Scan(&minValue, &maxValue)
		if err != nil {
			return nil, errors.Wrap(err, "no min / max aggregation found")
		}
	}
	// check values exist
	if minValue == nil || maxValue == nil {
		return nil, errors.Errorf("no min / max aggregation values found")
	}
	// assign attributes
	return &api.Extrema{
		Key:  timeVar.Name,
		Type: timeVar.Type,
		Min:  float64(*minValue),
		Max:  float64(*maxValue),
	}, nil
}

func (f *TextField) getTimeseriesHistogramAggQuery(extrema *api.Extrema, interval int) (string, string, string) {

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", api.HistogramAggPrefix, extrema.Key)

	binning := extrema.GetTimeseriesBinningArgs(interval)

	timeSelect := fmt.Sprintf("CAST(\"%s\" AS INTEGER)", extrema.Key)
	if extrema.Type == model.DateTimeType {
		timeSelect = fmt.Sprintf("CAST(extract(epoch from \"%s\") AS INTEGER)", extrema.Key)
	}

	bucketQueryString := ""
	// if only a single value, then return a simple count.
	if binning.Rounded.Max == binning.Rounded.Min {
		// want to return the count under bucket 0.
		bucketQueryString = fmt.Sprintf("(%s - %s)", timeSelect, timeSelect)
	} else {
		bucketQueryString = fmt.Sprintf("width_bucket(%s, %d, %d, %d) - 1",
			timeSelect, int(binning.Rounded.Min), int(binning.Rounded.Max), binning.Count)
	}

	histogramQueryString := fmt.Sprintf("(%s) * %d + %d", bucketQueryString, int(binning.Interval), int(binning.Rounded.Min))

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (f *TextField) parseTimeHistogram(rows *pgx.Rows, extrema *api.Extrema, interval int) (*api.Histogram, error) {
	// get histogram agg name
	histogramAggName := api.HistogramAggPrefix + extrema.Key

	// Parse bucket results
	binning := extrema.GetTimeseriesBinningArgs(interval)

	keys := make([]string, binning.Count)
	key := binning.Rounded.Min
	for i := 0; i < len(keys); i++ {
		keyString := ""
		if model.IsFloatingPoint(extrema.Type) {
			keyString = fmt.Sprintf("%f", key)
		} else {
			keyString = strconv.Itoa(int(key))
		}

		keys[i] = keyString

		key = key + binning.Interval
	}

	categoryBuckets := make(map[string][]*api.Bucket)

	for rows.Next() {
		var bucketValue float64
		var bucketCount int64
		var bucket int64
		var category string
		err := rows.Scan(&bucket, &bucketValue, &category, &bucketCount)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", histogramAggName))
		}

		buckets, ok := categoryBuckets[category]
		if !ok {
			buckets = make([]*api.Bucket, binning.Count)
			for i := range buckets {
				buckets[i] = &api.Bucket{
					Count: 0,
					Key:   keys[i],
				}
			}
			categoryBuckets[category] = buckets
		}

		if bucket < 0 {
			// Due to float representation, sometimes the lowest value <
			// first bucket interval and so ends up in bucket -1.
			buckets[0].Count = bucketCount
		} else if bucket < int64(len(buckets)) {
			buckets[bucket].Count = bucketCount
		} else {
			// Since the max can match the limit, an extra bucket may exist.
			// Add the value to the second to last bucket.
			buckets[len(buckets)-1].Count += bucketCount
		}
	}
	// assign histogram attributes
	return &api.Histogram{
		Extrema:         binning.Rounded,
		CategoryBuckets: categoryBuckets,
	}, nil
}

func (f *TextField) getTopCategories(filterParams *api.FilterParams, invert bool) ([]string, error) {

	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, filterParams, invert)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT w.word as %s, COUNT(*) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(\"%s\"))) as stem FROM %s %s) as r "+
		"INNER JOIN %s as w on r.stem = w.stem "+
		"GROUP BY w.word ORDER BY count desc, w.word LIMIT %d;",
		f.Key, f.Key, f.StorageName, where, wordStemTableName, 5)

	// execute the postgres query
	rows, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch text histogram for variable summaries from postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	var categories []string
	if rows != nil {
		for rows.Next() {
			var category string
			var count int64
			err := rows.Scan(&category, &count)
			if err != nil {
				return nil, err
			}
			categories = append(categories, category)
		}
	}
	return categories, nil
}

func (f *TextField) fetchHistogram(filterParams *api.FilterParams, invert bool) (*api.Histogram, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, filterParams, invert)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT w.word as %s, COUNT(*) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(\"%s\"))) as stem FROM %s %s) as r "+
		"INNER JOIN %s as w on r.stem = w.stem "+
		"GROUP BY w.word ORDER BY count desc, w.word LIMIT %d;",
		f.Key, f.Key, f.StorageName, where, wordStemTableName, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch text histogram for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res)
}

func (f *TextField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams) (*api.Histogram, error) {

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.StorageName, resultURI, filterParams)
	if err != nil {
		return nil, err
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT w.word as \"%s\", COUNT(*) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(\"%s\"))) as stem "+
		"FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $%d %s) as r "+
		"INNER JOIN %s as w on r.stem = w.stem "+
		"GROUP BY w.word ORDER BY count desc, w.word LIMIT %d;",
		f.Key, f.Key, f.StorageName, f.Storage.getResultTable(f.StorageName),
		model.D3MIndexFieldName, len(params), where, wordStemTableName, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch text histogram for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res)
}

func (f *TextField) parseHistogram(rows *pgx.Rows) (*api.Histogram, error) {
	termsAggName := api.TermsAggPrefix + f.Key

	buckets := make([]*api.Bucket, 0)
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

			buckets = append(buckets, &api.Bucket{
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
	return &api.Histogram{
		Buckets: buckets,
		Extrema: &api.Extrema{
			Min: float64(min),
			Max: float64(max),
		},
	}, nil
}

// FetchPredictedSummaryData pulls data from the result table and builds
// the categorical histogram for the field.
func (f *TextField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	baseline, err = f.fetchPredictedSummaryData(resultURI, datasetResult, nil, extrema)
	if err != nil {
		return nil, err
	}
	if !filterParams.Empty() {
		filtered, err = f.fetchPredictedSummaryData(resultURI, datasetResult, filterParams, extrema)
		if err != nil {
			return nil, err
		}
	}
	return &api.VariableSummary{
		Label:    f.Label,
		Key:      f.Key,
		Type:     model.CategoricalType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
	}, nil
}

func (f *TextField) fetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	targetName := f.Key

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.StorageName, resultURI, filterParams)
	if err != nil {
		return nil, err
	}

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, targetName)

	query := fmt.Sprintf("SELECT word_b.word as \"%s\", word_v.word as value, COUNT(*) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(base.\"%s\"))) as stem_b, "+
		"unnest(tsvector_to_array(to_tsvector(result.value))) as stem_v "+
		"FROM %s AS result INNER JOIN %s AS base ON result.index = base.\"d3mIndex\" "+
		"WHERE %s) r INNER JOIN %s word_b ON r.stem_b = word_b.stem INNER JOIN %s word_v ON r.stem_v = word_v.stem "+
		"GROUP BY word_v.word, word_b.word "+
		"ORDER BY count desc;", targetName, targetName, datasetResult, f.StorageName, strings.Join(wheres, " AND "), wordStemTableName, wordStemTableName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return f.parseHistogram(res)
}
