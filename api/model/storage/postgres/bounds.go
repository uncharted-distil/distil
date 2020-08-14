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
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// BoundsField defines behaviour for the remote sensing field type.
type BoundsField struct {
	BasicField
}

// NewBoundsField creates a new field for remote sensing types.
func NewBoundsField(storage *Storage, datasetName string, datasetStorageName string, key string, label string, typ string, count string) *BoundsField {
	count = getCountSQL(count)

	field := &BoundsField{
		BasicField: BasicField{
			Key:                key,
			Storage:            storage,
			DatasetName:        datasetName,
			DatasetStorageName: datasetStorageName,
			Label:              label,
			Type:               typ,
			Count:              count,
		},
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *BoundsField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool, mode api.SummaryMode) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	// update the highlight key to use the cluster if necessary
	if err = f.updateClusterHighlight(filterParams, mode); err != nil {
		return nil, err
	}

	if resultURI == "" {
		baseline, err = f.fetchHistogram(nil, invert, coordinateBuckets)
		if err != nil {
			return nil, err
		}
		if !filterParams.Empty() {
			filtered, err = f.fetchHistogram(filterParams, invert, coordinateBuckets)
			if err != nil {
				return nil, err
			}
		}
	} else {
		baseline, err = f.fetchHistogramByResult(resultURI, nil, coordinateBuckets)
		if err != nil {
			return nil, err
		}
		if !filterParams.Empty() {
			filtered, err = f.fetchHistogramByResult(resultURI, filterParams, coordinateBuckets)
			if err != nil {
				return nil, err
			}
		}
	}

	return &api.VariableSummary{
		Key:      f.Key,
		Label:    f.Label,
		Type:     model.RemoteSensingType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
		Timeline: nil,
	}, nil
}

func (f *BoundsField) fetchHistogram(filterParams *api.FilterParams, invert bool, numBuckets int) (*api.Histogram, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(f.GetDatasetName(), wheres, params, "", filterParams, invert)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// get the extrema for each axis
	xExtrema, yExtrema, err := f.fetchExtrema()
	if err != nil {
		return nil, err
	}

	xNumBuckets, yNumBuckets := getEqualBivariateBuckets(numBuckets, xExtrema, yExtrema)

	// generate a histogram query for each
	xMinHistogramName, xMinBucketQuery, intervalX, minX := f.getHistogramAggQuery(xExtrema, xNumBuckets, 1)
	_, xMaxBucketQuery, _, _ := f.getHistogramAggQuery(xExtrema, xNumBuckets, 5)
	yMinHistogramName, yMinBucketQuery, intervalY, minY := f.getHistogramAggQuery(yExtrema, yNumBuckets, 2)
	_, yMaxBucketQuery, _, _ := f.getHistogramAggQuery(yExtrema, yNumBuckets, 6)

	// Get count by x & y
	query := fmt.Sprintf(`
        SELECT
          xbuckets, CAST(xbuckets * %g + %g AS double precision) AS %s,
          ybuckets, CAST(ybuckets * %g + %g AS double precision) AS %s, COUNT(%s) FROM
        (
          SELECT %s,
          %s AS minx, %s AS maxx, %s AS miny, %s AS maxy
          FROM %s %s
        ) AS points
        INNER JOIN generate_series(0, %d) AS xbuckets ON xbuckets >= minx AND xbuckets <= maxx
        INNER JOIN generate_series(0, %d) AS ybuckets ON ybuckets >= miny AND ybuckets <= maxy
        GROUP BY xbuckets, ybuckets
        ORDER BY xbuckets, ybuckets;`,
		intervalX, minX, xMinHistogramName, intervalY, minY, yMinHistogramName,
		f.Count, f.Count, xMinBucketQuery, xMaxBucketQuery, yMinBucketQuery, yMaxBucketQuery,
		f.DatasetStorageName, where, xNumBuckets, yNumBuckets)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res, xExtrema, yExtrema, xNumBuckets, yNumBuckets)
	if err != nil {
		return nil, err
	}

	return histogram, nil
}

