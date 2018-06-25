package postgres

import (
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// PersistSolution persists the solution to Postgres.
func (s *Storage) PersistSolution(requestID string, solutionID string, progress string, createdTime time.Time) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, solution_id, progress, created_time) VALUES ($1, $2, $3, $4);", solutionTableName)

	_, err := s.client.Exec(sql, requestID, solutionID, progress, createdTime)

	return err
}

// PersistSolutionResult persists the solution result metadata to Postgres.
func (s *Storage) PersistSolutionResult(solutionID string, resultUUID string, resultURI string, progress string, createdTime time.Time) error {
	sql := fmt.Sprintf("INSERT INTO %s (solution_id, result_uuid, result_uri, progress, created_time) VALUES ($1, $2, $3, $4, $5);", solutionResultTableName)

	_, err := s.client.Exec(sql, solutionID, resultUUID, resultURI, progress, createdTime)

	return err
}

// PersistSolutionScore persist the solution score to Postgres.
func (s *Storage) PersistSolutionScore(solutionID string, metric string, score float64) error {
	sql := fmt.Sprintf("INSERT INTO %s (solution_id, metric, score) VALUES ($1, $2, $3);", solutionScoreTableName)

	_, err := s.client.Exec(sql, solutionID, metric, score)

	return err
}

// FetchSolution pulls solution information from Postgres.
func (s *Storage) FetchSolution(solutionID string) (*model.Solution, error) {
	sql := fmt.Sprintf("SELECT request_id, solution_id, progress, created_time FROM %s WHERE solution_id = $1 ORDER BY created_time desc LIMIT 1;", solutionTableName)

	rows, err := s.client.Query(sql, solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}
	rows.Next()

	return s.parseSolution(rows)
}

func (s *Storage) parseSolution(rows *pgx.Rows) (*model.Solution, error) {
	var requestID string
	var solutionID string
	var progress string
	var createdTime time.Time

	err := rows.Scan(&requestID, &solutionID, &progress, &createdTime)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution from Postgres")
	}

	result, err := s.FetchSolutionResult(solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution result from Postgres")
	}

	scores, err := s.FetchSolutionScores(solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse solution scores from Postgres")
	}

	return &model.Solution{
		RequestID:   requestID,
		SolutionID:  solutionID,
		Progress:    progress,
		CreatedTime: createdTime,
		Result:      result,
		Scores:      scores,
	}, nil
}

func (s *Storage) parseSolutionResult(rows *pgx.Rows) ([]*model.SolutionResult, error) {
	results := make([]*model.SolutionResult, 0)
	for rows.Next() {
		var solutionID string
		var resultUUID string
		var resultURI string
		var progress string
		var createdTime time.Time
		var dataset string

		err := rows.Scan(&solutionID, &resultUUID, &resultURI, &progress, &createdTime, &dataset)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse solution results from Postgres")
		}

		results = append(results, &model.SolutionResult{
			SolutionID:  solutionID,
			ResultURI:   resultURI,
			ResultUUID:  resultUUID,
			Progress:    progress,
			CreatedTime: createdTime,
			Dataset:     dataset,
		})
	}

	return results, nil
}

// FetchSolutionResult pulls solution result information from Postgres.
func (s *Storage) FetchSolutionResult(solutionID string) (*model.SolutionResult, error) {
	sql := fmt.Sprintf("SELECT result.solution_id, result.result_uuid, "+
		"result.result_uri, result.progress, result.created_time, request.dataset "+
		"FROM %s AS result INNER JOIN %s AS solution ON result.solution_id = solution.solution_id "+
		"INNER JOIN %s AS request ON solution.request_id = request.request_id "+
		"WHERE result.solution_id = $1 "+
		"ORDER BY result.created_time desc LIMIT 1;", solutionResultTableName, solutionTableName, requestTableName)

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

	var res *model.SolutionResult
	if results != nil && len(results) > 0 {
		res = results[0]
	}

	return res, nil
}

// FetchSolutionResultByUUID pulls solution result information from Postgres.
func (s *Storage) FetchSolutionResultByUUID(resultUUID string) (*model.SolutionResult, error) {
	sql := fmt.Sprintf("SELECT result.solution_id, result.result_uuid, "+
		"result.result_uri, result.progress, result.created_time, request.dataset "+
		"FROM %s AS result INNER JOIN %s AS solution ON result.solution_id = solution.solution_id "+
		"INNER JOIN %s AS request ON solution.request_id = request.request_id "+
		"WHERE result.result_uuid = $1;", solutionResultTableName, solutionTableName, requestTableName)

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

	var res *model.SolutionResult
	if results != nil && len(results) > 0 {
		res = results[0]
	}

	return res, nil
}

// FetchSolutionScores pulls solution score from Postgres.
func (s *Storage) FetchSolutionScores(solutionID string) ([]*model.SolutionScore, error) {
	sql := fmt.Sprintf("SELECT solution_id, metric, score FROM %s WHERE solution_id = $1;", solutionScoreTableName)

	rows, err := s.client.Query(sql, solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull solution score from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	var results []*model.SolutionScore
	for rows.Next() {
		var solutionID string
		var metric string
		var score float64

		err = rows.Scan(&solutionID, &metric, &score)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse result score from Postgres")
		}

		results = append(results, &model.SolutionScore{
			SolutionID: solutionID,
			Metric:     metric,
			Score:      score,
		})
	}

	return results, nil
}
