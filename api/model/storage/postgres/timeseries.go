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
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// TimeSeriesField defines behaviour for the timeseries field type.
type TimeSeriesField struct {
	BasicField
	IDCol      string
	ClusterCol string
	XCol       string
	XColType   string
	YCol       string
	YColType   string
}

// NewTimeSeriesField creates a new field for timeseries types.
func NewTimeSeriesField(storage *Storage, datasetName string, datasetStorageName string, clusterCol string, key string, label string, typ string,
	idCol string, xCol string, xColType string, yCol string, yColType string) *TimeSeriesField {
	field := &TimeSeriesField{
		BasicField: BasicField{
			Storage:            storage,
			DatasetName:        datasetName,
			DatasetStorageName: datasetStorageName,
			Label:              label,
			Type:               typ,
			Key:                key,
		},
		IDCol:      idCol,
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
				return nil, errors.Wrap(err, "failed to parse row result")
			}
			points = append(points, []float64{x, y})
		}
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i][0] < points[j][0]
	})

	return points, nil
}

func (s *Storage) parseDateTimeTimeseries(rows *pgx.Rows) ([][]float64, error) {
	var points [][]float64
	if rows != nil {
		for rows.Next() {
			var time time.Time
			var value float64
			err := rows.Scan(&time, &value)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse row result")
			}
			points = append(points, []float64{float64(time.Unix() * 1000), value})
		}
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i][0] < points[j][0]
	})

	return points, nil
}

func (f *TimeSeriesField) fetchRepresentationTimeSeries(categoryBuckets []*api.Bucket, mode api.SummaryMode) ([]string, error) {

	var timeseriesExemplars []string

	for _, bucket := range categoryBuckets {

		keyColName := f.keyColName(mode)

		// pull sample row containing bucket
		query := fmt.Sprintf("SELECT \"%s\" FROM %s WHERE \"%s\" = $1 LIMIT 1;",
			f.IDCol, f.DatasetStorageName, keyColName)

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
func (s *Storage) FetchTimeseries(dataset string, storageName string, timeseriesColName string, xColName string, yColName string, timeseriesURI string, filterParams *api.FilterParams, invert bool) (*api.TimeseriesData, error) {
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

	xColVariable, err := s.metadata.FetchVariable(dataset, xColName)
	if err != nil {
		return nil, err
	}
	var response [][]float64
	var dateTime bool
	if xColVariable.Type == model.DateTimeType {
		response, err = s.parseDateTimeTimeseries(res)
		dateTime = true
		if err != nil {
			return nil, err
		}
	} else {
		// sum duplicate timestamps
		response, err = s.parseTimeseries(res)
		if err != nil {
			return nil, err
		}
	}
	response, err = removeDuplicates(response), nil
	if err != nil {
		return nil, err
	}
	return &api.TimeseriesData{
		Timeseries: response,
		IsDateTime: dateTime,
	}, nil
}

// FetchTimeseriesForecast fetches a timeseries.
func (s *Storage) FetchTimeseriesForecast(dataset string, storageName string, timeseriesColName string, xColName string, yColName string, timeseriesURI string, resultURI string, filterParams *api.FilterParams) (*api.TimeseriesData, error) {
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

	// Fetch the timeseries data point.  They are stored either as an int value
	// or as postgres Timestamp vlue.
	xColVariable, err := s.metadata.FetchVariable(dataset, xColName)
	if err != nil {
		return nil, err
	}
	var response [][]float64
	var dateTime bool
	if xColVariable.Type == model.DateTimeType {
		response, err = s.parseDateTimeTimeseries(res)
		dateTime = true
		if err != nil {
			return nil, err
		}
	} else {
		response, err = s.parseTimeseries(res)
		if err != nil {
			return nil, err
		}
	}
	// Sum duplicate timestamps
	response = removeDuplicates(response)
	return &api.TimeseriesData{
		Timeseries: response,
		IsDateTime: dateTime,
	}, nil
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *TimeSeriesField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool, mode api.SummaryMode) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	// update the highlight key to use the cluster if necessary
	if err = f.updateClusterHighlight(filterParams, mode); err != nil {
		return nil, err
	}

	if resultURI == "" {
		baseline, err = f.fetchHistogram(nil, invert, mode)
		if err != nil {
			return nil, err
		}
		if !filterParams.Empty() {
			filtered, err = f.fetchHistogram(filterParams, invert, mode)
			if err != nil {
				return nil, err
			}
		}
	} else {
		baseline, err = f.fetchHistogramByResult(resultURI, nil, mode)
		if err != nil {
			return nil, err
		}
		if !filterParams.Empty() {
			filtered, err = f.fetchHistogramByResult(resultURI, filterParams, mode)
			if err != nil {
				return nil, err
			}
		}
	}

	// Handle timeseries that use a timestamp/int as their time value, or those that use a date time.
	var timelineField TimelineField
	if f.XColType == model.DateTimeType {
		timelineField = NewDateTimeField(f.Storage, f.DatasetName, f.DatasetStorageName, f.XCol, f.XCol, f.XColType, f.Count)
	} else if f.XColType == model.TimestampType || f.XColType == model.IntegerType {
		timelineField = NewNumericalField(f.Storage, f.DatasetName, f.DatasetStorageName, f.XCol, f.XCol, f.XColType, f.Count)
	} else {
		return nil, errors.Errorf("unsupported timeseries field variable type %s:%s", f.XCol, f.XColType)
	}

	timelineBaseline, err := timelineField.fetchHistogram(nil, invert, api.MaxNumBuckets)
	if err != nil {
		return nil, err
	}
	timeline, err := timelineField.fetchHistogram(filterParams, invert, api.MaxNumBuckets)
	if err != nil {
		return nil, err
	}

	return &api.VariableSummary{
		Key:              f.Key,
		Label:            f.Label,
		Type:             model.CategoricalType,
		VarType:          f.Type,
		Baseline:         baseline,
		Filtered:         filtered,
		Timeline:         timeline,
		TimelineBaseline: timelineBaseline,
		TimelineType:     f.XColType,
	}, nil
}

func (f *TimeSeriesField) keyColName(mode api.SummaryMode) string {
	if mode == api.ClusterMode && api.HasClusterData(f.GetDatasetName(), f.ClusterCol, f.GetStorage().metadata) {
		return f.ClusterCol
	}
	return f.IDCol
}

func (f *TimeSeriesField) fetchHistogram(filterParams *api.FilterParams, invert bool, mode api.SummaryMode) (*api.Histogram, error) {

	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, "", filterParams, false)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	colName := f.keyColName(mode)
	query := fmt.Sprintf("SELECT \"%s\", COUNT(DISTINCT \"%s\") AS __count__ FROM %s %s GROUP BY \"%s\" ORDER BY __count__ desc, \"%s\" LIMIT %d;",
		colName, f.IDCol, f.DatasetStorageName, where, colName, colName, timeSeriesCatResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res, mode)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationTimeSeries(histogram.Buckets, mode)
	if err != nil {
		return nil, err
	}
	histogram.Exemplars = files
	return histogram, nil
}