func (f *BoundsField) fetchExtrema() (*api.Extrema, *api.Extrema, error) {
	// add min / max aggregation
	aggQueryX := f.getMinMaxAggsQuery(f.Key, "x", 1, 5)
	aggQueryY := f.getMinMaxAggsQuery(f.Key, "y", 2, 6)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s, %s FROM %s;", aggQueryX, aggQueryY, f.DatasetStorageName)

	// execute the postgres query
	res, err := f.Storage.client.Query(queryString)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseExtrema(res)
}

func (f *BoundsField) getHistogramAggQuery(extrema *api.Extrema, numBuckets int, index int) (string, string, float64, float64) {
	interval := extrema.GetBucketInterval(numBuckets)

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", api.HistogramAggPrefix, extrema.Key)
	rounded := extrema.GetBucketMinMax(numBuckets)

	bucketQueryString := ""
	// if only a single value, then return a simple count.
	if rounded.Max == rounded.Min {
		// want to return the count under bucket 0.
		bucketQueryString = fmt.Sprintf("(\"%s\" - \"%s\")", extrema.Key, extrema.Key)
	} else {
		bucketQueryString = fmt.Sprintf("width_bucket(\"%s\"[%d], %g, %g, %d) - 1",
			extrema.Key, index, rounded.Min, rounded.Max, extrema.GetBucketCount(numBuckets))
	}

	return histogramAggName, bucketQueryString, interval, rounded.Min
}

func (f *BoundsField) getMinMaxAggsQuery(key string, label string, indexMin int, indexMax int) string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + label
	maxAggName := api.MaxAggPrefix + label

	// create aggregations
	queryPart := fmt.Sprintf("MIN(\"%s\"[%d]) AS \"%s\", MAX(\"%s\"[%d]) AS \"%s\"",
		key, indexMin, minAggName, key, indexMax, maxAggName)

	return queryPart
}

func (f *BoundsField) parseExtrema(rows pgx.Rows) (*api.Extrema, *api.Extrema, error) {
	var minXValue *float64
	var maxXValue *float64
	var minYValue *float64
	var maxYValue *float64
	if rows != nil {
		// Expect one row of data.
		exists := rows.Next()
		if !exists {
			return nil, nil, fmt.Errorf("no rows in extrema query result")
		}
		err := rows.Scan(&minXValue, &maxXValue, &minYValue, &maxYValue)
		if err != nil {
			return nil, nil, errors.Wrap(err, "no min / max aggregation found")
		}
	}
	// check values exist
	if minXValue == nil || maxXValue == nil || minYValue == nil || maxYValue == nil {
		return nil, nil, errors.Errorf("no min / max aggregation values found")
	}
	// assign attributes
	xExtrema := &api.Extrema{
		Key:  f.Key,
		Type: model.LongitudeType,
		Min:  *minXValue,
		Max:  *maxXValue,
	}
	yExtrema := &api.Extrema{
		Key:  f.Key,
		Type: model.LatitudeType,
		Min:  *minYValue,
		Max:  *maxYValue,
	}

	return xExtrema, yExtrema, nil
}

