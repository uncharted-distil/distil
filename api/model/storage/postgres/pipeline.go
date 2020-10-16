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
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/model"
	postgres "github.com/uncharted-distil/distil/api/postgres"
)

// PersistSolution persists the solution to Postgres.
func (s *Storage) PersistSolution(requestID string, solutionID string, explainedSolutionID string, createdTime time.Time) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, solution_id, explained_solution_id, created_time) VALUES ($1, $2, $3, $4);", postgres.SolutionTableName)

	_, err := s.client.Exec(sql, requestID, solutionID, explainedSolutionID, createdTime)
	if err != nil {
		return errors.Wrap(err, "unable to persist solution")
	}

	return nil
}

// UpdateSolution updates the solution in Postgres.
func (s *Storage) UpdateSolution(solutionID string, explainedSolutionID string) error {
	sql := fmt.Sprintf("UPDATE %s SET explained_solution_id = $2 WHERE solution_id = $1;", postgres.SolutionTableName)

	_, err := s.client.Exec(sql, solutionID, explainedSolutionID)
	if err != nil {
		return errors.Wrap(err, "unable to update solution")
	}

	return nil
}

// PersistSolutionWeight persists the solution feature weight to Postgres.
func (s *Storage) PersistSolutionWeight(solutionID string, featureName string, featureIndex int64, weight float64) error {
	sql := fmt.Sprintf("INSERT INTO %s (solution_id, feature_name, feature_index, weight) VALUES ($1, $2, $3, $4);", postgres.SolutionFeatureWeightTableName)

	_, err := s.client.Exec(sql, solutionID, featureName, featureIndex, weight)

	return err
}

// PersistSolutionState persists the solution state to Postgres.
func (s *Storage) PersistSolutionState(solutionID string, progress string, createdTime time.Time) error {
	sql := fmt.Sprintf("INSERT INTO %s (solution_id, progress, created_time) VALUES ($1, $2, $3);", postgres.SolutionStateTableName)

	_, err := s.client.Exec(sql, solutionID, progress, createdTime)

	return err
}

// PersistSolutionResult persists the solution result metadata to Postgres.
func (s *Storage) PersistSolutionResult(solutionID string, fittedSolutionID string, produceRequestID string, resultType string,
	resultUUID string, resultURI string, progress string, explainOutput map[string]*api.SolutionExplainResult, createdTime time.Time) error {
	sql := fmt.Sprintf("INSERT INTO %s (solution_id, fitted_solution_id, produce_request_id, result_type, result_uuid, result_uri, progress, created_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);", postgres.SolutionResultTableName)

	_, err := s.client.Exec(sql, solutionID, fittedSolutionID, produceRequestID, resultType, resultUUID, resultURI, progress, createdTime)
	if err != nil {
		return errors.Wrap(err, "unable to persist solution result")
	}

	for typ, uri := range explainOutput {
		sql = fmt.Sprintf("INSERT INTO %s (result_id, explain_uri, explain_type) VALUES ($1, $2, $3)", postgres.SolutionResultExplainOutputTableName)
		_, err = s.client.Exec(sql, resultUUID, uri, typ)
		if err != nil {
			return errors.Wrap(err, "unable to persist solution result explain output")
		}
	}

	return err
}

// PersistSolutionScore persist the solution score to Postgres.
func (s *Storage) PersistSolutionScore(solutionID string, metric string, score float64) error {
	sql := fmt.Sprintf("INSERT INTO %s (solution_id, metric, score) VALUES ($1, $2, $3);", postgres.SolutionScoreTableName)

	_, err := s.client.Exec(sql, solutionID, metric, score)

	return err
}

// FetchSolution pulls solution information from Postgres.
func (s *Storage) FetchSolution(solutionID string) (*api.Solution, error) {
	sql := fmt.Sprintf("SELECT request_id, solution_id, explained_solution_id, created_time FROM %s WHERE solution_id = $1 ORDER BY created_time desc LIMIT 1;", postgres.SolutionTableName)

	rows, err := s.client.Query(sql, solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}
	rows.Next()
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	solution, err := s.parseSolution(rows)
	if err != nil {
		return nil, err
	}

	solution.IsBad = false
	return solution, nil
}

func (s *Storage) parseSolution(rows pgx.Rows) (*api.Solution, error) {
	var requestID string
	var solutionID string
	var explainedSolutionID string
	var createdTime time.Time

	err := rows.Scan(&requestID, &solutionID, &explainedSolutionID, &createdTime)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution from Postgres")
	}

	state, err := s.FetchSolutionState(solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution result from Postgres")
	}

	results, err := s.FetchSolutionResults(solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution result from Postgres")
	}

	scores, err := s.FetchSolutionScores(solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution scores from Postgres")
	}

	return &api.Solution{
		RequestID:           requestID,
		SolutionID:          solutionID,
		ExplainedSolutionID: explainedSolutionID,
		State:               state,
		CreatedTime:         createdTime,
		Results:             results,
		Scores:              scores,
	}, nil
}

