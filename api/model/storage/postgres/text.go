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
	postgres "github.com/uncharted-distil/distil/api/postgres"
)

// TextField defines behaviour for the text field type.
type TextField struct {
	BasicField
}

// NewTextField creates a new field for text types.
func NewTextField(storage *Storage, datasetName string, datasetStorageName string, key string, label string, typ string, count string) *TextField {
	field := &TextField{
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

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *TextField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool, mode api.SummaryMode) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	// update the highlight key to use the cluster if necessary
	if err = f.updateClusterHighlight(filterParams, mode); err != nil {
		return nil, err
	}

	if resultURI == "" {
		baseline, err = f.fetchHistogram(api.GetBaselineFilter(filterParams), invert)
		if err != nil {
			return nil, err
		}
		if !filterParams.Empty(true) {
			filtered, err = f.fetchHistogram(filterParams, invert)
			if err != nil {
				return nil, err
			}
		}
	} else {
		baseline, err = f.fetchHistogramByResult(resultURI, api.GetBaselineFilter(filterParams))
		if err != nil {
			return nil, err
		}
		if !filterParams.Empty(true) {
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

func (f *TextField) fetchHistogram(filterParams *api.FilterParams, invert bool) (*api.Histogram, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(f.GetDatasetName(), wheres, params, "", filterParams, invert)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	countSubselect := ""
	if f.Count != "" {
		countSubselect = fmt.Sprintf(", \"%s\"", f.Count)
	}

	query := fmt.Sprintf("SELECT COALESCE(w.word, r.stem) as %s, COUNT(%s) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(\"%s\"))) as stem %s FROM %s %s) as r "+
		"LEFT OUTER JOIN %s as w on r.stem = w.stem "+
		"GROUP BY COALESCE(w.word, r.stem) ORDER BY count desc, COALESCE(w.word, r.stem) LIMIT %d;",
		f.Key, getCountSQL(f.Count), f.Key, countSubselect, f.DatasetStorageName, where, postgres.WordStemTableName, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch text histogram for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res)
}

func (f *TextField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams) (*api.Histogram, error) {

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
	countSubselect := ""
	if f.Count != "" {
		countSubselect = fmt.Sprintf(", \"%s\"", f.Count)
	}

	query := fmt.Sprintf("SELECT COALESCE(w.word, r.stem) as \"%s\", COUNT(%s) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(\"%s\"))) as stem %s "+
		"FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $%d %s) as r "+
		"LEFT OUTER JOIN %s as w on r.stem = w.stem "+
		"GROUP BY COALESCE(w.word, r.stem) ORDER BY count desc, COALESCE(w.word, r.stem) LIMIT %d;",
		f.Key, getCountSQL(f.Count), f.Key, countSubselect, f.DatasetStorageName, f.Storage.getResultTable(f.DatasetStorageName),
		model.D3MIndexFieldName, len(params), where, postgres.WordStemTableName, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch text histogram for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res)
}

func (f *TextField) parseHistogram(rows pgx.Rows) (*api.Histogram, error) {
	termsAggName := api.TermsAggPrefix + f.Key

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

// FetchPredictedSummaryData pulls data from the result table and builds
// the categorical histogram for the field.
func (f *TextField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	// update the highlight key to use the cluster if necessary
	if err = f.updateClusterHighlight(filterParams, mode); err != nil {
		return nil, err
	}

	baseline, err = f.fetchPredictedSummaryData(resultURI, datasetResult, nil, extrema)
	if err != nil {
		return nil, err
	}
	if !filterParams.Empty(true) {
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

func (f *TextField) fetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	targetName := f.Key

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.GetDatasetName(), f.DatasetStorageName, resultURI, filterParams, "base")
	if err != nil {
		return nil, err
	}

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, targetName)

	countSubselect := ""
	if f.Count != "" {
		countSubselect = fmt.Sprintf(", \"%s\"", f.Count)
	}

	query := fmt.Sprintf("SELECT COALESCE(word_b.word, r.stem_b) as \"%s\", COALESCE(word_v.word, r.stem_v) as value, COUNT(%s) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(base.\"%s\"))) as stem_b, "+
		"unnest(tsvector_to_array(to_tsvector(result.value))) as stem_v %s"+
		"FROM %s AS result INNER JOIN %s AS base ON result.index = base.\"d3mIndex\" "+
		"WHERE %s) r LEFT OUTER JOIN %s word_b ON r.stem_b = word_b.stem LEFT OUTER JOIN %s word_v ON r.stem_v = word_v.stem "+
		"GROUP BY COALESCE(word_v.word, r.stem_v), COALESCE(word_b.word, r.stem_b) "+
		"ORDER BY count desc;", targetName, getCountSQL(f.Count), targetName, countSubselect, datasetResult,
		f.DatasetStorageName, strings.Join(wheres, " AND "), postgres.WordStemTableName, postgres.WordStemTableName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return f.parseHistogram(res)
}
