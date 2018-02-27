package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// PersistSession persists a session to Postgres.
func (s *Storage) PersistSession(sessionID string) error {
	// Insert the session.
	sql := fmt.Sprintf("INSERT INTO %s (session_id) VALUES ($1);", sessionTableName)

	_, err := s.client.Exec(sql, sessionID)

	return err
}

// PersistRequest persists a request to Postgres.
func (s *Storage) PersistRequest(sessionID string, requestID string, dataset string, progress string, createdTime time.Time) error {
	// Insert the request.
	sql := fmt.Sprintf("INSERT INTO %s (session_id, request_id, dataset, progress, created_time, last_updated_time) VALUES ($1, $2, $3, $4, $5, $5);", requestTableName)

	_, err := s.client.Exec(sql, sessionID, requestID, dataset, progress, createdTime)

	return err
}

// UpdateRequest updates a request in Postgres.
func (s *Storage) UpdateRequest(requestID string, progress string, updatedTime time.Time) error {
	// Update the request.
	sql := fmt.Sprintf("UPDATE %s SET progress = $1, last_updated_time = $2 WHERE request_id = $3;", requestTableName)

	_, err := s.client.Exec(sql, progress, updatedTime, requestID)

	return err
}

// PersistResultMetadata persists the result metadata to Postgres.
func (s *Storage) PersistResultMetadata(requestID string, pipelineID string, resultUUID string, resultURI string, progress string, outputType string, createdTime time.Time) error {
	// Insert the result (metadata, not result data).
	sql := fmt.Sprintf("INSERT INTO %s (request_id, pipeline_id, result_uuid, result_uri, progress, output_type, created_time) VALUES ($1, $2, $3, $4, $5, $6, $7);", resultTableName)

	_, err := s.client.Exec(sql, requestID, pipelineID, resultUUID, resultURI, progress, outputType, createdTime)

	return err
}

// PersistResultScore persist the result score to Postgres.
func (s *Storage) PersistResultScore(pipelineID string, metric string, score float64) error {
	sql := fmt.Sprintf("INSERT INTO %s (pipeline_id, metric, score) VALUES ($1, $2, $3);", resultScoreTableName)

	_, err := s.client.Exec(sql, pipelineID, metric, score)

	return err
}

// PersistRequestFeature persists request feature information to Postgres.
func (s *Storage) PersistRequestFeature(requestID string, featureName string, featureType string) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, feature_name, feature_type) VALUES ($1, $2, $3);", featureTableName)

	_, err := s.client.Exec(sql, requestID, featureName, featureType)

	return err
}

