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

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// FetchResidualsExtremaByURI fetches the residual extrema by resultURI.
func (s *Storage) FetchResidualsExtremaByURI(dataset string, storageName string, resultURI string) (*api.Extrema, error) {
	storageNameResult := s.getResultTable(storageName)
	targetName, err := s.getResultTargetName(storageNameResult, resultURI)
	if err != nil {
		return nil, err
	}
	targetVariable, err := s.getResultTargetVariable(dataset, targetName)
	if err != nil {
		return nil, err
	}
	resultVariable := &model.Variable{
		Key:  "value",
		Type: model.StringType,
	}
	return s.fetchResidualsExtrema(resultURI, storageName, targetVariable, resultVariable)
}

// FetchResidualsSummary fetches a histogram of the residuals associated with a set of numerical predictions.
func (s *Storage) FetchResidualsSummary(dataset string, storageName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	storageNameResult := s.getResultTable(storageName)
	targetName, err := s.getResultTargetName(storageNameResult, resultURI)
	if err != nil {
		return nil, err
	}

	variable, err := s.getResultTargetVariable(dataset, targetName)
	if err != nil {
		return nil, err
	}

	var baseline *api.Histogram
	var filtered *api.Histogram
	baseline, err = s.fetchResidualsSummary(dataset, storageName, variable, resultURI, nil, extrema, api.MaxNumBuckets, mode)
	if err != nil {
		return nil, err
	}
	if !filterParams.Empty(true) {
		filtered, err = s.fetchResidualsSummary(dataset, storageName, variable, resultURI, filterParams, extrema, api.MaxNumBuckets, mode)
		if err != nil {
			return nil, err
		}
	}

	return &api.VariableSummary{
		Label:    variable.DisplayName,
		Key:      variable.Key,
		Type:     model.NumericalType,
		VarType:  variable.Type,
		Baseline: baseline,
		Filtered: filtered,
	}, nil
}

func (s *Storage) fetchResidualsSummary(dataset string, storageName string, variable *model.Variable, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, numBuckets int, mode api.SummaryMode) (*api.Histogram, error) {
	// Just return a nil in the case where we were asked to return residuals for a non-numeric variable.
	if model.IsNumerical(variable.Type) || variable.Type == model.TimeSeriesType {
		// update the highlight key to use the cluster if necessary
		if err := updateClusterFilters(s.metadata, dataset, filterParams, mode); err != nil {
			return nil, err
		}

		// fetch numeric histograms
		residuals, err := s.fetchResidualsHistogram(resultURI, dataset, storageName, variable, filterParams, extrema, numBuckets)
		if err != nil {
			return nil, err
		}
		return residuals, nil
	}
	return nil, errors.Errorf("variable of type %s - should be numeric", variable.Type)
}

func getErrorTyped(alias string, variableName string) string {
	fullName := fmt.Sprintf("\"%s\"", variableName)
	if alias != "" {
		fullName = fmt.Sprintf("%s.%s", alias, fullName)
	}
	return fmt.Sprintf("(cast(value as double precision) - cast(%s as double precision))", fullName)
}

func (s *Storage) getResidualsHistogramAggQuery(extrema *api.Extrema, variableName string, resultVariable *model.Variable, numBuckets int, alias string) (string, string, string) {
	// compute the bucket interval for the histogram
	interval := extrema.GetBucketInterval(numBuckets)

	// Only numeric types should occur.
	errorTyped := getErrorTyped(alias, variableName)

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", api.HistogramAggPrefix, extrema.Key)
	rounded := extrema.GetBucketMinMax(numBuckets)
	bucketQueryString := fmt.Sprintf("width_bucket(%s, %g, %g, %d) - 1",
		errorTyped, rounded.Min, rounded.Max, extrema.GetBucketCount(numBuckets))
	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, rounded.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func getResultJoin(alias string, storageName string) string {
	// FROM clause to join result and base data on d3mIdex value
	return fmt.Sprintf("%s_result as %s inner join %s as data on data.\"%s\" = %s.index",
		storageName, alias, storageName, model.D3MIndexFieldName, alias)
}

func getResidualsMinMaxAggsQuery(variableName string, resultVariable *model.Variable) string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + resultVariable.Key
	maxAggName := api.MaxAggPrefix + resultVariable.Key

	// Only numeric types should occur.
	errorTyped := getErrorTyped("", variableName)

	// create aggregations
	queryPart := fmt.Sprintf("MIN(%s) AS \"%s\", MAX(%s) AS \"%s\"", errorTyped, minAggName, errorTyped, maxAggName)

	return queryPart
}

func (s *Storage) fetchResidualsExtrema(resultURI string, storageName string, variable *model.Variable,
	resultVariable *model.Variable) (*api.Extrema, error) {

	targetName := variable.Key

	// add min / max aggregation
	aggQuery := getResidualsMinMaxAggsQuery(targetName, resultVariable)

	// from clause to join result and base data
	fromClause := getResultJoin("res", storageName)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE result_id = $1 AND target = $2 AND value != '';", aggQuery, fromClause)

	// execute the postgres query
	res, err := s.client.Query(queryString, resultURI, targetName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for result from postgres")
	}
	defer res.Close()

	return s.parseExtrema(res, variable)
}

func (s *Storage) fetchResidualsHistogram(resultURI string, datasetName, storageName string, variable *model.Variable, filterParams *api.FilterParams,
	extrema *api.Extrema, numBuckets int) (*api.Histogram, error) {
	resultVariable := &model.Variable{
		Key:  "value",
		Type: model.StringType,
	}

	// need the extrema to calculate the histogram interval
	var err error
	if extrema == nil {
		extrema, err = s.fetchResidualsExtrema(resultURI, storageName, variable, resultVariable)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch result variable extrema for summary")
		}
	} else {
		extrema.Key = variable.Key
		extrema.Type = variable.Type
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := s.getResidualsHistogramAggQuery(extrema, variable.Key, resultVariable, numBuckets, baseTableAlias)

	fromClause := getResultJoin("result", storageName)

	// create the filter for the query
	params := make([]interface{}, 0)
	params = append(params, resultURI)
	params = append(params, variable.Key)

	wheres := make([]string, 0)
	wheres, params = s.buildFilteredQueryWhere(datasetName, wheres, params, "", filterParams, false)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	// Create the complete query string.
	query := fmt.Sprintf(`
		SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count
		FROM %s
		WHERE result.result_id = $1 AND result.target = $2 AND result.%s != '' %s
		GROUP BY %s ORDER BY %s;`, bucketQuery, histogramQuery, histogramName,
		fromClause, resultVariable.Key, where, bucketQuery, histogramName)

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result variable summaries from postgres")
	}
	defer res.Close()

	field := NewNumericalField(s, datasetName, storageName, variable.Key, variable.DisplayName, variable.Type, "")

	return field.parseHistogram(res, extrema, numBuckets)
}