func (f *BoundsField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams, numBuckets int) (*api.Histogram, error) {
	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.GetDatasetName(), f.DatasetStorageName, resultURI, filterParams, baseTableAlias)
	if err != nil {
		return nil, err
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	// get the extrema for each axis
	xExtrema, yExtrema, err := f.fetchExtrema()
	if err != nil {
		return nil, err
	}

	xNumBuckets, yNumBuckets := getEqualBivariateBuckets(numBuckets, xExtrema, yExtrema)

	// generate a histogram query for each
	xMinHistogramName, xMinBucketQuery, intervalX, minX := f.getHistogramAggQuery(xExtrema, xNumBuckets, 1)
	_, xMaxBucketQuery, _, _ := f.getHistogramAggQuery(xExtrema, xNumBuckets, 5)
	yMinHistogramName, yMinBucketQuery, intervalY, minY := f.getHistogramAggQuery(yExtrema, yNumBuckets, 2)
	_, yMaxBucketQuery, _, _ := f.getHistogramAggQuery(yExtrema, yNumBuckets, 6)

	// Get count by x & y
	query := fmt.Sprintf(`
        SELECT
          xbuckets, CAST(xbuckets * %g + %g AS double precision) AS %s,
          ybuckets, CAST(ybuckets * %g + %g AS double precision) AS %s, COUNT(%s) FROM
        (
          SELECT %s,
          %s AS minx, %s AS maxx, %s AS miny, %s AS maxy
          FROM %s data INNER JOIN %s result ON data."%s" = result.index
					WHERE result.result_id = $%d %s
        ) AS points
        INNER JOIN generate_series(0, %d) AS xbuckets ON xbuckets >= minx AND xbuckets <= maxx
        INNER JOIN generate_series(0, %d) AS ybuckets ON ybuckets >= miny AND ybuckets <= maxy
        GROUP BY xbuckets, ybuckets
        ORDER BY xbuckets, ybuckets;`,
		intervalX, minX, xMinHistogramName, intervalY, minY, yMinHistogramName,
		f.Count, f.Count, xMinBucketQuery, xMaxBucketQuery, yMinBucketQuery, yMaxBucketQuery,
		f.DatasetStorageName, f.Storage.getResultTable(f.DatasetStorageName), model.D3MIndexFieldName,
		len(params), where, xNumBuckets, yNumBuckets)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res, xExtrema, yExtrema, xNumBuckets, yNumBuckets)
	if err != nil {
		return nil, err
	}

	return histogram, nil
}

func (f *BoundsField) parseHistogram(rows pgx.Rows, xExtrema *api.Extrema, yExtrema *api.Extrema, xNumBuckets int, yNumBuckets int) (*api.Histogram, error) {
	// get histogram agg name
	histogramAggName := api.HistogramAggPrefix + f.Key

	// Parse bucket results.
	xInterval := xExtrema.GetBucketInterval(xNumBuckets)
	yInterval := yExtrema.GetBucketInterval(yNumBuckets)
	xRounded := xExtrema.GetBucketMinMax(xNumBuckets)
	yRounded := yExtrema.GetBucketMinMax(yNumBuckets)

	xBucketCount := int64(xExtrema.GetBucketCount(xNumBuckets))
	yBucketCount := int64(yExtrema.GetBucketCount(yNumBuckets))
	xBuckets := make([]*api.Bucket, xBucketCount)

	// initialize empty histogram structure
	i := 0
	for xVal := xRounded.Min; xVal < xRounded.Max; xVal += xInterval {
		yBuckets := make([]*api.Bucket, yBucketCount)
		j := 0
		for yVal := yRounded.Min; yVal < yRounded.Max; yVal += yInterval {
			yBuckets[j] = &api.Bucket{
				Key:     fmt.Sprintf("%f", yVal),
				Count:   0,
				Buckets: nil,
			}
			j++
		}
		xBuckets[i] = &api.Bucket{
			Key:     fmt.Sprintf("%f", xVal),
			Count:   0,
			Buckets: yBuckets,
		}
		i++
	}

	for rows.Next() {
		var xBucketValue float64
		var yBucketValue float64
		var xBucket int64
		var yBucket int64
		var yRowBucketCount int64
		err := rows.Scan(&xBucket, &xBucketValue, &yBucket, &yBucketValue, &yRowBucketCount)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", histogramAggName))
		}

		// Due to float representation, sometimes the lowest value <
		// first bucket interval and so ends up in bucket -1.
		// Since the max can match the limit, an extra bucket may exist.
		// Add the value to the second to last bucket.
		if xBucket < 0 {
			xBucket = 0
		} else if xBucket >= xBucketCount {
			xBucket = xBucketCount - 1
		}
		if yBucket < 0 {
			yBucket = 0
		} else if yBucket >= yBucketCount {
			yBucket = yBucketCount - 1
		}
		xBuckets[xBucket].Buckets[yBucket].Count += yRowBucketCount
	}
	err := rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	// assign histogram attributes
	return &api.Histogram{
		Buckets: xBuckets,
	}, nil
}

// FetchPredictedSummaryData pulls predicted data from the result table and builds
// the coordinate histogram for the field.
func (f *BoundsField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	return nil, fmt.Errorf("not implemented")
}
