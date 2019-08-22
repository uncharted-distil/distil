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
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// CoordinateField defines behaviour for the coordinate field type.
type CoordinateField struct {
	Key         string
	Storage     *Storage
	StorageName string
	XCol        string
	YCol        string
	Label       string
	Type        string
}

// NewCoordinateField creates a new field for coordinate types.
func NewCoordinateField(key string, storage *Storage, storageName string, xCol string, yCol string, label string, typ string) *CoordinateField {
	field := &CoordinateField{
		Key:         key,
		Storage:     storage,
		StorageName: storageName,
		XCol:        xCol,
		YCol:        yCol,
		Label:       label,
		Type:        typ,
	}

	return field
}

// FetchTimeseriesSummaryData pulls summary data from the database and builds a histogram.
func (f *CoordinateField) FetchTimeseriesSummaryData(timeVar *model.Variable, interval int, resultURI string, filterParams *api.FilterParams, invert bool) (*api.VariableSummary, error) {
	return nil, fmt.Errorf("not implemented")
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *CoordinateField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool) (*api.VariableSummary, error) {
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
		Key:      f.Key,
		Label:    f.Label,
		Type:     model.GeoCoordinateType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
		Timeline: nil,
	}, nil
}

func (f *CoordinateField) fetchExtrema(fieldName string, filterParams *api.FilterParams) *api.Extrema {
	// cycle through the filters and find the one for the field
	var filter *model.Filter
	if filterParams != nil {
		for _, p := range filterParams.Filters {
			if p.Key == fieldName {
				filter = p
				break
			}
		}
	}
	if filter == nil {
		filter = f.fetchDefaultFilter(fieldName)
	}

	// use the filter to build the extrema
	return &api.Extrema{
		Key:  fieldName,
		Type: model.RealType,
		Min:  *filter.Min,
		Max:  *filter.Max,
	}
}

func (f *CoordinateField) fetchHistogram(filterParams *api.FilterParams, invert bool) (*api.Histogram, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, filterParams, false)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	xField := NewNumericalField(f.Storage, f.StorageName, f.XCol, f.XCol, model.RealType)
	yField := NewNumericalField(f.Storage, f.StorageName, f.YCol, f.YCol, model.RealType)
	xExtrema := f.fetchExtrema(f.XCol, filterParams)
	yExtrema := f.fetchExtrema(f.YCol, filterParams)

	xHistogramName, xBucketQuery, xHistogramQuery := xField.getHistogramAggQuery(xExtrema)
	yHistogramName, yBucketQuery, yHistogramQuery := yField.getHistogramAggQuery(yExtrema)

	// Get count by x & y
	query := fmt.Sprintf(`SELECT %s as bucket, CAST(%s as double precision) AS %s, %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count
        FROM %s %s
        GROUP BY %s, %s
        ORDER BY %s, %s;`,
		xBucketQuery, xHistogramQuery, xHistogramName, yBucketQuery, yHistogramQuery, yHistogramName,
		f.StorageName, where, xBucketQuery, yBucketQuery, xHistogramName, yHistogramName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res, xExtrema, yExtrema)
	if err != nil {
		return nil, err
	}

	return histogram, nil
}

func (f *CoordinateField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams) (*api.Histogram, error) {

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

	xField := NewNumericalField(f.Storage, f.StorageName, f.XCol, f.XCol, model.RealType)
	yField := NewNumericalField(f.Storage, f.StorageName, f.YCol, f.YCol, model.RealType)
	xExtrema := f.fetchExtrema(f.XCol, filterParams)
	yExtrema := f.fetchExtrema(f.YCol, filterParams)

	xHistogramName, xBucketQuery, xHistogramQuery := xField.getHistogramAggQuery(xExtrema)
	yHistogramName, yBucketQuery, yHistogramQuery := yField.getHistogramAggQuery(yExtrema)

	// Get count by x & y
	query := fmt.Sprintf(`
		SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count, %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count
		FROM %s data INNER JOIN %s result ON data."%s" = result.index
		WHERE result.result_id = $%d %s
		GROUP BY %s, %s
		ORDER BY %s, %s;`,
		xBucketQuery, xHistogramQuery, xHistogramName, yBucketQuery, yHistogramQuery, yHistogramName,
		f.StorageName, f.Storage.getResultTable(f.StorageName), model.D3MIndexFieldName,
		len(params), where, xBucketQuery, yBucketQuery, xHistogramName, yHistogramName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res, xExtrema, yExtrema)
	if err != nil {
		return nil, err
	}

	return histogram, nil
}

func (f *CoordinateField) parseHistogram(rows *pgx.Rows, xExtrema *api.Extrema, yExtrema *api.Extrema) (*api.Histogram, error) {
	// get histogram agg name
	histogramAggName := api.HistogramAggPrefix + f.Key

	// Parse bucket results.
	xInterval := xExtrema.GetBucketInterval()
	yInterval := yExtrema.GetBucketInterval()
	xRounded := xExtrema.GetBucketMinMax()
	yRounded := yExtrema.GetBucketMinMax()

	xBucketCount := int64(xExtrema.GetBucketCount())
	yBucketCount := int64(yExtrema.GetBucketCount())
	buckets := make([]*api.Bucket, xBucketCount*yBucketCount)
	i := 0
	for xVal := xRounded.Min; xVal < xRounded.Max; xVal += xInterval {
		for yVal := yRounded.Min; yVal < yRounded.Max; yVal += yInterval {
			buckets[i] = &api.Bucket{
				Key:   fmt.Sprintf("%f,%f", xVal, yVal),
				Count: 0,
			}
			i++
		}
	}

	for rows.Next() {
		var xBucketValue float64
		var yBucketValue float64
		var xBucket int64
		var yBucket int64
		var bucketCount int64
		err := rows.Scan(&xBucket, &xBucketValue, &yBucket, &yBucketValue, &bucketCount)
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
		buckets[xBucket*yBucketCount+yBucket].Count += bucketCount
	}

	// assign histogram attributes
	return &api.Histogram{
		Buckets: buckets,
	}, nil
}

// FetchPredictedSummaryData pulls predicted data from the result table and builds
// the coordinate histogram for the field.
func (f *CoordinateField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.VariableSummary, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *CoordinateField) fetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	return nil, fmt.Errorf("not implemented")
}

// FetchForecastingSummaryData pulls data from the result table and builds the
// forecasting histogram for the field.
func (f *CoordinateField) FetchForecastingSummaryData(timeVar *model.Variable, interval int, resultURI string, filterParams *api.FilterParams) (*api.VariableSummary, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *CoordinateField) fetchDefaultFilter(fieldName string) *model.Filter {
	// provide a useful default based on type
	// geo can default to lat and lon max values
	min := -float64(math.MaxInt64)
	max := float64(math.MaxInt64)
	if model.IsGeoCoordinate(f.Type) {
		if fieldName == f.XCol {
			min = float64(-180)
			max = float64(180)
		} else if fieldName == f.YCol {
			min = float64(-90)
			max = float64(90)
		}
	}

	return &model.Filter{
		Min: &min,
		Max: &max,
	}
}
