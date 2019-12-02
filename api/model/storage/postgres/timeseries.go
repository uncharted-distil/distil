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
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// TimeSeriesField defines behaviour for the timeseries field type.
type TimeSeriesField struct {
	BasicField
	ClusterCol string
	XCol       string
	XColType   string
	YCol       string
	YColType   string
}

// NewTimeSeriesField creates a new field for timeseries types.
func NewTimeSeriesField(storage *Storage, storageName string, clusterCol string, key string, label string, typ string,
	xCol string, xColType string, yCol string, yColType string) *TimeSeriesField {
	field := &TimeSeriesField{
		BasicField: BasicField{
			Storage:     storage,
			StorageName: storageName,
			Label:       label,
			Type:        typ,
			Key:         key,
		},
		XCol:       xCol,
		XColType:   xColType,
		YCol:       yCol,
		YColType:   yColType,
		ClusterCol: clusterCol,
	}

	return field
}

func (s *Storage) parseTimeseries(rows *pgx.Rows) ([][]float64, error) {
	var points [][]float64
	if rows != nil {
		for rows.Next() {
			var x float64
			var y float64
			err := rows.Scan(&x, &y)
			if err != nil {
				return nil, err
			}
			points = append(points, []float64{x, y})
		}
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i][0] < points[j][0]
	})

	return points, nil
}

func (f *TimeSeriesField) fetchRepresentationTimeSeries(categoryBuckets []*api.Bucket) ([]string, error) {

	var timeseriesExemplars []string

	for _, bucket := range categoryBuckets {

		keyColName := f.keyColName()

		// pull sample row containing bucket
		query := fmt.Sprintf("SELECT \"%s\" FROM %s WHERE \"%s\" = $1 LIMIT 1;",
			f.Key, f.StorageName, keyColName)

		// execute the postgres query
		rows, err := f.Storage.client.Query(query, bucket.Key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
		}

		if rows.Next() {

			values, err := rows.Values()
			if err != nil {
				rows.Close()
				return nil, errors.Wrap(err, "unable to parse solution from Postgres")
			}

			if len(values) < 1 {
				return nil, errors.Wrap(fmt.Errorf("missing values"), "unable to parse timeseries id")
			}

			timeseriesExemplarInt, ok := values[0].(int)
			if ok {
				timeseriesExemplars = append(timeseriesExemplars, strconv.FormatInt(int64(timeseriesExemplarInt), 10))
				rows.Close()
				continue
			}

			timeseriesExemplar, ok := values[0].(string)
			if ok {
				timeseriesExemplars = append(timeseriesExemplars, timeseriesExemplar)
				rows.Close()
				continue
			}

			rows.Close()
			return nil, errors.Wrap(fmt.Errorf("timeseries id type not recognized %v", values[0]), "unable to parse timeseries id")
		}
	}

	if len(timeseriesExemplars) == 0 {
		return nil, errors.New("No exemplars found for timeseries data")
	}

	return timeseriesExemplars, nil
}

// FetchTimeseries fetches a timeseries.
func (s *Storage) FetchTimeseries(dataset string, storageName string, timeseriesColName string, xColName string, yColName string, timeseriesURI string, filterParams *api.FilterParams, invert bool) ([][]float64, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	wheres = append(wheres, fmt.Sprintf("\"%s\" = $1", timeseriesColName))
	params = append(params, timeseriesURI)

	wheres, params = s.buildFilteredQueryWhere(wheres, params, "", filterParams, invert)
	where := fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))

	// Get count by category.
	query := fmt.Sprintf("SELECT \"%s\", \"%s\" FROM %s %s",
		xColName, yColName, storageName, where)

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch timeseries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseTimeseries(res)
}

