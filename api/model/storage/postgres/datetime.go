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
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// DateTimeField defines behaviour for the numerical field type.
type DateTimeField struct {
	Storage     *Storage
	StorageName string
	Key         string
	Label       string
	Type        string
	subSelect   func() string
}

// NewDateTimeField creates a new field for numerical types.
func NewDateTimeField(storage *Storage, storageName string, key string, label string, typ string) *DateTimeField {
	field := &DateTimeField{
		Storage:     storage,
		StorageName: storageName,
		Key:         key,
		Label:       label,
		Type:        typ,
	}

	return field
}

// NewDateTimeFieldSubSelect creates a new field for numerical types
// and specifies a sub select query to pull the raw data.
func NewDateTimeFieldSubSelect(storage *Storage, storageName string, key string, label string, typ string, fieldSubSelect func() string) *DateTimeField {
	field := &DateTimeField{
		Storage:     storage,
		StorageName: storageName,
		Key:         key,
		Label:       label,
		Type:        typ,
		subSelect:   fieldSubSelect,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *DateTimeField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	var histogram *api.Histogram
	var err error
	if resultURI == "" {
		histogram, err = f.fetchHistogram(filterParams)
		if err != nil {
			return nil, err
		}
	} else {
		histogram, err = f.fetchHistogramByResult(resultURI, filterParams, extrema)
		if err != nil {
			return nil, err
		}
	}
	return histogram, nil
}

func (f *DateTimeField) fetchHistogram(filterParams *api.FilterParams) (*api.Histogram, error) {
	fromClause := f.getFromClause(true)

	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, filterParams.Filters)

	// need the extrema to calculate the histogram interval
	extrema, err := f.fetchExtrema()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable extrema for summary")
	}

	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := f.getHistogramAggQuery(extrema)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Create the complete query string.
	query := fmt.Sprintf("SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count FROM %s %s GROUP BY %s ORDER BY %s;",
		bucketQuery, histogramQuery, histogramName, fromClause, where, bucketQuery, histogramName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res, extrema)
}

func (f *DateTimeField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	fromClause := f.getFromClause(false)

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

	// need the extrema to calculate the histogram interval
	if extrema == nil {
		extrema, err = f.fetchExtremaByURI(resultURI)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch variable extrema for summary")
		}
	} else {
		extrema.Key = f.Key
		extrema.Type = f.Type
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := f.getHistogramAggQuery(extrema)

	// Create the complete query string.
	query := fmt.Sprintf(`
		SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count
		FROM %s data INNER JOIN %s result ON data."%s" = result.index
		WHERE result.result_id = $%d %s
		GROUP BY %s
		ORDER BY %s;`,
		bucketQuery, histogramQuery, histogramName, fromClause,
		f.Storage.getResultTable(f.StorageName), model.D3MIndexFieldName, len(params), where, bucketQuery, histogramName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res, extrema)
}

