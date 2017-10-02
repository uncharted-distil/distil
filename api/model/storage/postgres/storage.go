package postgres

import (
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	es "github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/postgres"
	elastic "gopkg.in/olivere/elastic.v5"
)

const (
	sessionTableName     = "session"
	requestTableName     = "request"
	resultTableName      = "result"
	resultScoreTableName = "result_score"
	featureTableName     = "request_feature"
)

// Storage accesses the underlying postgres database.
type Storage struct {
	client   postgres.DatabaseDriver
	clientES *elastic.Client
}

// NewStorage returns a constructor for a Storage.
func NewStorage(clientCtor postgres.ClientCtor, clientESCtor es.ClientCtor) model.StorageCtor {
	return func() (model.Storage, error) {
		client, err := clientCtor()
		if err != nil {
			return nil, err
		}

		clientES, err := clientESCtor()
		if err != nil {
			return nil, err
		}

		return &Storage{
			client:   client,
			clientES: clientES,
		}, nil
	}
}

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
		var sessionID string
		var requestID string
		var dataset string
		var progress string
		var createdTime time.Time
		var lastUpdatedTime time.Time

		err = rows.Scan(&sessionID, &requestID, &dataset, &progress, &createdTime, &lastUpdatedTime)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse session requests from Postgres")
		}

		results, err := s.FetchResultMetadata(requestID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get request results from Postgres")
		}

		features, err := s.FetchRequestFeature(requestID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get request features from Postgres")
		}

		requests = append(requests, &model.Request{
			SessionID:       sessionID,
			RequestID:       requestID,
			Dataset:         dataset,
			Progress:        progress,
			CreatedTime:     createdTime,
			LastUpdatedTime: lastUpdatedTime,
			Results:         results,
			Features:        features,
		})
	}

	return requests, nil
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

// FetchRequestFeature pulls request feature information from Postgres.
func (s *Storage) FetchRequestFeature(requestID string) ([]*model.RequestFeature, error) {
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