// FetchTimeseriesForecast fetches a timeseries.
func (s *Storage) FetchTimeseriesForecast(dataset string, storageName string, timeseriesColName string, xColName string, yColName string, timeseriesURI string, resultURI string, filterParams *api.FilterParams) ([][]float64, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	wheres = append(wheres, fmt.Sprintf("\"%s\" = $1", timeseriesColName))
	params = append(params, timeseriesURI)

	wheres, params = s.buildFilteredQueryWhere(wheres, params, "", filterParams, false)

	params = append(params, resultURI)
	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d", len(params)))

	where := fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))

	// Get count by category.
	query := fmt.Sprintf(`SELECT "%s", CAST(result.value as double precision)
		FROM %s data INNER JOIN %s result ON data."%s" = result.index
		%s`,
		xColName, storageName, s.getResultTable(storageName),
		model.D3MIndexFieldName, where)

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch timeseries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseTimeseries(res)
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *TimeSeriesField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var timeline *api.Histogram
	var err error

	// update the highlight key to use the cluster if necessary
	if err = f.updateClusterHighlight(filterParams); err != nil {
		return nil, err
	}

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

	timelineField := NewNumericalField(f.Storage, f.StorageName, f.XCol, f.XCol, f.XColType)

	timeline, err = timelineField.fetchHistogram(nil, invert, api.MaxNumBuckets)
	if err != nil {
		return nil, err
	}

	return &api.VariableSummary{
		Key:      f.Key,
		Label:    f.Label,
		Type:     model.CategoricalType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
		Timeline: timeline,
	}, nil
}

func (f *TimeSeriesField) keyColName() string {
	if f.hasClusterData(f.GetKey()) {
		if f.ClusterCol != "" {
			return fmt.Sprintf("%s%s", model.ClusterVarPrefix, f.ClusterCol)
		}
		return f.Key
	}
	return f.Key
}

func (f *TimeSeriesField) fetchHistogram(filterParams *api.FilterParams, invert bool) (*api.Histogram, error) {

	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, "", filterParams, false)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	colName := f.GetKey()
	if f.hasClusterData(colName) {
		colName = f.keyColName()
	}
	query := fmt.Sprintf("SELECT \"%s\", COUNT(*) AS __count__ FROM %s %s GROUP BY \"%s\" ORDER BY __count__ desc, \"%s\" LIMIT %d;",
		colName, f.StorageName, where, colName, colName, timeSeriesCatResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationTimeSeries(histogram.Buckets)
	if err != nil {
		return nil, err
	}
	histogram.Exemplars = files
	return histogram, nil
}

func (f *TimeSeriesField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams) (*api.Histogram, error) {

	wheres := []string{}
	params := []interface{}{}
	var err error
	if f.Type != "timeseries" {
		// get filter where / params
		wheres, params, err = f.Storage.buildResultQueryFilters(f.StorageName, resultURI, filterParams)
		if err != nil {
			return nil, err
		}
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	keyColName := f.keyColName()

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(*) AS __count__
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d %s
		 GROUP BY "%s"
		 ORDER BY __count__ desc, "%s" LIMIT %d;`,
		keyColName, f.StorageName, f.Storage.getResultTable(f.StorageName),
		model.D3MIndexFieldName, len(params), where, keyColName,
		keyColName, timeSeriesCatResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationTimeSeries(histogram.Buckets)
	if err != nil {
		return nil, err
	}
	histogram.Exemplars = files
	return histogram, nil
}

func (f *TimeSeriesField) parseHistogram(rows *pgx.Rows) (*api.Histogram, error) {
	keyColName := f.keyColName()

	termsAggName := api.TermsAggPrefix + keyColName

	// Parse bucket results.
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

// FetchPredictedSummaryData pulls predicted data from the result table and builds
// the timeseries histogram for the field.
func (f *TimeSeriesField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	// update the highlight key to use the cluster if necessary
	if err = f.updateClusterHighlight(filterParams); err != nil {
		return nil, err
	}

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
		Key:      f.Key,
		Label:    f.Label,
		Type:     model.CategoricalType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
	}, nil
}

func (f *TimeSeriesField) fetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {

	wheres := []string{}
	params := []interface{}{}
	var err error
	if f.Type != "timeseries" {
		// get filter where / params
		wheres, params, err = f.Storage.buildResultQueryFilters(f.StorageName, resultURI, filterParams)
		if err != nil {
			return nil, err
		}
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	keyColName := f.keyColName()

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(*) AS __count__
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d %s
		 GROUP BY "%s"
		 ORDER BY __count__ desc, "%s" LIMIT %d;`,
		keyColName, f.StorageName, f.Storage.getResultTable(f.StorageName),
		model.D3MIndexFieldName, len(params), where, keyColName,
		keyColName, timeSeriesCatResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationTimeSeries(histogram.Buckets)
	if err != nil {
		return nil, err
	}
	histogram.Exemplars = files
	return histogram, nil
}