func (s *Storage) parseSolutionWeight(rows pgx.Rows) ([]*api.SolutionWeight, error) {
	results := make([]*api.SolutionWeight, 0)
	for rows.Next() {
		var solutionID string
		var featureName string
		var featureIndex int64
		var weight float64

		err := rows.Scan(&solutionID, &featureName, &featureIndex, &weight)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse solution feature weight from Postgres")
		}

		results = append(results, &api.SolutionWeight{
			SolutionID:   solutionID,
			FeatureName:  featureName,
			FeatureIndex: featureIndex,
			Weight:       weight,
		})
	}
	err := rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return results, nil
}

// FetchSolutionWeights fetches solution feature weights from Postgres.
func (s *Storage) FetchSolutionWeights(solutionID string) ([]*api.SolutionWeight, error) {
	sql := fmt.Sprintf("SELECT solution_id, feature_name, feature_index, weight FROM %s WHERE solution_id = $1;", postgres.SolutionFeatureWeightTableName)

	rows, err := s.client.Query(sql, solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution feature weights from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parseSolutionWeight(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution feature weights from Postgres")
	}

	return results, nil
}

func (s *Storage) parseSolutionState(rows pgx.Rows) ([]*api.SolutionState, error) {
	results := make([]*api.SolutionState, 0)
	for rows.Next() {
		var solutionID string
		var progress string
		var createdTime time.Time

		err := rows.Scan(&solutionID, &progress, &createdTime)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse solution states from Postgres")
		}

		results = append(results, &api.SolutionState{
			SolutionID:  solutionID,
			Progress:    progress,
			CreatedTime: createdTime,
		})
	}
	err := rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return results, nil
}

func (s *Storage) parseSolutionResult(rows pgx.Rows) ([]*api.SolutionResult, error) {
	results := make([]*api.SolutionResult, 0)
	for rows.Next() {
		var solutionID string
		var fittedSolutionID string
		var produceRequestID string
		var resultType string
		var resultUUID string
		var resultURI string
		var progress string
		var createdTime time.Time
		var dataset string

		err := rows.Scan(&solutionID, &fittedSolutionID, &produceRequestID, &resultType, &resultUUID, &resultURI, &progress, &createdTime, &dataset)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse solution results from Postgres")
		}

		results = append(results, &api.SolutionResult{
			SolutionID:       solutionID,
			FittedSolutionID: fittedSolutionID,
			ProduceRequestID: produceRequestID,
			ResultType:       resultType,
			ResultURI:        resultURI,
			ResultUUID:       resultUUID,
			Progress:         progress,
			CreatedTime:      createdTime,
			Dataset:          dataset,
		})
	}
	err := rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return results, nil
}

func (s *Storage) parseSolutionFeatureWeight(resultURI string, rows pgx.Rows) (*api.SolutionFeatureWeight, error) {
	result := &api.SolutionFeatureWeight{
		ResultURI: resultURI,
	}

	if rows != nil {
		if rows.Next() {
			fields := rows.FieldDescriptions()
			columns := make([]string, len(fields))
			for i, f := range fields {
				columns[i] = string(f.Name)
			}

			columnValues, err := rows.Values()
			if err != nil {
				return nil, errors.Wrap(err, "Unable to extract fields from query result")
			}

			output := make(map[string]float64)
			for i := 0; i < len(columnValues); i++ {
				columnName := columns[i]
				if columnName == model.D3MIndexFieldName {
					result.D3MIndex = int64(columnValues[i].(float64))
				} else if columnName != "result_id" && columnValues[i] != nil {
					output[columnName] = columnValues[i].(float64)
				}
			}

			result.Weights = output
		}
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}
	}

	return result, nil
}

// FetchSolutionFeatureWeights fetches solution feature weights from Postgres.
func (s *Storage) FetchSolutionFeatureWeights(dataset string, storageName string, resultURI string, d3mIndex int64) (*api.SolutionFeatureWeight, error) {
	sql := fmt.Sprintf("SELECT * FROM %s WHERE result_id = $1 and \"d3mIndex\" = $2;",
		s.getSolutionFeatureWeightTable(storageName))

	rows, err := s.client.Query(sql, resultURI, d3mIndex)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution feature weights from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	result, err := s.parseSolutionFeatureWeight(resultURI, rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution feature weight from Postgres")
	}

	return result, nil
}

