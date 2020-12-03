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

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// FetchCorrectnessSummary fetches a histogram of the residuals associated with a set of numerical predictions.
func (s *Storage) FetchCorrectnessSummary(dataset string, storageName string, resultURI string, filterParams *api.FilterParams, mode api.SummaryMode) (*api.VariableSummary, error) {

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
	baseline, err = s.fetchHistogram(dataset, storageName, variable, targetName, resultURI, nil, mode)
	if err != nil {
		return nil, err
	}
	if !filterParams.Empty() {
		filtered, err = s.fetchHistogram(dataset, storageName, variable, targetName, resultURI, filterParams, mode)
		if err != nil {
			return nil, err
		}
	}

	return &api.VariableSummary{
		Label:    variable.DisplayName,
		Key:      variable.StorageName,
		Type:     model.CategoricalType,
		VarType:  variable.Type,
		Baseline: baseline,
		Filtered: filtered,
	}, nil
}

func (s *Storage) fetchHistogram(dataset string, storageName string, variable *model.Variable, targetName string, resultURI string, filterParams *api.FilterParams, mode api.SummaryMode) (*api.Histogram, error) {
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

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, targetName)

	query := fmt.Sprintf(
		`SELECT data."%s", result.value, COUNT(%s) AS count
		 FROM %s AS result INNER JOIN %s AS data ON result.index = data."%s"
		 WHERE %s
		 GROUP BY result.value, data."%s"
		 ORDER BY count desc;`,
		targetName, countCol, storageNameResult, storageName, model.D3MIndexFieldName, strings.Join(wheres, " AND "), targetName)

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return s.parseHistogram(res, variable)
}

func (s *Storage) getCountCol(dataset string, mode api.SummaryMode) (string, error) {
	countCol := ""
	if mode == api.MultiBandImageMode {
		// remote sensing group should be distinct by group id
		vars, err := s.metadata.FetchVariables(dataset, false, true)
		if err != nil {
			return "", err
		}

		for _, v := range vars {
			if v.IsGrouping() {
				countCol = v.Grouping.GetIDCol()
			}
		}

	}

	return countCol, nil
}

func (s *Storage) parseHistogram(rows pgx.Rows, variable *model.Variable) (*api.Histogram, error) {

	termsAggName := api.TermsAggPrefix + variable.StorageName

	// extract the counts
	countMap := map[string]map[string]int64{}
	if rows != nil {
		for rows.Next() {
			var predictedTerm string
			var targetTerm string
			var bucketCount int64
			err := rows.Scan(&targetTerm, &predictedTerm, &bucketCount)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", termsAggName))
			}
			if len(countMap[predictedTerm]) == 0 {
				countMap[predictedTerm] = map[string]int64{}
			}
			countMap[predictedTerm][targetTerm] = bucketCount
		}
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}
	}

	correctBucket := &api.Bucket{
		Key: "Correct",
	}
	incorrectBucket := &api.Bucket{
		Key: "Incorrect",
	}

	for predictedKey, targetCounts := range countMap {
		for targetKey, count := range targetCounts {
			if predictedKey == targetKey {
				correctBucket.Count += count
			} else {
				incorrectBucket.Count += count
			}
		}
	}

	var min int64
	var max int64
	if incorrectBucket.Count < correctBucket.Count {
		min = incorrectBucket.Count
		max = correctBucket.Count
	} else {
		min = correctBucket.Count
		max = incorrectBucket.Count
	}

	// assign histogram attributes
	return &api.Histogram{
		Buckets: []*api.Bucket{
			correctBucket,
			incorrectBucket,
		},
		Extrema: &api.Extrema{
			Min: float64(min),
			Max: float64(max),
		},
	}, nil
}
