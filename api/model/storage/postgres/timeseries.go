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
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
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

func (s *Storage) parseTimeseries(rows pgx.Rows) ([]*api.TimeseriesObservation, error) {
	var points []*api.TimeseriesObservation
	if rows != nil {
		for rows.Next() {
			var x float64
			var y float64
			err := rows.Scan(&x, &y)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse row result")
			}
			points = append(points, &api.TimeseriesObservation{
				Value: api.NullableFloat64(y),
				Time:  x,
			})
		}
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}
	}

	return points, nil
}

func (s *Storage) parseDateTimeTimeseries(rows pgx.Rows) ([]*api.TimeseriesObservation, error) {
	var points []*api.TimeseriesObservation
	if rows != nil {
		for rows.Next() {
			var time time.Time
			var value float64

			err := rows.Scan(&time, &value)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse row result")
			}

			points = append(points, &api.TimeseriesObservation{
				Value: api.NullableFloat64(value),
				Time:  float64(time.Unix() * 1000),
			})
		}
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}
	}

	return points, nil
}

func (s *Storage) parseTimeseriesForecast(rows pgx.Rows) ([]*api.TimeseriesObservation, error) {
	var points []*api.TimeseriesObservation
	if rows != nil {
		for rows.Next() {
			var time float64
			var value float64
			var explainValues api.SolutionExplainValues
			err := rows.Scan(&time, &value, &explainValues)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse row result")
			}
			points = append(points, &api.TimeseriesObservation{
				Value:          api.NullableFloat64(value),
				Time:           time,
				ConfidenceLow:  api.NullableFloat64(explainValues.LowConfidence),
				ConfidenceHigh: api.NullableFloat64(explainValues.HighConfidence),
			})
		}
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}
	}

	return points, nil
}

func (s *Storage) parseDateTimeTimeseriesForecast(rows pgx.Rows) ([]*api.TimeseriesObservation, error) {
	var points []*api.TimeseriesObservation
	if rows != nil {
		for rows.Next() {
			var time time.Time
			var value float64
			var explainValuesJSON string

			err := rows.Scan(&time, &value, &explainValuesJSON)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse row result")
			}
			var explainValues *api.SolutionExplainValues
			err = json.Unmarshal([]byte(explainValuesJSON), explainValues)
			if err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal explained values")
			}

			points = append(points, &api.TimeseriesObservation{
				Value:          api.NullableFloat64(value),
				Time:           float64(time.Unix() * 1000),
				ConfidenceLow:  api.NullableFloat64(explainValues.LowConfidence),
				ConfidenceHigh: api.NullableFloat64(explainValues.HighConfidence),
			})
		}
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}
	}

	return points, nil
}

// Calculate the Min, Max, and Mean of a list of TimerseriesObservation
func getMinMaxMean(timeseries []*api.TimeseriesObservation) (float64, float64, float64) {
	min := math.Inf(1)
	max := math.Inf(-1)
	sum := float64(0)

	var value float64
	for _, timeserie := range timeseries {
		value = float64(timeserie.Value)
		if !math.IsNaN(value) {
			min = math.Min(value, min)
			max = math.Max(value, max)
			sum += value
		}
	}

	// Check that the values have been updated
	minOk := !math.IsInf(min, 0)
	maxOk := !math.IsInf(max, 0)
	if minOk && maxOk {

		// Calculate the mean
		mean := sum / float64(len(timeseries))

		// Send them back as NullableFloat64
		return min, max, mean
	}

	// Otherwise, send a NaN
	var null = math.NaN()
	return null, null, null
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
		err = rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
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

	wheres, params = s.buildFilteredQueryWhere(dataset, wheres, params, "", filterParams, invert)
	where := fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))

	// Get count by category.
	query := fmt.Sprintf("SELECT timeline.TimeStamps, COALESCE(filteredEvents.Counts, 'NaN') FROM "+
		"(SELECT DISTINCT \"%s\" as TimeStamps FROM %s) timeline LEFT JOIN "+
		"(SELECT \"%s\" as TimeStamps, \"%s\" as Counts FROM %s %s ) filteredEvents "+
		"ON timeline.TimeStamps = filteredEvents.TimeStamps ORDER BY timeline.TimeStamps",
		xColName, storageName, xColName, yColName, storageName, where)

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
	var response []*api.TimeseriesObservation
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

	// sort the timeseries
	sort.Slice(response, func(i, j int) bool {
		return response[i].Time < response[j].Time
	})

	// sum duplicate timestamps
	response = removeDuplicates(response)

	// Calculate Min/Max/Mean
	var min, max, mean = getMinMaxMean(response)

	return &api.TimeseriesData{
		Timeseries: response,
		IsDateTime: dateTime,
		Min:        min,
		Max:        max,
		Mean:       mean,
	}, nil
}

