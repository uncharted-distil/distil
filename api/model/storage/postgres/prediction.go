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
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	api "github.com/uncharted-distil/distil/api/model"
	postgres "github.com/uncharted-distil/distil/api/postgres"
)

// PersistPrediction persists a prediction request to Postgres.
func (s *Storage) PersistPrediction(requestID string, dataset string, target string, fittedSolutionID string, progress string, createdTime time.Time) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, dataset, target, fitted_solution_id, progress, created_time, last_updated_time) VALUES ($1, $2, $3, $4, $5, $6, $6);", postgres.PredictionTableName)

	_, err := s.client.Exec(sql, requestID, dataset, target, fittedSolutionID, progress, createdTime)
	if err != nil {
		return errors.Wrapf(err, "failed to persist prediction request to PostGres")
	}
	return nil
}

// FetchPrediction pulls the specified prediction.
func (s *Storage) FetchPrediction(requestID string) (*api.Prediction, error) {
	sql := fmt.Sprintf("SELECT request_id, dataset, target, fitted_solution_id, progress, created_time, last_updated_time FROM %s "+
		"WHERE request_id = $1 ORDER BY created_time desc LIMIT 1;", postgres.PredictionTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}
	rows.Next()
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return s.loadPrediction(rows)
}

// FetchPredictionsByFittedSolutionID fetches all prediction requests using a given
// fitted solution id.
func (s *Storage) FetchPredictionsByFittedSolutionID(fittedSolutionID string) ([]*api.Prediction, error) {
	sql := fmt.Sprintf("SELECT request_id, dataset, target, fitted_solution_id, progress, created_time, last_updated_time FROM %s "+
		"WHERE fitted_solution_id = $1 ORDER BY created_time desc;", postgres.PredictionTableName)

	rows, err := s.client.Query(sql, fittedSolutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	predictions := []*api.Prediction{}
	for rows.Next() {
		prediction, err := s.loadPrediction(rows)
		if err != nil {
			return nil, err
		}
		predictions = append(predictions, prediction)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	return predictions, nil
}

// FetchPredictionResultByProduceRequestID pulls prediction result information from Postgres.  This is slightly different
// than FetchSolutionResultByProduceRequestID in that it joins with data from the `prediction` rather than the `request` table
// to extract additional information.
func (s *Storage) FetchPredictionResultByProduceRequestID(produceRequestID string) (*api.SolutionResult, error) {
	sql := fmt.Sprintf("SELECT result.solution_id, result.fitted_solution_id, result.produce_request_id, result.result_type, result.result_uuid, "+
		"result.result_uri, result.progress, result.created_time, prediction.dataset "+
		"FROM %s AS result INNER JOIN %s AS solution ON result.solution_id = solution.solution_id "+
		"INNER JOIN %s AS prediction ON result.produce_request_id = prediction.request_id "+
		"WHERE result.produce_request_id = $1;", postgres.SolutionResultTableName, postgres.SolutionTableName, postgres.PredictionTableName)

	rows, err := s.client.Query(sql, produceRequestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull prediction results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parseSolutionResult(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse predicdtion results from Postgres")
	}

	var res *api.SolutionResult
	if len(results) > 0 {
		res = results[0]
	}

	return res, nil
}

// FetchPredictionResultByUUID pulls solution result information from Postgres.
func (s *Storage) FetchPredictionResultByUUID(resultUUID string) (*api.SolutionResult, error) {
	sql := fmt.Sprintf("SELECT result.solution_id, result.fitted_solution_id, result.produce_request_id, result.result_type, result.result_uuid, "+
		"result.result_uri, result.progress, result.created_time, prediction.dataset "+
		"FROM %s AS result INNER JOIN %s AS solution ON result.solution_id = solution.solution_id "+
		"INNER JOIN %s AS prediction ON result.produce_request_id = prediction.request_id "+
		"WHERE result.result_uuid = $1;", postgres.SolutionResultTableName, postgres.SolutionTableName, postgres.PredictionTableName)

	rows, err := s.client.Query(sql, resultUUID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull prediction results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parseSolutionResult(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse prediction results from Postgres")
	}

	var res *api.SolutionResult
	if len(results) > 0 {
		res = results[0]
	}

	return res, nil
}

func (s *Storage) loadPrediction(rows pgx.Rows) (*api.Prediction, error) {
	var requestID string
	var dataset string
	var target string
	var fittedSolutionID string
	var progress string
	var createdTime time.Time
	var lastUpdatedTime time.Time

	err := rows.Scan(&requestID, &dataset, &target, &fittedSolutionID, &progress, &createdTime, &lastUpdatedTime)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse prediction request from Postgres")
	}

	return &api.Prediction{
		RequestID:        requestID,
		Dataset:          dataset,
		Target:           target,
		FittedSolutionID: fittedSolutionID,
		Progress:         progress,
		CreatedTime:      createdTime,
		LastUpdatedTime:  lastUpdatedTime,
	}, nil
}