// FetchSolutionState pulls solution state information from Postgres.
func (s *Storage) FetchSolutionState(solutionID string) (*api.SolutionState, error) {
	sql := fmt.Sprintf("SELECT solution_id, progress, created_time "+
		"FROM %s AS state "+
		"WHERE state.solution_id = $1 "+
		"ORDER BY state.created_time desc LIMIT 1;", postgres.SolutionStateTableName)

	rows, err := s.client.Query(sql, solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution state from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parseSolutionState(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution state from Postgres")
	}

	var res *api.SolutionState
	if len(results) > 0 {
		res = results[0]
	}

	return res, nil
}

// FetchSolutionResults pulls solution result information from Postgres.
func (s *Storage) FetchSolutionResults(solutionID string) ([]*api.SolutionResult, error) {
	sql := fmt.Sprintf("SELECT result.solution_id, result.fitted_solution_id, result.produce_request_id, result.result_type, result.result_uuid, "+
		"result.result_uri, result.progress, result.created_time, request.dataset "+
		"FROM %s AS result INNER JOIN %s AS solution ON result.solution_id = solution.solution_id "+
		"INNER JOIN %s AS request ON solution.request_id = request.request_id "+
		"WHERE result.solution_id = $1 AND result.result_type = 'test' "+
		"ORDER BY result.created_time desc LIMIT 1;", postgres.SolutionResultTableName, postgres.SolutionTableName, postgres.RequestTableName)

	rows, err := s.client.Query(sql, solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parseSolutionResult(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution results from Postgres")
	}

	return results, nil
}

// FetchSolutionResultsByFittedSolutionID pulls solution result information from Postgres.
func (s *Storage) FetchSolutionResultsByFittedSolutionID(fittedSolutionID string) ([]*api.SolutionResult, error) {
	sql := fmt.Sprintf("SELECT result.solution_id, result.fitted_solution_id, result.produce_request_id, result.result_type, result.result_uuid, "+
		"result.result_uri, result.progress, result.created_time, request.dataset "+
		"FROM %s AS result INNER JOIN %s AS solution ON result.solution_id = solution.solution_id "+
		"INNER JOIN %s AS request ON solution.request_id = request.request_id "+
		"WHERE result.fitted_solution_id = $1 AND result.result_type = 'test' "+
		"ORDER BY result.created_time desc LIMIT 1;", postgres.SolutionResultTableName, postgres.SolutionTableName, postgres.RequestTableName)

	rows, err := s.client.Query(sql, fittedSolutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parseSolutionResult(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution results from Postgres")
	}

	return results, nil
}

// FetchSolutionResultByUUID pulls solution result information from Postgres.
func (s *Storage) FetchSolutionResultByUUID(resultUUID string) (*api.SolutionResult, error) {
	sql := fmt.Sprintf("SELECT result.solution_id, result.fitted_solution_id, result.produce_request_id, result.result_type, result.result_uuid, "+
		"result.result_uri, result.progress, result.created_time, request.dataset "+
		"FROM %s AS result INNER JOIN %s AS solution ON result.solution_id = solution.solution_id "+
		"INNER JOIN %s AS request ON solution.request_id = request.request_id "+
		"WHERE result.result_uuid = $1;", postgres.SolutionResultTableName, postgres.SolutionTableName, postgres.RequestTableName)

	rows, err := s.client.Query(sql, resultUUID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parseSolutionResult(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution results from Postgres")
	}

	// load the solution result explain output
	explained, err := s.fetchSolutionResultOutputExplain(resultUUID)
	if err != nil {
		return nil, err
	}

	var res *api.SolutionResult
	if len(results) > 0 {
		res = results[0]
		res.ExplainOutput = explained
	}

	return res, nil
}

func (s *Storage) fetchSolutionResultOutputExplain(resultID string) ([]*api.SolutionResultExplainOutput, error) {
	sql := fmt.Sprintf("SELECT result_id, explain_uri, explain_type FROM %s WHERE result_id = $1", postgres.SolutionResultExplainOutputTableName)

	rows, err := s.client.Query(sql, resultID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	return s.parseSolutionResultOutputExplain(rows)
}

func (s *Storage) parseSolutionResultOutputExplain(rows pgx.Rows) ([]*api.SolutionResultExplainOutput, error) {
	results := []*api.SolutionResultExplainOutput{}
	for rows.Next() {
		var resultID string
		var explainURI string
		var explainType string

		err := rows.Scan(&resultID, &explainURI, &explainType)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse solution results explain output from Postgres")
		}

		results = append(results, &api.SolutionResultExplainOutput{
			ResultID: resultID,
			URI:      explainURI,
			Type:     explainType,
		})
	}
	err := rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return results, nil
}

// FetchSolutionResultByProduceRequestID pulls solution result information from Postgres.
func (s *Storage) FetchSolutionResultByProduceRequestID(produceRequestID string) (*api.SolutionResult, error) {
	sql := fmt.Sprintf("SELECT result.solution_id, result.fitted_solution_id, result.produce_request_id, result.result_type, result.result_uuid, "+
		"result.result_uri, result.progress, result.created_time, request.dataset "+
		"FROM %s AS result INNER JOIN %s AS solution ON result.solution_id = solution.solution_id "+
		"INNER JOIN %s AS request ON solution.request_id = request.request_id "+
		"WHERE result.produce_request_id = $1;", postgres.SolutionResultTableName, postgres.SolutionTableName, postgres.RequestTableName)

	rows, err := s.client.Query(sql, produceRequestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parseSolutionResult(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution results from Postgres")
	}

	var res *api.SolutionResult
	if len(results) > 0 {
		res = results[0]
	}

	return res, nil
}

// FetchSolutionScores pulls solution score from Postgres.
func (s *Storage) FetchSolutionScores(solutionID string) ([]*api.SolutionScore, error) {
	sql := fmt.Sprintf("SELECT solution_id, metric, score FROM %s WHERE solution_id = $1;", postgres.SolutionScoreTableName)

	rows, err := s.client.Query(sql, solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution score from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	var results []*api.SolutionScore
	for rows.Next() {
		var solutionID string
		var metric string
		var score float64

		err = rows.Scan(&solutionID, &metric, &score)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse result score from Postgres")
		}

		results = append(results, &api.SolutionScore{
			SolutionID:     solutionID,
			Metric:         metric,
			Label:          compute.GetMetricLabel(metric),
			Score:          score,
			SortMultiplier: compute.GetMetricScoreMultiplier(metric),
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return results, nil
}

// FetchSolutionsByDatasetTarget fetches all solutions that apply to a particular dataset and target.
func (s *Storage) FetchSolutionsByDatasetTarget(dataset string, target string) ([]*api.Solution, error) {
	// get the solution ids
	sql := fmt.Sprintf("SELECT DISTINCT solution.solution_id "+
		"FROM %s request INNER JOIN %s rf ON request.request_id = rf.request_id "+
		"INNER JOIN %s solution ON request.request_id = solution.request_id ",
		postgres.RequestTableName, postgres.RequestFeatureTableName, postgres.SolutionTableName)
	params := make([]interface{}, 0)

	if dataset != "" {
		sql = fmt.Sprintf("%s AND request.dataset = $%d", sql, len(params)+1)
		params = append(params, dataset)
	}
	if target != "" {
		sql = fmt.Sprintf("%s AND rf.feature_name = $%d AND rf.feature_type = $%d", sql, len(params)+1, len(params)+2)
		params = append(params, target)
		params = append(params, model.FeatureTypeTarget)
	}

	sql = fmt.Sprintf("%s;", sql)
	rows, err := s.client.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	if rows != nil {
		defer rows.Close()
	}

	solutions := []*api.Solution{}
	for rows.Next() {
		var solutionID string

		err = rows.Scan(&solutionID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse solution id from Postgres")
		}

		solution, err := s.FetchSolution(solutionID)
		if err != nil {
			return nil, err
		}
		solutions = append(solutions, solution)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return solutions, nil
}

// FetchSolutionsByRequestID fetches solutions associated with a given request.
func (s *Storage) FetchSolutionsByRequestID(requestID string) ([]*api.Solution, error) {
	// get the solution ids
	sql := fmt.Sprintf("SELECT DISTINCT solution.solution_id "+
		"FROM %s request INNER JOIN %s rf ON request.request_id = rf.request_id "+
		"INNER JOIN %s solution ON request.request_id = solution.request_id "+
		"AND request.request_id = $1",
		postgres.RequestTableName, postgres.RequestFeatureTableName, postgres.SolutionTableName)

	params := []interface{}{requestID}
	sql = fmt.Sprintf("%s;", sql)
	rows, err := s.client.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	if rows != nil {
		defer rows.Close()
	}

	solutions := []*api.Solution{}
	for rows.Next() {
		var solutionID string

		err = rows.Scan(&solutionID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse solution id from Postgres")
		}

		solution, err := s.FetchSolution(solutionID)
		if err != nil {
			return nil, err
		}
		solutions = append(solutions, solution)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return solutions, nil
}
