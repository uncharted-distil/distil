package postgres

import (
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// PersistPipeline persists the pipeline to Postgres.
func (s *Storage) PersistPipeline(modelID string, pipelineID string, progress string, createdTime time.Time) error {
	sql := fmt.Sprintf("INSERT INTO %s (model_id, pipeline_id, progress, created_time) VALUES ($1, $2, $3, $4);", pipelineTableName)

	_, err := s.client.Exec(sql, modelID, pipelineID, progress, createdTime)

	return err
}

// PersistPipelineResult persists the pipeline result metadata to Postgres.
func (s *Storage) PersistPipelineResult(pipelineID string, resultUUID string, resultURI string, progress string, createdTime time.Time) error {
	sql := fmt.Sprintf("INSERT INTO %s (pipeline_id, result_uuid, result_uri, progress, created_time) VALUES ($1, $2, $3, $4, $5);", pipelineResultTableName)

	_, err := s.client.Exec(sql, pipelineID, resultUUID, resultURI, progress, createdTime)

	return err
}

// PersistPipelineScore persist the pipeline score to Postgres.
func (s *Storage) PersistPipelineScore(pipelineID string, metric string, score float64) error {
	sql := fmt.Sprintf("INSERT INTO %s (pipeline_id, metric, score) VALUES ($1, $2, $3);", pipelineScoreTableName)

	_, err := s.client.Exec(sql, pipelineID, metric, score)

	return err
}

// FetchPipeline pulls pipeline information from Postgres.
func (s *Storage) FetchPipeline(pipelineID string) (*model.Pipeline, error) {
	sql := fmt.Sprintf("SELECT model_id, pipelineID, progress, created_time FROM %s WHERE pipeline_id = $1;", pipelineTableName)

	rows, err := s.client.Query(sql, pipelineID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull pipeline from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}
	rows.Next()

	return s.parsePipeline(rows)
}

func (s *Storage) parsePipeline(rows *pgx.Rows) (*model.Pipeline, error) {
	var modelID string
	var pipelineID string
	var progress string
	var createdTime time.Time

	err := rows.Scan(&modelID, &pipelineID, &progress, &createdTime)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse pipeline from Postgres")
	}

	return &model.Pipeline{
		ModelID:     modelID,
		PipelineID:  pipelineID,
		Progress:    progress,
		CreatedTime: createdTime,
	}, nil
}

func (s *Storage) parsePipelineResult(rows *pgx.Rows) ([]*model.PipelineResult, error) {
	results := make([]*model.PipelineResult, 0)
	for rows.Next() {
		var pipelineID string
		var resultUUID string
		var resultURI string
		var progress string
		var createdTime time.Time
		var dataset string

		err := rows.Scan(&pipelineID, &resultUUID, &resultURI, &progress, &createdTime, &dataset)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse pipeline results from Postgres")
		}

		results = append(results, &model.PipelineResult{
			PipelineID:  pipelineID,
			ResultURI:   resultURI,
			ResultUUID:  resultUUID,
			Progress:    progress,
			CreatedTime: createdTime,
			Dataset:     dataset,
		})
	}

	// TODO: This should not be in the parsing code. The calling code
	// should be loading the required data.
	// for _, result := range results {
	// 	features, err := s.FetchmodelFeatures(result.RequestID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	result.Features = features
	// }
	//
	// for _, result := range results {
	// 	filters, err := s.FetchRequestFilters(result.RequestID, result.Features)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	result.Filters = filters
	// }

	return results, nil
}

