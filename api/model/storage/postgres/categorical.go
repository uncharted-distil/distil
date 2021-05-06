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
	"math"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// CategoricalField defines behaviour for the categorical field type.
type CategoricalField struct {
	BasicField
	subSelect func() string
}

// NewCategoricalField creates a new field for categorical types.
func NewCategoricalField(storage *Storage, datasetName string, datasetStorageName string, key string, label string, typ string, count string) *CategoricalField {
	count = getCountSQL(count)

	field := &CategoricalField{
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

// NewCategoricalFieldSubSelect creates a new field for categorical types
// and specifies a sub select query to pull the raw data.
func NewCategoricalFieldSubSelect(storage *Storage, datasetName string,
	datasetStorageName string, key string, label string, typ string, count string, fieldSubSelect func() string) *CategoricalField {
	count = getCountSQL(count)

	field := &CategoricalField{
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
func (f *CategoricalField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	// update the highlight key to use the cluster if necessary
	if err = f.updateClusterHighlight(filterParams, mode); err != nil {
		return nil, err
	}

	if resultURI == "" {
		baseline, err = f.fetchHistogram(api.GetBaselineFilter(filterParams))
		if err != nil {
			return nil, err
		}
		if !filterParams.IsEmpty(true) {
			filtered, err = f.fetchHistogram(filterParams)
			if err != nil {
				return nil, err
			}
		}
	} else {
		baseline, err = f.fetchHistogramByResult(resultURI, api.GetBaselineFilter(filterParams))
		if err != nil {
			return nil, err
		}
		if !filterParams.IsEmpty(true) {
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

func (f *CategoricalField) fetchExtremaStorage() (*api.Extrema, error) {
	// pull all unique values from the database
	sql := fmt.Sprintf("SELECT DISTINCT \"%s\" FROM %s;", f.Key, f.GetDatasetStorageName())
	res, err := f.Storage.client.Query(sql)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseExtrema(res)
}

func (f *CategoricalField) parseExtrema(rows pgx.Rows) (*api.Extrema, error) {
	result := &api.Extrema{
		Key:    f.Key,
		Type:   f.Type,
		Values: []string{},
	}
	for rows.Next() {
		var value string
		err := rows.Scan(&value)
		if err != nil {
			return nil, errors.Wrap(err, "unable to scan categorical value")
		}
		if len(value) < 1 {
			value = "<none>"
		}
		result.Values = append(result.Values, value)
	}
	err := rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return result, nil
}

func (f *CategoricalField) fetchHistogram(filterParams *api.FilterParams) (*api.Histogram, error) {
	fromClause := f.getFromClause(true)

	// create the filter for the query
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(f.GetDatasetName(), wheres, params, "", filterParams)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT \"%s\", COUNT(%s) AS count FROM %s %s GROUP BY \"%s\" ORDER BY count desc, \"%s\" LIMIT %d;",
		f.Key, f.Count, fromClause, where, f.Key, f.Key, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res)
}

func (f *CategoricalField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams) (*api.Histogram, error) {
	fromClause := f.getFromClause(false)

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

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(%s) AS count
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d AND result.value != '' %s
		 GROUP BY data."%s"
		 ORDER BY count desc, "%s" LIMIT %d;`,
		f.Key, f.Count, fromClause, f.Storage.getResultTable(f.DatasetStorageName),
		model.D3MIndexFieldName, len(params), where, f.Key,
		f.Key, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res)
}

func (f *CategoricalField) parseHistogram(rows pgx.Rows) (*api.Histogram, error) {
	termsAggName := api.TermsAggPrefix + f.Key

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
			if len(term) < 1 {
				term = "<none>"
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
// the categorical histogram for the field.
func (f *CategoricalField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	// update the highlight key to use the cluster if necessary
	if err = f.updateClusterHighlight(filterParams, mode); err != nil {
		return nil, err
	}

	baseline, err = f.fetchPredictedSummaryData(resultURI, datasetResult, api.GetBaselineFilter(filterParams), extrema)
	if err != nil {
		return nil, err
	}
	if !filterParams.IsEmpty(true) {
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

func (f *CategoricalField) fetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	targetName := f.Key

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.GetDatasetName(), f.DatasetStorageName, resultURI, filterParams, baseTableAlias)
	if err != nil {
		return nil, err
	}

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, targetName)

	query := fmt.Sprintf(
		`SELECT result.value, COUNT(%s) AS count
		 FROM %s AS result INNER JOIN %s AS data ON result.index = data."%s"
		 WHERE %s
		 GROUP BY result.value
		 ORDER BY count desc;`,
		f.Count, datasetResult, f.DatasetStorageName, model.D3MIndexFieldName, strings.Join(wheres, " AND "))

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return f.parseHistogram(res)
}

func (f *CategoricalField) getFromClause(alias bool) string {
	fromClause := f.DatasetStorageName
	if f.subSelect != nil {
		fromClause = f.subSelect()
		if alias {
			fromClause = fmt.Sprintf("%s as %s", fromClause, f.DatasetStorageName)
		}
	}

	return fromClause
}