func (f *DateTimeField) fetchExtrema() (*api.Extrema, error) {
	fromClause := f.getFromClause(true)
	// add min / max aggregation
	aggQuery := f.getMinMaxAggsQuery()

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s;", aggQuery, fromClause)

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

func (f *DateTimeField) getHistogramAggQuery(extrema *api.Extrema) (string, string, string) {
	interval := extrema.GetBucketInterval()

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", api.HistogramAggPrefix, extrema.Key)
	rounded := extrema.GetBucketMinMax()

	bucketQueryString := ""
	// if only a single value, then return a simple count.
	if rounded.Max == rounded.Min {
		// want to return the count under bucket 0.
		bucketQueryString = fmt.Sprintf("(\"%s\" - \"%s\")", extrema.Key, extrema.Key)
	} else {
		bucketQueryString = fmt.Sprintf("width_bucket(cast(extract(epoch from \"%s\") as integer), %g, %g, %d) - 1",
			extrema.Key, rounded.Min, rounded.Max, extrema.GetBucketCount())
	}

	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, rounded.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (f *DateTimeField) parseValueToDateString(value string) (string, error) {
	ival, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return "", err
	}
	return time.Unix(ival, 0).Format(time.RFC3339), nil
}

func (f *DateTimeField) parseHistogram(rows *pgx.Rows, extrema *api.Extrema) (*api.Histogram, error) {
	// get histogram agg name
	histogramAggName := api.HistogramAggPrefix + extrema.Key

	// Parse bucket results.
	interval := extrema.GetBucketInterval()

	buckets := make([]*api.Bucket, extrema.GetBucketCount())
	rounded := extrema.GetBucketMinMax()
	key := rounded.Min
	for i := 0; i < len(buckets); i++ {
		keyString := ""
		if model.IsFloatingPoint(extrema.Type) {
			keyString = fmt.Sprintf("%f", key)
		} else {
			keyString = strconv.Itoa(int(key))
		}

		dateString, err := f.parseValueToDateString(keyString)
		if err != nil {
			return nil, err
		}

		buckets[i] = &api.Bucket{
			Key:   dateString,
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

	// assign histogram attributes
	return &api.Histogram{
		Label:   f.Label,
		Key:     f.Key,
		Type:    model.NumericalType,
		VarType: f.Type,
		Extrema: rounded,
		Buckets: buckets,
	}, nil
}

func (f *DateTimeField) parseExtrema(rows *pgx.Rows) (*api.Extrema, error) {
	var minValue *time.Time
	var maxValue *time.Time
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
		Key:  f.Key,
		Type: f.Type,
		Min:  float64(minValue.Unix()),
		Max:  float64(maxValue.Unix()),
	}, nil
}

func (f *DateTimeField) getMinMaxAggsQuery() string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + f.Key
	maxAggName := api.MaxAggPrefix + f.Key

	// create aggregations
	queryPart := fmt.Sprintf("MIN(\"%s\") AS \"%s\", MAX(\"%s\") AS \"%s\"",
		f.Key, minAggName, f.Key, maxAggName)
	// add aggregations
	return queryPart
}

func (f *DateTimeField) fetchExtremaByURI(resultURI string) (*api.Extrema, error) {
	fromClause := f.getFromClause(false)

	// add min / max aggregation
	aggQuery := f.getMinMaxAggsQuery()

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $1;",
		aggQuery, fromClause, f.Storage.getResultTable(f.StorageName), model.D3MIndexFieldName)

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
func (f *DateTimeField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	resultVariable := &model.Variable{
		Name: "value",
		Type: model.TextType,
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
	histogramName, bucketQuery, histogramQuery := f.getResultHistogramAggQuery(extrema, resultVariable)

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.StorageName, resultURI, filterParams)
	if err != nil {
		return nil, err
	}

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, f.Key)

	// Create the complete query string.
	query := fmt.Sprintf(`
		SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count
		FROM %s data INNER JOIN %s result ON data."%s" = result.index
		WHERE %s
		GROUP BY %s
		ORDER BY %s;`,
		bucketQuery, histogramQuery, histogramName, f.StorageName, datasetResult,
		model.D3MIndexFieldName, strings.Join(wheres, " AND "), bucketQuery, histogramName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result variable summaries from postgres")
	}
	defer res.Close()

	return f.parseHistogram(res, extrema)
}

func (f *DateTimeField) getResultMinMaxAggsQuery(resultVariable *model.Variable) string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + resultVariable.Name
	maxAggName := api.MaxAggPrefix + resultVariable.Name

	// Only numeric types should occur.
	fieldTyped := fmt.Sprintf("cast(\"%s\" as double precision)", resultVariable.Name)

	// create aggregations
	queryPart := fmt.Sprintf("MIN(%s) AS \"%s\", MAX(%s) AS \"%s\"", fieldTyped, minAggName, fieldTyped, maxAggName)
	// add aggregations
	return queryPart
}

func (f *DateTimeField) getResultHistogramAggQuery(extrema *api.Extrema, resultVariable *model.Variable) (string, string, string) {
	// compute the bucket interval for the histogram
	interval := extrema.GetBucketInterval()

	// Only numeric types should occur.
	fieldTyped := fmt.Sprintf("cast(\"%s\" as double precision)", resultVariable.Name)

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", api.HistogramAggPrefix, extrema.Key)
	rounded := extrema.GetBucketMinMax()

	bucketQueryString := ""
	// if only a single value, then return a simple count.
	if rounded.Max == rounded.Min {
		// want to return the count under bucket 0.
		bucketQueryString = fmt.Sprintf("(\"%s\" - \"%s\")", fieldTyped, fieldTyped)
	} else {
		bucketQueryString = fmt.Sprintf("width_bucket(%s, %g, %g, %d) - 1",
			fieldTyped, rounded.Min, rounded.Max, extrema.GetBucketCount())
	}
	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, rounded.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (f *DateTimeField) fetchResultsExtrema(resultURI string, dataset string, resultVariable *model.Variable) (*api.Extrema, error) {
	// add min / max aggregation
	aggQuery := f.getResultMinMaxAggsQuery(resultVariable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE result_id = $1 AND target = $2;", aggQuery, dataset)

	// execute the postgres query
	res, err := f.Storage.client.Query(queryString, resultURI, f.Key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for result from postgres")
	}
	defer res.Close()

	return f.parseExtrema(res)
}

func (f *DateTimeField) getFromClause(alias bool) string {
	fromClause := f.StorageName
	if f.subSelect != nil {
		fromClause = f.subSelect()
		if alias {
			fromClause = fmt.Sprintf("%s as nested INNER JOIN %s as data on nested.\"%s\" = data.\"%s\"", fromClause, f.StorageName, model.D3MIndexFieldName, model.D3MIndexFieldName)
			//fromClause = fmt.Sprintf("%s as %s", fromClause, f.Dataset)
		}
	}

	return fromClause
}
