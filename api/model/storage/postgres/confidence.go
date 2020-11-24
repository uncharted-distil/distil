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
	"reflect"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// FetchConfidenceSummary fetches a histogram of the confidence and explanations associated with a set of classification predictions.
func (s *Storage) FetchConfidenceSummary(dataset string, storageName string, resultURI string, filterParams *api.FilterParams, mode api.SummaryMode) (map[string]*api.VariableSummary, error) {
	explainFields := s.listExplainFields()
	storageNameResult := s.getResultTable(storageName)
	targetName, err := s.getResultTargetName(storageNameResult, resultURI)
	if err != nil {
		return nil, err
	}

	variable, err := s.getResultTargetVariable(dataset, targetName)
	if err != nil {
		return nil, err
	}

	explainedSummaries := map[string]*api.VariableSummary{}
	for _, explainName := range explainFields {
		var baseline *api.Histogram
		var filtered *api.Histogram
		baseline, err = s.fetchExplainHistogram(dataset, storageName, targetName, explainName, resultURI, nil, mode)
		if err != nil {
			return nil, err
		}
		if !filterParams.Empty() {
			filtered, err = s.fetchExplainHistogram(dataset, storageName, targetName, explainName, resultURI, filterParams, mode)
			if err != nil {
				return nil, err
			}
		}

		explainedSummaries[explainName] = &api.VariableSummary{
			Label:    variable.DisplayName,
			Key:      variable.Name,
			Type:     model.NumericalType,
			VarType:  variable.Type,
			Baseline: baseline,
			Filtered: filtered,
		}
	}

	// TODO: FOLD CONFIDENCES INTO THE SAME JSON FIELD IN THE DATABASE AND THEN DELETE THIS!!!!!
	var baseline *api.Histogram
	var filtered *api.Histogram
	baseline, err = s.fetchConfidenceHistogram(dataset, storageName, variable, targetName, resultURI, nil, mode)
	if err != nil {
		return nil, err
	}
	if !filterParams.Empty() {
		filtered, err = s.fetchConfidenceHistogram(dataset, storageName, variable, targetName, resultURI, filterParams, mode)
		if err != nil {
			return nil, err
		}
	}

	explainedSummaries["confidence"] = &api.VariableSummary{
		Label:    variable.DisplayName,
		Key:      variable.Name,
		Type:     model.NumericalType,
		VarType:  variable.Type,
		Baseline: baseline,
		Filtered: filtered,
	}

	return explainedSummaries, nil
}

func (s *Storage) fetchHistograms(dataset string, storageName string, variable *model.Variable, targetName string,
	resultURI string, filterParams *api.FilterParams, mode api.SummaryMode) (map[string]*api.Histogram, error) {
	explainFields := s.listExplainFields()

	explainedSummaries := map[string]*api.Histogram{}
	for _, explainName := range explainFields {
		// get the histogram for that explaination field
		histo, err := s.fetchExplainHistogram(dataset, storageName, targetName, explainName, resultURI, filterParams, mode)
		if err != nil {
			return nil, err
		}
		explainedSummaries[explainName] = histo
	}

	return explainedSummaries, nil
}

func (s *Storage) fetchExplainHistogram(dataset string, storageName string, targetName string, explainFieldName string,
	resultURI string, filterParams *api.FilterParams, mode api.SummaryMode) (*api.Histogram, error) {
	// use a numerical sub select
	field := NewNumericalFieldSubSelect(s, dataset, storageName, explainFieldName, explainFieldName, model.IntegerType, "", s.explainSubSelect(storageName, explainFieldName))
	return field.fetchHistogram(filterParams, false, 20)
}

func (s *Storage) fetchConfidenceHistogram(dataset string, storageName string, variable *model.Variable, targetName string, resultURI string, filterParams *api.FilterParams, mode api.SummaryMode) (*api.Histogram, error) {
	storageNameResult := s.getResultTable(storageName)

	// get filter where / params
	wheres, params, err := s.buildResultQueryFilters(dataset, storageName, resultURI, filterParams, baseTableAlias)
	if err != nil {
		return nil, err
	}

	countCol, err := s.getCountCol(dataset, mode)
	if err != nil {
		return nil, err
	}
	if countCol == "" {
		countCol = "*"
	} else {
		countCol = fmt.Sprintf("DISTINCT \"%s\"", countCol)
	}

	wheres = append(wheres, "confidence is not null")
	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, targetName)

	query := fmt.Sprintf(
		`SELECT floor(result.confidence / 0.02) as bucket, COUNT(%s) AS count
		 FROM %s AS result INNER JOIN %s AS data ON result.index = data."%s"
		 WHERE %s
		 GROUP BY floor(result.confidence / 0.02);`,
		countCol, storageNameResult, storageName, model.D3MIndexFieldName, strings.Join(wheres, " AND "))

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return s.parseConfidenceHistogram(res, variable)
}

func (s *Storage) parseConfidenceHistogram(rows pgx.Rows, variable *model.Variable) (*api.Histogram, error) {

	// parse the confidence data
	countMap := map[int]int64{}
	if rows != nil {
		for rows.Next() {
			var bucket int
			var bucketCount int64
			err := rows.Scan(&bucket, &bucketCount)
			if err != nil {
				return nil, errors.Wrap(err, "no confidence histogram aggregation found")
			}
			countMap[bucket] = bucketCount
		}
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}
	}

	if len(countMap) == 0 {
		return nil, nil
	}

	// create buckets from 0 to 50
	// bucket 50 is the bucket for instances with confidence = 1
	buckets := make([]*api.Bucket, 51)
	for i := 0; i <= 50; i++ {
		buckets[i] = &api.Bucket{
			Key:   fmt.Sprintf("%f", float64(i)*0.02),
			Count: countMap[i],
		}
	}

	// assign histogram attributes
	return &api.Histogram{
		Buckets: buckets,
		Extrema: &api.Extrema{
			Min: float64(0),
			Max: float64(1),
		},
	}, nil
}

func (s *Storage) listExplainFields() []string {
	v := reflect.TypeOf(api.SolutionExplainValues{}).Elem()
	jsonNames := []string{}
	for j := 0; j < v.NumField(); j++ {
		f := v.Field(j)
		jsonNames = append(jsonNames, f.Tag.Get("json"))
	}

	return jsonNames
}

func (s *Storage) explainSubSelect(storageName string, fieldName string) func() string {
	return func() string {
		return fmt.Sprintf("(SELECT (explain_values ->> '%s')::double precision, index::text as \"%s\" from %s)", fieldName, model.D3MIndexFieldName, storageName)
	}
}