// FetchPipelineResultByModelID pulls pipeline result information from Postgres.
func (s *Storage) FetchPipelineResultByModelID(modelID string) ([]*model.PipelineResult, error) {
	sql := fmt.Sprintf("SELECT result.pipeline_id, result.result_uuid, "+
		"result.result_uri, result.progress, result.created_time, model.dataset "+
		"FROM %s AS result INNER JOIN %s AS pipeline ON result.pipeline_id = pipeline.pipeline_id "+
		"INNER JOIN %s AS model ON pipeline.model_id = model.model_id "+
		"WHERE model.model_id = $1;", pipelineResultTableName, pipelineTableName, modelTableName)

	rows, err := s.client.Query(sql, modelID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull model results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	return s.parsePipelineResult(rows)
}

// FetchPipelineResult pulls pipeline result information from Postgres.
func (s *Storage) FetchPipelineResult(pipelineID string) (*model.PipelineResult, error) {
	sql := fmt.Sprintf("SELECT result.pipeline_id, result.result_uuid, "+
		"result.result_uri, result.progress, result.created_time, model.dataset "+
		"FROM %s AS result INNER JOIN %s AS pipeline ON result.pipeline_id = pipeline.pipeline_id "+
		"INNER JOIN %s AS model ON pipeline.model_id = model.model_id "+
		"WHERE result.pipeline_id = $1 "+
		"ORDER BY result.created_time desc LIMIT 1;", pipelineResultTableName, pipelineTableName, modelTableName)

	rows, err := s.client.Query(sql, pipelineID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull pipeline results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parsePipelineResult(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse pipeline results from Postgres")
	}

	var res *model.PipelineResult
	if results != nil && len(results) > 0 {
		res = results[0]
	}

	return res, nil
}

// FetchPipelineResultByUUID pulls pipeline result information from Postgres.
func (s *Storage) FetchPipelineResultByUUID(resultUUID string) (*model.PipelineResult, error) {
	sql := fmt.Sprintf("SELECT result.pipeline_id, result.result_uuid, "+
		"result.result_uri, result.progress, result.created_time, model.dataset "+
		"FROM %s AS result INNER JOIN %s AS pipeline ON result.pipeline_id = pipeline.pipeline_id "+
		"INNER JOIN %s AS model ON pipeline.model_id = model.model_id "+
		"WHERE result.result_uuid = $1;", pipelineResultTableName, pipelineTableName, modelTableName)

	rows, err := s.client.Query(sql, resultUUID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull pipeline results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parsePipelineResult(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse pipeline results from Postgres")
	}

	var res *model.PipelineResult
	if results != nil && len(results) > 0 {
		res = results[0]
	}

	return res, nil
}

// FetchPipelineResultByDatasetTarget pulls pipeline result information from
// Postgres. Only the latest result for each pipeline is fetched.
func (s *Storage) FetchPipelineResultByDatasetTarget(dataset string, target string, pipelineID string) ([]*model.PipelineResult, error) {

	// get the pipeline ids
	sql := fmt.Sprintf("SELECT DISTINCT result.pipeline_id "+
		"FROM %s model INNER JOIN %s mf ON model.model_id = mf.model_id "+
		"INNER JOIN %s pipeline ON model.model_id = pipeline.model_id "+
		"INNER JOIN %s result ON pipeline.pipeline_id = result.pipeline_id",
		modelTableName, featureTableName, pipelineTableName, pipelineResultTableName)
	params := make([]interface{}, 0)

	if dataset != "" {
		sql = fmt.Sprintf("%s AND model.dataset = $%d", sql, len(params)+1)
		params = append(params, dataset)
	}
	if target != "" {
		sql = fmt.Sprintf("%s AND mf.feature_name = $%d AND mf.feature_type = $%d", sql, len(params)+1, len(params)+2)
		params = append(params, target)
		params = append(params, model.FeatureTypeTarget)
	}
	if pipelineID != "" {
		sql = fmt.Sprintf("%s AND pipeline.pipeline_id = $%d", sql, len(params)+1)
		params = append(params, pipelineID)
	}

	sql = fmt.Sprintf("%s;", sql)
	rows, err := s.client.Query(sql, params...)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull pipeline ids from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results := make([]*model.PipelineResult, 0)
	for rows.Next() {
		var pipelineID string

		err = rows.Scan(&pipelineID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse pipeline id from Postgres")
		}

		res, err := s.FetchPipelineResult(pipelineID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse pipeline result from Postgres")
		}
		results = append(results, res)
	}

	return results, nil
}

// FetchPipelineScore pulls pipeline score from Postgres.
func (s *Storage) FetchPipelineScore(pipelineID string) ([]*model.PipelineScore, error) {
	sql := fmt.Sprintf("SELECT pipeline_id, metric, score FROM %s WHERE pipeline_id = $1;", pipelineScoreTableName)

	rows, err := s.client.Query(sql, pipelineID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull pipeline score from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results := make([]*model.PipelineScore, 0)
	for rows.Next() {
		var pipelineID string
		var metric string
		var score float64

		err = rows.Scan(&pipelineID, &metric, &score)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse result score from Postgres")
		}

		results = append(results, &model.PipelineScore{
			PipelineID: pipelineID,
			Metric:     metric,
			Score:      score,
		})
	}

	return results, nil
}
