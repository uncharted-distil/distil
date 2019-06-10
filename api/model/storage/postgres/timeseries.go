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

// TimeSeriesField defines behaviour for the timeseries field type.
type TimeSeriesField struct {
	Storage     *Storage
	StorageName string
	ClusterCol  string
	IDCol       string
	Label       string
	Type        string
}

// NewTimeSeriesField creates a new field for timeseries types.
func NewTimeSeriesField(storage *Storage, storageName string, clusterCol string, idCol string, label string, typ string) *TimeSeriesField {
	field := &TimeSeriesField{
		Storage:     storage,
		StorageName: storageName,
		IDCol:       idCol,
		ClusterCol:  clusterCol,
		Label:       label,
		Type:        typ,
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
	return points, nil
}

func (f *TimeSeriesField) fetchRepresentationTimeSeries(categoryBuckets []*api.Bucket) ([]string, error) {

	var timeseriesExemplars []string

	for _, bucket := range categoryBuckets {

		timeseriesIDField := "series_id"

		prefixedVarName := f.clusterVarName(f.ClusterCol)

		// pull sample row containing bucket
		query := fmt.Sprintf("SELECT \"%s\" FROM %s WHERE \"%s\" = $1 LIMIT 1;",
			timeseriesIDField, f.StorageName, prefixedVarName)

		// execute the postgres query
		rows, err := f.Storage.client.Query(query, bucket.Key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
		}

		if rows.Next() {
			var timeseriesExemplar int
			err = rows.Scan(&timeseriesExemplar)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to parse solution from Postgres")
			}
			timeseriesExemplars = append(timeseriesExemplars, strconv.FormatInt(int64(timeseriesExemplar), 10))
		}
		rows.Close()
	}

	if len(timeseriesExemplars) == 0 {
		return nil, fmt.Errorf("No exemplars found for timeseries data")
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

	wheres, params = s.buildFilteredQueryWhere(wheres, params, filterParams, invert)
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

// FetchTimeseriesSummaryData pulls summary data from the database and builds a histogram.
func (f *TimeSeriesField) FetchTimeseriesSummaryData(timeVar *model.Variable, interval int, resultURI string, filterParams *api.FilterParams, invert bool) (*api.VariableSummary, error) {
	return nil, fmt.Errorf("not implemented")
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *TimeSeriesField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error
	if resultURI == "" {
		baseline, err = f.fetchHistogram(nil, invert)
		if err != nil {
			return nil, err
		}
		if filterParams.Filters != nil {
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
		if filterParams.Filters != nil {
			filtered, err = f.fetchHistogramByResult(resultURI, filterParams)
			if err != nil {
				return nil, err
			}
		}
	}

	return &api.VariableSummary{
		Key:      f.IDCol,
		Label:    f.Label,
		Type:     model.CategoricalType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
	}, nil
}

func (f *TimeSeriesField) clusterVarName(varName string) string {
	return fmt.Sprintf("%s%s", model.ClusterVarPrefix, varName)
}

func (f *TimeSeriesField) fetchHistogram(filterParams *api.FilterParams, invert bool) (*api.Histogram, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, filterParams, false)

	prefixedVarName := f.clusterVarName(f.ClusterCol)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT \"%s\", COUNT(*) AS count FROM %s %s GROUP BY \"%s\" ORDER BY count desc, \"%s\" LIMIT %d;",
		prefixedVarName, f.StorageName, where, prefixedVarName, prefixedVarName, catResultLimit)

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

	prefixedVarName := f.clusterVarName(f.ClusterCol)

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(*) AS count
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d %s
		 GROUP BY "%s"
		 ORDER BY count desc, "%s" LIMIT %d;`,
		prefixedVarName, f.StorageName, f.Storage.getResultTable(f.StorageName),
		model.D3MIndexFieldName, len(params), where, prefixedVarName,
		prefixedVarName, catResultLimit)

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
	prefixedVarName := f.clusterVarName(f.ClusterCol)

	termsAggName := api.TermsAggPrefix + prefixedVarName

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

	baseline, err = f.fetchPredictedSummaryData(resultURI, datasetResult, nil, extrema)
	if err != nil {
		return nil, err
	}
	if filterParams.Filters != nil {
		filtered, err = f.fetchPredictedSummaryData(resultURI, datasetResult, filterParams, extrema)
		if err != nil {
			return nil, err
		}
	}
	return &api.VariableSummary{
		Key:      f.IDCol,
		Label:    f.Label,
		Type:     model.CategoricalType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
	}, nil
}

func (f *TimeSeriesField) fetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	targetName := f.clusterVarName(f.ClusterCol)

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.StorageName, resultURI, filterParams)
	if err != nil {
		return nil, err
	}

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, targetName)

	query := fmt.Sprintf(
		`SELECT data."%s", result.value, COUNT(*) AS count
		 FROM %s AS result INNER JOIN %s AS data ON result.index = data."%s"
		 WHERE %s
		 GROUP BY result.value, data."%s"
		 ORDER BY count desc;`,
		targetName, datasetResult, f.StorageName, model.D3MIndexFieldName, strings.Join(wheres, " AND "), targetName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

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

// FetchForecastingSummaryData pulls data from the result table and builds the
// forecasting histogram for the field.
func (f *TimeSeriesField) FetchForecastingSummaryData(timeVar *model.Variable, interval int, resultURI string, filterParams *api.FilterParams) (*api.VariableSummary, error) {
	return nil, fmt.Errorf("not implemented")
}