// PersistRequestFilters persists request filters information to Postgres.
func (s *Storage) PersistRequestFilters(requestID string, filters *model.FilterParams) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, feature_name, filter_type, filter_min, filter_max, filter_categories) VALUES ($1, $2, $3, $4, $5, $6);", filterTableName)

	for _, filter := range filters.Filters {
		switch filter.Type {
		case model.NumericalFilter:
			_, err := s.client.Exec(sql, requestID, filter.Name, model.NumericalFilter, filter.Min, filter.Max, "")
			if err != nil {
				return err
			}
		case model.CategoricalFilter:
			_, err := s.client.Exec(sql, requestID, filter.Name, model.CategoricalFilter, 0, 0, strings.Join(filter.Categories, ","))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// FetchRequests pulls session request information from Postgres.
func (s *Storage) FetchRequests(sessionID string) ([]*model.Request, error) {
	sql := fmt.Sprintf("SELECT session_id, request_id, dataset, progress, created_time, last_updated_time FROM %s WHERE session_id = $1;", requestTableName)

	rows, err := s.client.Query(sql, sessionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull session requests from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	requests := make([]*model.Request, 0)
	for rows.Next() {
		request, err := s.loadRequest(rows)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to load request from Postgres")
		}

		requests = append(requests, request)
	}

	return requests, nil
}

// FetchRequest pulls request information from Postgres.
func (s *Storage) FetchRequest(requestID string) (*model.Request, error) {
	sql := fmt.Sprintf("SELECT session_id, request_id, dataset, progress, created_time, last_updated_time FROM %s WHERE request_id = $1;", requestTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}
	rows.Next()

	return s.loadRequest(rows)
}

func (s *Storage) loadRequest(rows *pgx.Rows) (*model.Request, error) {
	var sessionID string
	var requestID string
	var dataset string
	var progress string
	var createdTime time.Time
	var lastUpdatedTime time.Time

	err := rows.Scan(&sessionID, &requestID, &dataset, &progress, &createdTime, &lastUpdatedTime)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse session requests from Postgres")
	}

	results, err := s.FetchResultMetadata(requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get request results from Postgres")
	}

	features, err := s.FetchRequestFeatures(requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get request features from Postgres")
	}

	filters, err := s.FetchRequestFilters(requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get request filters from Postgres")
	}

	return &model.Request{
		SessionID:       sessionID,
		RequestID:       requestID,
		Dataset:         dataset,
		Progress:        progress,
		CreatedTime:     createdTime,
		LastUpdatedTime: lastUpdatedTime,
		Results:         results,
		Features:        features,
		Filters:         filters,
	}, nil
}

func (s *Storage) parseResultMetadata(rows *pgx.Rows) ([]*model.Result, error) {
	results := make([]*model.Result, 0)
	for rows.Next() {
		var requestID string
		var pipelineID string
		var resultUUID string
		var resultURI string
		var progress string
		var outputType string
		var createdTime time.Time

		err := rows.Scan(&requestID, &pipelineID, &resultUUID, &resultURI, &progress, &outputType, &createdTime)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse requests results from Postgres")
		}

		scores, err := s.FetchResultScore(pipelineID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get result scores from Postgres")
		}

		results = append(results, &model.Result{
			RequestID:   requestID,
			PipelineID:  pipelineID,
			ResultURI:   resultURI,
			ResultUUID:  resultUUID,
			Progress:    progress,
			OutputType:  outputType,
			CreatedTime: createdTime,
			Scores:      scores,
		})
	}

	for _, result := range results {
		features, err := s.FetchRequestFeatures(result.RequestID)
		if err != nil {
			return nil, err
		}
		result.Features = features
	}

	return results, nil
}

