package postgres

import (
	"fmt"

	"github.com/pkg/errors"
	es "github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/postgres"
	elastic "gopkg.in/olivere/elastic.v5"
)

const (
	sessionTableName = "session"
	requestTableName = "request"
	resultTableName  = "result"
	featureTableName = "requestFeature"
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
func (s *Storage) PersistRequest(sessionID string, requestID string, pipelineID string, dataset string, progress string) error {
	// Insert the request.
	sql := fmt.Sprintf("INSERT INTO %s (session_id, request_id, pipeline_id, dataset, progress) VALUES ($1, $2, $3, $4, $5);", requestTableName)

	_, err := s.client.Exec(sql, sessionID, requestID, pipelineID, dataset, progress)

	return err
}

// UpdateRequest updates a request in Postgres.
func (s *Storage) UpdateRequest(requestID string, pipelineID string, progress string) error {
	// Update the request.
	sql := fmt.Sprintf("UPDATE %s SET progress = $1, pipeline_id = $2 WHERE request_id = $3;", requestTableName)

	_, err := s.client.Exec(sql, progress, pipelineID, requestID)

	return err
}

// PersistResultMetadata persists the result metadata to Postgres.
func (s *Storage) PersistResultMetadata(requestID string, resultUUID string, resultURI string, progress string) error {
	// Insert the result (metadata, not result data).
	sql := fmt.Sprintf("INSERT INTO %s (request_id, result_uuid, result_uri, progress) VALUES ($1, $2, $3, $4);", resultTableName)

	_, err := s.client.Exec(sql, requestID, resultUUID, resultURI, progress)

	return err
}

// PersistRequestFeature persists request feature information to Postgres.
func (s *Storage) PersistRequestFeature(requestID string, featureName string, featureType string) error {
	sql := fmt.Sprintf("INSERT INTO %s (request_id, feature_name, feature_type) VALUES ($1, $2, $3);", resultTableName)

	_, err := s.client.Exec(sql, requestID, featureName, featureType)

	return err
}

// FetchRequests pulls session request information from Postgres. NOTE: Not implemented!
func (s *Storage) FetchRequests(sessionID string) ([]*model.Request, error) {
	sql := fmt.Sprintf("SELECT session_id, request_id, pipeline_id, dataset, progress FROM %s WHERE session_id = $1;", requestTableName)

	rows, err := s.client.Query(sql, sessionID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull session requests from Postgres")
	}

	requests := make([]*model.Request, 0)
	for rows.Next() {
		var sessionID string
		var requestID string
		var pipelineID string
		var dataset string
		var progress string

		err = rows.Scan(&sessionID, &requestID, &pipelineID, &dataset, &progress)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse session requests from Postgres")
		}

		results, err := s.FetchResultMetadata(requestID)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get request results from Postgres")
		}

		requests = append(requests, &model.Request{
			SessionID:  sessionID,
			RequestID:  requestID,
			PipelineID: pipelineID,
			Dataset:    dataset,
			Progress:   progress,
			Results:    results,
		})
	}

	return requests, nil
}

// FetchResultMetadata pulls request result information from Psotgres.
func (s *Storage) FetchResultMetadata(requestID string) ([]*model.Result, error) {
	sql := fmt.Sprintf("SELECT request_id, result_uuid, result_uri, progress FROM %s WHERE request_id = $1;", resultTableName)

	rows, err := s.client.Query(sql, requestID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to pull request results from Postgres")
	}

	results := make([]*model.Result, 0)
	for rows.Next() {
		var requestID string
		var resultUUID string
		var resultURI string
		var progress string

		err = rows.Scan(&requestID, &resultUUID, &resultURI, &progress)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse requests results from Postgres")
		}

		results = append(results, &model.Result{
			RequestID:  requestID,
			ResultURI:  resultURI,
			ResultUUID: resultUUID,
			Progress:   progress,
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