func (f *TimeSeriesField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams, mode api.SummaryMode) (*api.Histogram, error) {

	wheres := []string{}
	params := []interface{}{}
	var err error
	if f.Type != "timeseries" {
		// get filter where / params
		wheres, params, err = f.Storage.buildResultQueryFilters(f.DatasetStorageName, resultURI, filterParams)
		if err != nil {
			return nil, err
		}
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	keyColName := f.keyColName(mode)

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(DISTINCT "%s") AS __count__
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d %s
		 GROUP BY "%s"
		 ORDER BY __count__ desc, "%s" LIMIT %d;`,
		keyColName, f.IDCol, f.DatasetStorageName, f.Storage.getResultTable(f.DatasetStorageName),
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

	histogram, err := f.parseHistogram(res, mode)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationTimeSeries(histogram.Buckets, mode)
	if err != nil {
		return nil, err
	}
	histogram.Exemplars = files
	return histogram, nil
}

func (f *TimeSeriesField) parseHistogram(rows *pgx.Rows, mode api.SummaryMode) (*api.Histogram, error) {
	keyColName := f.keyColName(mode)

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
func (f *TimeSeriesField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	// update the highlight key to use the cluster if necessary
	if err = f.updateClusterHighlight(filterParams, mode); err != nil {
		return nil, err
	}

	baseline, err = f.fetchPredictedSummaryData(resultURI, datasetResult, nil, extrema, mode)
	if err != nil {
		return nil, err
	}
	if !filterParams.Empty() {
		filtered, err = f.fetchPredictedSummaryData(resultURI, datasetResult, filterParams, extrema, mode)
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

func (f *TimeSeriesField) fetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.Histogram, error) {

	wheres := []string{}
	params := []interface{}{}
	var err error
	if f.Type != "timeseries" {
		// get filter where / params
		wheres, params, err = f.Storage.buildResultQueryFilters(f.DatasetStorageName, resultURI, filterParams)
		if err != nil {
			return nil, err
		}
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	keyColName := f.keyColName(mode)

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(DISTINCT "%s") AS __count__
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d %s
		 GROUP BY "%s"
		 ORDER BY __count__ desc, "%s" LIMIT %d;`,
		keyColName, f.IDCol, f.DatasetStorageName, f.Storage.getResultTable(f.DatasetStorageName),
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

	histogram, err := f.parseHistogram(res, mode)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationTimeSeries(histogram.Buckets, mode)
	if err != nil {
		return nil, err
	}
	histogram.Exemplars = files
	return histogram, nil
}

// Sums any duplicate timestamps encountered.  Assumes data is sorted.
func removeDuplicates(timeseriesData [][]float64) [][]float64 {
	cleanedData := [][]float64{}
	currIdx := 0
	for currIdx < len(timeseriesData)-1 {
		timestamp := timeseriesData[currIdx][0]
		nextTimestamp := timeseriesData[currIdx+1][0]

		if timestamp != nextTimestamp {
			count := timeseriesData[currIdx][1]
			cleanedData = append(cleanedData, []float64{timestamp, count})
			currIdx++
		} else {
			first := true
			for timestamp == nextTimestamp {
				count := timeseriesData[currIdx][1]
				if first {
					// add current until next doesn't match or next is out of bounds
					cleanedData = append(cleanedData, []float64{timestamp, count})
					first = false
				} else {
					cleanedData[len(cleanedData)-1][1] += count
				}
				currIdx++
				if currIdx == len(timeseriesData)-1 {
					break
				}
				timestamp = timeseriesData[currIdx][0]
				nextTimestamp = timeseriesData[currIdx+1][0]
			}
			count := timeseriesData[currIdx][1]
			cleanedData[len(cleanedData)-1][1] += count
			currIdx++
		}
	}

	// last element is different than second last
	if currIdx == len(timeseriesData)-1 {
		cleanedData = append(cleanedData, timeseriesData[currIdx])
	}
	return cleanedData
}
