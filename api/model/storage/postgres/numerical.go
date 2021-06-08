//
//   Copyright © 2021 Uncharted Software Inc.
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
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	log "github.com/unchartedsoftware/plog"
)

// NumericalField defines behaviour for the numerical field type.
type NumericalField struct {
	BasicField
	subSelect func() string
}

// NumericalStats contains summary information on a numerical fields.
type NumericalStats struct {
	StdDev          float64
	Mean            float64
	NoDataAvailable bool
}

// NewNumericalField creates a new field for numerical types.
func NewNumericalField(storage *Storage, datasetName string, datasetStorageName string, key string, label string, typ string, count string) *NumericalField {
	count = getCountSQL(count)

	field := &NumericalField{
		BasicField: BasicField{
			Storage:            storage,
			DatasetName:        datasetName,
			DatasetStorageName: datasetStorageName,
			Key:                key,
			Label:              label,
			Type:               typ,
			Count:              count,
		},
	}

	return field
}

// NewNumericalFieldSubSelect creates a new field for numerical types
// and specifies a sub select query to pull the raw data.
func NewNumericalFieldSubSelect(storage *Storage, datasetName string, datasetStorageName string, key string, label string, typ string, count string, fieldSubSelect func() string) *NumericalField {
	count = getCountSQL(count)

	field := &NumericalField{
		BasicField: BasicField{
			Storage:            storage,
			DatasetName:        datasetName,
			DatasetStorageName: datasetStorageName,
			Key:                key,
			Label:              label,
			Type:               typ,
			Count:              count,
		},
		subSelect: fieldSubSelect,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *NumericalField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {

	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	if resultURI == "" {
		baseline, err = f.fetchHistogram(api.GetBaselineFilter(filterParams), api.MaxNumBuckets)
		if err != nil {
			return nil, err
		}
		if !filterParams.IsEmpty(true) {
			filtered, err = f.fetchHistogram(filterParams, api.MaxNumBuckets)
			if err != nil {
				return nil, err
			}
		}
	} else {
		baseline, err = f.fetchHistogramByResult(resultURI, api.GetBaselineFilter(filterParams), extrema, api.MaxNumBuckets)
		if err != nil {
			return nil, err
		}
		if !filterParams.IsEmpty(true) {
			filtered, err = f.fetchHistogramByResult(resultURI, filterParams, extrema, api.MaxNumBuckets)
			if err != nil {
				return nil, err
			}
		}
	}
	return &api.VariableSummary{
		Label:    f.Label,
		Key:      f.Key,
		Type:     model.NumericalType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
	}, nil
}

func (f *NumericalField) fetchHistogram(filterParams *api.FilterParams, numBuckets int) (*api.Histogram, error) {
	return f.fetchHistogramWithJoins(filterParams, numBuckets, nil, []string{}, []interface{}{})
}

func (f *NumericalField) fetchHistogramWithJoins(filterParams *api.FilterParams, numBuckets int, joins []*joinDefinition, wheres []string, params []interface{}) (*api.Histogram, error) {
	fromClause := f.getFromClause(true)

	// create the filter for the query.
	wheres, params = f.Storage.buildFilteredQueryWhere(f.GetDatasetName(), wheres, params, "", filterParams)
	wheres = append(wheres, f.getDefaultFilter(true))

	// need the extrema to calculate the histogram interval
	extrema, err := f.fetchExtrema()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable extrema for summary")
	}
	if extrema == nil {
		log.Warnf("no extrema retrieved for variable summary")
		return nil, nil
	}

	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := f.getHistogramAggQuery(extrema, numBuckets, "")

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	joinSQL := createJoinStatements(joins)

	// Create the complete query string.
	query := fmt.Sprintf("SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(%s) AS count FROM %s %s %s GROUP BY %s ORDER BY %s;",
		bucketQuery, histogramQuery, histogramName, f.Count, fromClause, joinSQL, where, bucketQuery, histogramName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res, extrema, numBuckets)
	if err != nil {
		return nil, err
	}

	stats, err := f.FetchNumericalStats(filterParams)
	if err != nil {
		return nil, err
	}
	histogram.StdDev = stats.StdDev
	histogram.Mean = stats.Mean

	return histogram, nil
}

func (f *NumericalField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, numBuckets int) (*api.Histogram, error) {
	fromClause := f.getFromClause(false)
	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.GetDatasetName(), f.DatasetStorageName, resultURI, filterParams, baseTableAlias)
	if err != nil {
		return nil, err
	}
	wheres = append(wheres, f.getDefaultFilter(true))

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	// need the extrema to calculate the histogram interval
	if extrema == nil {
		extrema, err = f.fetchExtremaByURI(resultURI)
		if extrema == nil {
			return nil, nil
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch variable extrema for summary")
		}
	} else {
		extrema.Key = f.Key
		extrema.Type = f.Type
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := f.getHistogramAggQuery(extrema, numBuckets, baseTableAlias)

	// Create the complete query string.
	query := fmt.Sprintf(`
		SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(%s) AS count
		FROM %s INNER JOIN %s result ON %s."%s" = result.index
		WHERE result.result_id = $%d %s
		GROUP BY %s
		ORDER BY %s;`,
		bucketQuery, histogramQuery, histogramName, f.Count, fromClause,
		f.Storage.getResultTable(f.DatasetStorageName), baseTableAlias,
		model.D3MIndexFieldName, len(params), where, bucketQuery, histogramName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res, extrema, numBuckets)
	if err != nil {
		return nil, err
	}

	stats, err := f.FetchNumericalStatsByResult(resultURI, filterParams)
	if err != nil {
		return nil, err
	}
	histogram.StdDev = stats.StdDev
	histogram.Mean = stats.Mean

	return histogram, nil
}

func (f *NumericalField) fetchExtrema() (*api.Extrema, error) {
	fromClause := f.getFromClause(true)
	// add min / max aggregation
	aggQuery := f.getMinMaxAggsQuery()

	// create a query that does min and max aggregations for each variable
	// need to ignore the NaN values
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE %s;", aggQuery, fromClause, f.getDefaultFilter(true))

	// execute the postgres query
	// NOTE: We may want to use the regular Query operation since QueryRow
	// hides db exceptions.
	res, err := f.Storage.client.Query(queryString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseExtrema(res)
}

func (f *NumericalField) getHistogramAggQuery(extrema *api.Extrema, numBuckets int, alias string) (string, string, string) {
	interval := extrema.GetBucketInterval(numBuckets)

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", api.HistogramAggPrefix, extrema.Key)
	rounded := extrema.GetBucketMinMax(numBuckets)

	bucketQueryString := ""
	if alias != "" {
		alias = alias + "."
	}
	// if only a single value, then return a simple count.
	if rounded.Max == rounded.Min {
		// want to return the count under bucket 0.
		bucketQueryString = fmt.Sprintf("(%s\"%s\" - %s\"%s\")", alias, extrema.Key, alias, extrema.Key)
	} else {
		bucketQueryString = fmt.Sprintf("width_bucket(%s\"%s\", %g, %g, %d) - 1",
			alias, extrema.Key, rounded.Min, rounded.Max, extrema.GetBucketCount(numBuckets))
	}

	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, rounded.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (f *NumericalField) parseHistogram(rows pgx.Rows, extrema *api.Extrema, numBuckets int) (*api.Histogram, error) {
	// get histogram agg name
	histogramAggName := api.HistogramAggPrefix + extrema.Key

	// Parse bucket results.
	interval := extrema.GetBucketInterval(numBuckets)

	buckets := make([]*api.Bucket, extrema.GetBucketCount(numBuckets))
	rounded := extrema.GetBucketMinMax(numBuckets)
	key := rounded.Min
	for i := 0; i < len(buckets); i++ {
		keyString := ""
		if model.IsFloatingPoint(extrema.Type) {
			keyString = fmt.Sprintf("%f", key)
		} else {
			keyString = strconv.Itoa(int(key))
		}

		buckets[i] = &api.Bucket{
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
	err := rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	// Assign histogram attributes.  Extrema reflects the extrema of the data the histogram
	// is created from, not the extrema of the buckets.
	return &api.Histogram{
		Extrema: extrema,
		Buckets: buckets,
	}, nil
}

func (f *NumericalField) parseExtrema(rows pgx.Rows) (*api.Extrema, error) {
	var minValue *float64
	var maxValue *float64
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
		log.Warnf("no min / max aggregation values found")
		return nil, nil
	}
	// assign attributes
	return &api.Extrema{
		Key:  f.Key,
		Type: f.Type,
		Min:  *minValue,
		Max:  *maxValue,
	}, nil
}

func (f *NumericalField) getMinMaxAggsQuery() string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + f.Key
	maxAggName := api.MaxAggPrefix + f.Key

	// create aggregations
	queryPart := fmt.Sprintf("MIN(\"%s\") AS \"%s\", MAX(\"%s\") AS \"%s\"",
		f.Key, minAggName, f.Key, maxAggName)
	// add aggregations
	return queryPart
}

func (f *NumericalField) fetchExtremaByURI(resultURI string) (*api.Extrema, error) {
	fromClause := f.getFromClause(false)

	// add min / max aggregation
	aggQuery := f.getMinMaxAggsQuery()

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s INNER JOIN %s result ON %s.\"%s\" = result.index WHERE result.result_id = $1 AND %s;",
		aggQuery, fromClause, f.Storage.getResultTable(f.DatasetStorageName), baseTableAlias, model.D3MIndexFieldName, f.getDefaultFilter(true))

	// execute the postgres query
	// NOTE: We may want to use the regular Query operation since QueryRow
	// hides db exceptions.
	res, err := f.Storage.client.Query(queryString, resultURI)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseExtrema(res)
}

// FetchPredictedSummaryData pulls data from the result table and builds
// the numerical histogram for the field.
func (f *NumericalField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	baseline, err = f.fetchPredictedSummaryData(resultURI, datasetResult, nil, extrema, api.MaxNumBuckets)
	if err != nil {
		return nil, err
	}
	if !filterParams.IsEmpty(true) {
		filtered, err = f.fetchPredictedSummaryData(resultURI, datasetResult, filterParams, extrema, api.MaxNumBuckets)
		if err != nil {
			return nil, err
		}
	}
	return &api.VariableSummary{
		Label:    f.Label,
		Key:      f.Key,
		Type:     model.NumericalType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
	}, nil
}

func (f *NumericalField) fetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema, numBuckets int) (*api.Histogram, error) {
	resultVariable := &model.Variable{
		Key:  "value",
		Type: model.StringType,
	}

	// need the extrema to calculate the histogram interval
	var err error
	if extrema == nil {
		extrema, err = f.fetchResultsExtrema(resultURI, datasetResult, resultVariable)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch result variable extrema for summary")
		}
	} else {
		extrema.Key = f.Key
		extrema.Type = f.Type
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := f.getResultHistogramAggQuery(extrema, resultVariable, numBuckets)

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.GetDatasetName(), f.DatasetStorageName, resultURI, filterParams, baseTableAlias)
	if err != nil {
		return nil, err
	}

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, f.Key)
	wheres = append(wheres, fmt.Sprintf("%s != ''", resultVariable.Key))

	// Create the complete query string.
	query := fmt.Sprintf(`
		SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(%s) AS count
		FROM %s data INNER JOIN %s result ON data."%s" = result.index
		WHERE %s
		GROUP BY %s
		ORDER BY %s;`,
		bucketQuery, histogramQuery, histogramName, f.Count, f.DatasetStorageName, datasetResult,
		model.D3MIndexFieldName, strings.Join(wheres, " AND "), bucketQuery, histogramName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result variable summaries from postgres")
	}
	defer res.Close()

	return f.parseHistogram(res, extrema, numBuckets)
}