// FetchTimeseriesForecast fetches a timeseries.
func (s *Storage) FetchTimeseriesForecast(dataset string, storageName string, timeseriesColName string, xColName string, yColName string, timeseriesURI string, resultURI string, filterParams *api.FilterParams) (*api.TimeseriesData, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	wheres = append(wheres, fmt.Sprintf("\"%s\" = $1", timeseriesColName))
	params = append(params, timeseriesURI)

	wheres, params = s.buildFilteredQueryWhere(dataset, wheres, params, "", filterParams, false)

	params = append(params, resultURI)
	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d", len(params)))
	wheres = append(wheres, "result.value != ''")

	where := fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))

	// Get count by category.
	query := fmt.Sprintf(`
		SELECT "%s", CAST(CASE WHEN result.value = '' THEN 'NaN' ELSE result.value END as double precision),
		coalesce(result.explain_values, '{}') AS explain_values
		FROM %s data INNER JOIN %s result ON data."%s" = result.index
		%s
		ORDER BY %s`,
		xColName, storageName, s.getResultTable(storageName),
		model.D3MIndexFieldName, where, xColName)

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
	var response []*api.TimeseriesObservation
	var dateTime bool
	if xColVariable.Type == model.DateTimeType {
		response, err = s.parseDateTimeTimeseriesForecast(res)
		dateTime = true
		if err != nil {
			return nil, err
		}
	} else {
		response, err = s.parseTimeseriesForecast(res)
		if err != nil {
			return nil, err
		}
	}
	// Sum duplicate timestamps
	response = removeDuplicates(response)

	// Calculate Min/Max/Mean
	var min, max, mean = getMinMaxMean(response)

	return &api.TimeseriesData{
		Timeseries: response,
		IsDateTime: dateTime,
		Min:        min,
		Max:        max,
		Mean:       mean,
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

	// split the filters to make sure the result based filters can be applied properly
	filtersSplit := splitFilters(filterParams)
	joins := make([]*joinDefinition, 0)
	wheres := []string{}
	params := []interface{}{}
	if filtersSplit.correctnessFilter != nil {

	}
	if filtersSplit.predictedFilter != nil {

	}
	if filtersSplit.residualFilter != nil {
		wheres, params, err = f.Storage.buildErrorResultWhere(wheres, params, filtersSplit.residualFilter)
		if err != nil {
			return nil, err
		}

		joins = append(joins, &joinDefinition{
			baseAlias:  baseTableAlias,
			baseColumn: f.IDCol,
			joinAlias:  "r",
			joinColumn: "k",
			joinTableName: fmt.Sprintf("(SELECT DISTINCT \"%s\" AS k FROM %s AS b INNER JOIN %s AS r ON b.\"%s\" = r.index WHERE r.value != '' AND %s)",
				f.IDCol, f.GetDatasetStorageName(), f.Storage.getResultTable(f.GetDatasetStorageName()), model.D3MIndexFieldName, strings.Join(wheres, " AND ")),
		},
		)
	}

	// reset the filter params since the residual filter has been handled already
	filterParamsClone := filterParams.Clone()
	filterParamsClone.Highlight = nil
	filterParamsClone.Filters = filtersSplit.genericFilters

	// clear filters since they are used in subselect
	wheres = []string{}

	// Handle timeseries that use a timestamp/int as their time value, or those that use a date time.
	var timelineField TimelineField
	if f.XColType == model.DateTimeType {
		timelineField = NewDateTimeField(f.Storage, f.DatasetName, f.DatasetStorageName, f.XCol, f.XCol, f.XColType, f.Count)
	} else if f.XColType == model.TimestampType || f.XColType == model.IntegerType {
		timelineField = NewNumericalField(f.Storage, f.DatasetName, f.DatasetStorageName, f.XCol, f.XCol, f.XColType, f.Count)
	} else {
		return nil, errors.Errorf("unsupported timeseries field variable type %s:%s", f.XCol, f.XColType)
	}

	timelineBaseline, err := timelineField.fetchHistogramWithJoins(nil, invert, api.MaxNumBuckets, joins, wheres, params)
	if err != nil {
		return nil, err
	}
	timeline, err := timelineField.fetchHistogramWithJoins(filterParamsClone, invert, api.MaxNumBuckets, joins, wheres, params)
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
	wheres, params = f.Storage.buildFilteredQueryWhere(f.GetDatasetName(), wheres, params, "", filterParams, false)

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

	keyColName := f.keyColName(mode)

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(DISTINCT "%s") AS __count__
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d AND result.value != '' %s
		 GROUP BY data."%s"
		 ORDER BY __count__ desc, data."%s" LIMIT %d;`,
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

func (f *TimeSeriesField) parseHistogram(rows pgx.Rows, mode api.SummaryMode) (*api.Histogram, error) {
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
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
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

	keyColName := f.keyColName(mode)

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(DISTINCT "%s") AS __count__
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d %s
		 GROUP BY data."%s"
		 ORDER BY __count__ desc, data."%s" LIMIT %d;`,
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
func removeDuplicates(timeseriesData []*api.TimeseriesObservation) []*api.TimeseriesObservation {
	cleanedData := []*api.TimeseriesObservation{}
	observationClone := &api.TimeseriesObservation{
		Value:          timeseriesData[0].Value,
		Time:           timeseriesData[0].Time,
		ConfidenceHigh: timeseriesData[0].ConfidenceHigh,
		ConfidenceLow:  timeseriesData[0].ConfidenceLow,
	}

	// sum the timestamp values, thereby removing duplicate timestamps
	currIdx := 1
	for currIdx < len(timeseriesData) {
		observation := timeseriesData[currIdx]
		if observationClone.Time == observation.Time {
			// still dealing with the same timestamp
			observationClone.Value += observation.Value
		} else {
			// new timestamp so append the rolling count and initialize
			cleanedData = append(cleanedData, observationClone)

			observationClone = &api.TimeseriesObservation{
				Value:          observation.Value,
				Time:           observation.Time,
				ConfidenceHigh: observation.ConfidenceHigh,
				ConfidenceLow:  observation.ConfidenceLow,
			}
		}
		currIdx++
	}

	// add the last timestamp
	cleanedData = append(cleanedData, observationClone)
	return cleanedData
}