// FetchResultMetadata pulls request result information from Postgres.
func (s *Storage) FetchResultMetadata(requestID string) ([]*model.Result, error) {
	sql := fmt.Sprintf("SELECT request_id, pipeline_id, result_uuid, result_uri, progress, output_type, created_time FROM %s WHERE request_id = $1;", resultTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	return s.parseResultMetadata(rows)
}

// FetchResultMetadataByPipelineID pulls request result information from Postgres.
func (s *Storage) FetchResultMetadataByPipelineID(pipelineID string) (*model.Result, error) {
	sql := fmt.Sprintf("SELECT request_id, pipeline_id, result_uuid, result_uri, progress, output_type, created_time FROM %s WHERE pipeline_id = $1 ORDER BY created_time desc LIMIT 1;", resultTableName)

	rows, err := s.client.Query(sql, pipelineID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parseResultMetadata(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse request results from Postgres")
	}

	var res *model.Result
	if results != nil && len(results) > 0 {
		res = results[0]
	}

	return res, nil
}

// FetchResultMetadataByUUID pulls request result information from Postgres.
func (s *Storage) FetchResultMetadataByUUID(resultUUID string) (*model.Result, error) {
	sql := fmt.Sprintf("SELECT request_id, pipeline_id, result_uuid, result_uri, progress, output_type, created_time FROM %s WHERE result_uuid = $1;", resultTableName)

	rows, err := s.client.Query(sql, resultUUID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request results from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results, err := s.parseResultMetadata(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse request results from Postgres")
	}

	var res *model.Result
	if results != nil && len(results) > 0 {
		res = results[0]
	}

	return res, nil
}

// FetchResultMetadataByDatasetTarget pulls request result information from
// Postgres. Only the latest result for each pipeline is fetched.
func (s *Storage) FetchResultMetadataByDatasetTarget(sessionID string, dataset string, target string, pipelineID string) ([]*model.Result, error) {

	// get the pipeline ids
	sql := fmt.Sprintf(`SELECT DISTINCT result.pipeline_id
			FROM %s request INNER JOIN %s rf ON request.request_id = rf.request_id INNER JOIN %s result ON request.request_id = result.request_id
			WHERE request.session_id = $1 `, requestTableName, featureTableName, resultTableName)
	params := make([]interface{}, 0)
	params = append(params, sessionID)

	if dataset != "" {
		sql = fmt.Sprintf("%s AND request.dataset = $%d", sql, len(params))
		params = append(params, dataset)
	}
	if target != "" {
		sql = fmt.Sprintf("%s AND rf.feature_name = $%d AND rf.feature_type = $%d", sql, len(params), len(params)+1)
		params = append(params, target)
		params = append(params, model.FeatureTypeTarget)
	}
	if pipelineID != "" {
		sql = fmt.Sprintf("%s AND result.pipeline_id = $%d", sql, len(params))
		params = append(params, pipelineID)
	}

	sql = fmt.Sprintf("%s;", sql)
	rows, err := s.client.Query(sql, params...)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request pipeline ids from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results := make([]*model.Result, 0)
	for rows.Next() {
		var pipelineID string

		err = rows.Scan(&pipelineID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse pipeline id from Postgres")
		}

		res, err := s.FetchResultMetadataByPipelineID(pipelineID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse pipeline result from Postgres")
		}
		results = append(results, res)
	}

	return results, nil
}

// FetchResultScore pulls result score from Postgres.
func (s *Storage) FetchResultScore(pipelineID string) ([]*model.ResultScore, error) {
	sql := fmt.Sprintf("SELECT pipeline_id, metric, score FROM %s WHERE pipeline_id = $1;", resultScoreTableName)

	rows, err := s.client.Query(sql, pipelineID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull result score from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results := make([]*model.ResultScore, 0)
	for rows.Next() {
		var pipelineID string
		var metric string
		var score float64

		err = rows.Scan(&pipelineID, &metric, &score)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse result score from Postgres")
		}

		results = append(results, &model.ResultScore{
			PipelineID: pipelineID,
			Metric:     metric,
			Score:      score,
		})
	}

	return results, nil
}

// FetchRequestFeatures pulls request feature information from Postgres.
func (s *Storage) FetchRequestFeatures(requestID string) ([]*model.RequestFeature, error) {
	sql := fmt.Sprintf("SELECT request_id, feature_name, feature_type FROM %s WHERE request_id = $1;", featureTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request features from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	results := make([]*model.RequestFeature, 0)
	for rows.Next() {
		var requestID string
		var featureName string
		var featureType string

		err = rows.Scan(&requestID, &featureName, &featureType)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse requests features from Postgres")
		}

		results = append(results, &model.RequestFeature{
			RequestID:   requestID,
			FeatureName: featureName,
			FeatureType: featureType,
		})
	}

	return results, nil
}

// FetchRequestFilters pulls request filter information from Postgres.
func (s *Storage) FetchRequestFilters(requestID string) (*model.FilterParams, error) {
	sql := fmt.Sprintf("SELECT request_id, feature_name, filter_type, filter_min, filter_max, filter_categories FROM %s WHERE request_id = $1;", filterTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request filters from Postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	filters := &model.FilterParams{
		Size: model.DefaultFilterSize,
	}

	for rows.Next() {
		var requestID string
		var featureName string
		var filterType string
		var filterMin float64
		var filterMax float64
		var filterCategories string

		err = rows.Scan(&requestID, &featureName, &filterType, &filterMin, &filterMax, &filterCategories)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse requests filters from Postgres")
		}

		switch filterType {
		case model.CategoricalFilter:
			filters.Filters = append(filters.Filters, model.NewCategoricalFilter(
				featureName,
				strings.Split(filterCategories, ","),
			))
		case model.NumericalFilter:
			filters.Filters = append(filters.Filters, model.NewNumericalFilter(
				featureName,
				filterMin,
				filterMax,
			))
		}
	}

	return filters, nil
}
