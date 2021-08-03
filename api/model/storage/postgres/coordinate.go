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
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/postgres"
)

const coordinateBuckets = 20

// CoordinateField defines behaviour for the coordinate field type.
type CoordinateField struct {
	BasicField
	XCol string
	YCol string
}

// NewCoordinateField creates a new field for coordinate types.
func NewCoordinateField(key string, storage *Storage, datasetName string, datasetStorageName string, xCol string, yCol string, label string, typ string, count string) *CoordinateField {
	count = getCountSQL(count)

	field := &CoordinateField{
		BasicField: BasicField{
			Key:                key,
			Storage:            storage,
			DatasetName:        datasetName,
			DatasetStorageName: datasetStorageName,
			Label:              label,
			Type:               typ,
			Count:              count,
		},
		XCol: xCol,
		YCol: yCol,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *CoordinateField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	if resultURI == "" {
		baseline, err = f.fetchHistogram(nil, coordinateBuckets)
		if err != nil {
			return nil, err
		}
		if !filterParams.IsEmpty(true) {
			filtered, err = f.fetchHistogram(filterParams, coordinateBuckets)
			if err != nil {
				return nil, err
			}
		}
	} else {
		baseline, err = f.fetchHistogramByResult(resultURI, nil, coordinateBuckets)
		if err != nil {
			return nil, err
		}
		if !filterParams.IsEmpty(true) {
			filtered, err = f.fetchHistogramByResult(resultURI, filterParams, coordinateBuckets)
			if err != nil {
				return nil, err
			}
		}
	}

	return &api.VariableSummary{
		Key:      f.Key,
		Label:    f.Label,
		Type:     model.GeoCoordinateType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
		Timeline: nil,
	}, nil
}

func (f *CoordinateField) fetchHistogram(filterParams *api.FilterParams, numBuckets int) (*api.Histogram, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(f.GetDatasetName(), wheres, params, "", filterParams)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	// treat each axis as a separate field for the purposes of query generation
	xField := NewNumericalField(f.Storage, f.DatasetName, f.DatasetStorageName, f.XCol, f.XCol, model.RealType, "")
	yField := NewNumericalField(f.Storage, f.DatasetName, f.DatasetStorageName, f.YCol, f.YCol, model.RealType, "")

	// get the extrema for each axis
	xExtrema, err := xField.fetchExtrema()
	if err != nil {
		return nil, err
	}
	yExtrema, err := yField.fetchExtrema()
	if err != nil {
		return nil, err
	}

	xNumBuckets, yNumBuckets := getEqualBivariateBuckets(numBuckets, xExtrema, yExtrema)

	// generate a histogram query for each
	xHistogramName, xBucketQuery, xHistogramQuery := xField.getHistogramAggQuery(xExtrema, xNumBuckets, "")
	yHistogramName, yBucketQuery, yHistogramQuery := yField.getHistogramAggQuery(yExtrema, yNumBuckets, "")

	// Get count by x & y
	query := fmt.Sprintf(`SELECT %s as bucket, CAST(%s as double precision) AS %s, %s as bucket, CAST(%s as double precision) AS %s, COUNT(%s) AS count
        FROM %s
        WHERE "%s" != 'NaN' AND "%s" != 'NaN' %s
        GROUP BY %s, %s
        ORDER BY %s, %s;`,
		xBucketQuery, xHistogramQuery, xHistogramName, yBucketQuery, yHistogramQuery, yHistogramName, f.Count,
		f.DatasetStorageName, f.XCol, f.YCol, where, xBucketQuery, yBucketQuery, xHistogramName, yHistogramName)

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

func (f *CoordinateField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams, numBuckets int) (*api.Histogram, error) {

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

	// create a numerical field for each of X and Y
	xField := NewNumericalField(f.Storage, f.DatasetName, f.DatasetStorageName, f.XCol, f.XCol, model.RealType, "")
	yField := NewNumericalField(f.Storage, f.DatasetName, f.DatasetStorageName, f.YCol, f.YCol, model.RealType, "")

	// get the extrema for each
	xExtrema, err := xField.fetchExtrema()
	if err != nil {
		return nil, err
	}
	yExtrema, err := yField.fetchExtrema()
	if err != nil {
		return nil, err
	}

	xNumBuckets, yNumBuckets := getEqualBivariateBuckets(numBuckets, xExtrema, yExtrema)

	// create histograms given the the extrema
	xHistogramName, xBucketQuery, xHistogramQuery := xField.getHistogramAggQuery(xExtrema, xNumBuckets, baseTableAlias)
	yHistogramName, yBucketQuery, yHistogramQuery := yField.getHistogramAggQuery(yExtrema, yNumBuckets, baseTableAlias)

	// Get count by x & y
	query := fmt.Sprintf(`
		SELECT %s as bucket, CAST(%s as double precision) AS %s, %s as bucket, CAST(%s as double precision) AS %s, COUNT(%s) AS count
		FROM %s data INNER JOIN %s result ON data."%s" = result.index
		WHERE result.result_id = $%d AND "%s" != 'NaN' AND "%s" != 'NaN' %s
		GROUP BY %s, %s
		ORDER BY %s, %s;`,
		xBucketQuery, xHistogramQuery, xHistogramName, yBucketQuery, yHistogramQuery, yHistogramName, f.Count,
		f.DatasetStorageName, f.Storage.getResultTable(f.DatasetStorageName), model.D3MIndexFieldName,
		len(params), f.XCol, f.YCol, where, xBucketQuery, yBucketQuery, xHistogramName, yHistogramName)

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

func (f *CoordinateField) parseHistogram(rows pgx.Rows, xExtrema *api.Extrema, yExtrema *api.Extrema, xNumBuckets int, yNumBuckets int) (*api.Histogram, error) {
	// get histogram agg name
	histogramAggName := api.HistogramAggPrefix + f.Key

	// Parse bucket results.
	xInterval := xExtrema.GetBucketInterval(xNumBuckets)
	yInterval := yExtrema.GetBucketInterval(yNumBuckets)
	xRounded := xExtrema.GetBucketMinMax(xNumBuckets)
	yRounded := yExtrema.GetBucketMinMax(yNumBuckets)

	xBucketCount := int64(xExtrema.GetBucketCount(xNumBuckets))
	yBucketCount := int64(yExtrema.GetBucketCount(yNumBuckets))

	// initialize empty histogram structure
	// float representation of the data could cause very slight deviation
	// leading to boundaries not being perfectly aligned
	xBuckets := make([]*api.Bucket, xBucketCount)
	for i := int64(0); i < xBucketCount; i++ {
		yBuckets := make([]*api.Bucket, yBucketCount)
		for j := int64(0); j < yBucketCount; j++ {
			yVal := yRounded.Min + float64(j)*yInterval
			yBuckets[j] = &api.Bucket{
				Key:     fmt.Sprintf("%f", yVal),
				Count:   0,
				Buckets: nil,
			}
		}
		xVal := xRounded.Min + float64(i)*xInterval
		xBuckets[i] = &api.Bucket{
			Key:     fmt.Sprintf("%f", xVal),
			Count:   0,
			Buckets: yBuckets,
		}
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
func (f *CoordinateField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *CoordinateField) getDefaultFilter(inverse bool) string {
	defaultValue := fmt.Sprintf("%v", postgres.DefaultPostgresValueFromD3MType(f.GetType()))

	// null comparison is done with "is" rather than "="
	op := "="
	if defaultValue == "NULL" {
		op = "is"
	}
	negate := ""
	if inverse {
		negate = "NOT"
	}

	return fmt.Sprintf("%s(\"%s\" %s %s AND \"%s\" %s %s)", negate, f.XCol, op, defaultValue, f.YCol, op, defaultValue)
}