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
func (s *Storage) UpdateRequest(requestID string, progress string) error {
	// Update the request.
	sql := fmt.Sprintf("UPDATE %s SET progress = $1 WHERE request_id = $2;", requestTableName)

	_, err := s.client.Exec(sql, progress, requestID)

	return err
}

// PersistResultMetadata persists the result metadata to Postgres.
func (s *Storage) PersistResultMetadata(requestID string, resultUUID string, resultURI string) error {
	// Insert the result (metadata, not result data).
	sql := fmt.Sprintf("INSERT INTO %s (request_id, result_uuid, result_uri) VALUES ($1, $2, $3);", resultTableName)

	_, err := s.client.Exec(sql, requestID, resultUUID, resultURI)

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

		requests = append(requests, &model.Request{
			SessionID:  sessionID,
			RequestID:  requestID,
			PipelineID: pipelineID,
			Dataset:    dataset,
			Progress:   progress,
		})
	}

	return requests, nil
}