func (f *NumericalField) getResultMinMaxAggsQuery(resultVariable *model.Variable) string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + resultVariable.Key
	maxAggName := api.MaxAggPrefix + resultVariable.Key

	// Only numeric types should occur.
	fieldTyped := fmt.Sprintf("CAST(CASE WHEN \"%s\" = '' THEN 'NaN' ELSE \"%s\" END as double precision)", resultVariable.Key, resultVariable.Key)

	// create aggregations
	queryPart := fmt.Sprintf("MIN(%s) AS \"%s\", MAX(%s) AS \"%s\"", fieldTyped, minAggName, fieldTyped, maxAggName)

	// add aggregations
	return queryPart
}

func (f *NumericalField) getResultHistogramAggQuery(extrema *api.Extrema, resultVariable *model.Variable, numBuckets int) (string, string, string) {
	// compute the bucket interval for the histogram
	interval := extrema.GetBucketInterval(numBuckets)

	// Only numeric types should occur.
	fieldTyped := fmt.Sprintf("CAST(CASE WHEN \"%s\" = '' THEN 'NaN' ELSE \"%s\" END as double precision)", resultVariable.Key, resultVariable.Key)

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", api.HistogramAggPrefix, extrema.Key)
	rounded := extrema.GetBucketMinMax(numBuckets)

	bucketQueryString := ""
	// if only a single value, then return a simple count.
	if rounded.Max == rounded.Min {
		// want to return the count under bucket 0.
		bucketQueryString = fmt.Sprintf("(\"%s\" - \"%s\")", fieldTyped, fieldTyped)
	} else {
		bucketQueryString = fmt.Sprintf("width_bucket(%s, %g, %g, %d) - 1",
			fieldTyped, rounded.Min, rounded.Max, extrema.GetBucketCount(numBuckets))
	}
	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, rounded.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (f *NumericalField) fetchResultsExtrema(resultURI string, dataset string, resultVariable *model.Variable) (*api.Extrema, error) {
	// add min / max aggregation
	aggQuery := f.getResultMinMaxAggsQuery(resultVariable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE result_id = $1 AND target = $2 AND %s != '';", aggQuery, dataset, resultVariable.Key)

	// execute the postgres query
	res, err := f.Storage.client.Query(queryString, resultURI, f.Key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for result from postgres")
	}
	defer res.Close()

	return f.parseExtrema(res)
}

// FetchNumericalStats gets the variable's numerical summary info (mean, stddev).
func (f *NumericalField) FetchNumericalStats(filterParams *api.FilterParams) (*NumericalStats, error) {
	fromClause := f.getFromClause(true)

	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(f.GetDatasetName(), wheres, params, "", filterParams)
	wheres = append(wheres, f.getDefaultFilter(true))

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Create the complete query string.
	query := fmt.Sprintf("SELECT coalesce(stddev(\"%s\"), 0) as stddev, avg(\"%s\") as avg FROM %s %s;", f.Key, f.Key, fromClause, where)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch stats for variable from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseStats(res)
}

// FetchNumericalStatsByResult gets the variable's numerical summary info (mean, stddev) for a result set.
func (f *NumericalField) FetchNumericalStatsByResult(resultURI string, filterParams *api.FilterParams) (*NumericalStats, error) {
	fromClause := f.getFromClause(false)

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.GetDatasetName(), f.DatasetStorageName, resultURI, filterParams, baseTableAlias)
	if err != nil {
		return nil, err
	}
	wheres = append(wheres, f.getDefaultFilter(true))

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	// Create the complete query string.
	query := fmt.Sprintf("SELECT coalesce(stddev(\"%s\"), 0) as stddev, avg(\"%s\") as avg FROM %s INNER JOIN %s result ON %s.\"%s\" = result.index WHERE result.result_id = $%d %s;",
		f.Key, f.Key, fromClause, f.Storage.getResultTable(f.DatasetStorageName), baseTableAlias, model.D3MIndexFieldName, len(params), where)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch stats for variable from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseStats(res)
}

func (f *NumericalField) parseStats(row pgx.Rows) (*NumericalStats, error) {
	var stats *NumericalStats
	if row != nil {
		var stddev *float64
		var mean *float64
		// Expect one row of data.
		exists := row.Next()
		if !exists {
			return nil, fmt.Errorf("no result found")
		}
		err := row.Scan(&stddev, &mean)
		if err != nil {
			return nil, errors.Wrap(err, "no stats found")
		}

		stats = &NumericalStats{}

		if stddev != nil {
			stats.StdDev = *stddev
		}

		if mean != nil {
			stats.Mean = *mean
		}

		if mean == nil && stddev == nil {
			stats.NoDataAvailable = true
		}

	} else {
		return nil, errors.Errorf("no stats found")
	}

	return stats, nil
}

func (f *NumericalField) getFromClause(alias bool) string {
	fromClause := fmt.Sprintf("%s AS %s", f.DatasetStorageName, baseTableAlias)
	if f.subSelect != nil {
		fromClause = f.subSelect()
		if alias {
			fromClause = fmt.Sprintf("%s AS nested INNER JOIN %s AS %s on nested.\"%s\" = %s.\"%s\"",
				fromClause, f.DatasetStorageName, baseTableAlias, model.D3MIndexFieldName, baseTableAlias, model.D3MIndexFieldName)
		}
	}

	return fromClause
}

func (f *NumericalField) fetchExtremaStorage() (*api.Extrema, error) {
	aggQuery := f.getMinMaxAggsQuery()

	// numerical columns need to filter NaN out
	filter := fmt.Sprintf("WHERE \"%s\" != 'NaN'", f.Key)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s %s;", aggQuery, f.GetDatasetStorageName(), filter)

	// execute the postgres query
	// NOTE: We may want to use the regular Query operation since QueryRow
	// hides db exceptions.
	res, err := f.Storage.client.Query(queryString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseExtrema(res)
}
