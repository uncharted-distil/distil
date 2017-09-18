package elastic

import (
	"errors"

	es "github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
	elastic "gopkg.in/olivere/elastic.v5"
)

// Storage accesses the underlying ES instance
type Storage struct {
	client *elastic.Client
}

// NewStorage returns a constructor for an ES storage.
func NewStorage(clientCtor es.ClientCtor) model.StorageCtor {
	return func() (model.Storage, error) {
		esClient, err := clientCtor()
		if err != nil {
			return nil, err
		}

		return &Storage{
			client: esClient,
		}, nil
	}
}

// PersistResult persists a pipeline result to ES. NOTE: Not implemented!
func (s *Storage) PersistResult(dataset string, resultURI string) error {
	return errors.New("ElasticSearch PersistResult not implemented")
}

// PersistSession persists a session to ES. NOTE: Not implemented!
func (s *Storage) PersistSession(sessionID string) error {
	return errors.New("ElasticSearch PersisSession not implemented")
}

// PersistRequest persists a request to ES. NOTE: Not implemented!
func (s *Storage) PersistRequest(sessionID string, requestID string, pipelineID string, dataset string, progress string) error {
	return errors.New("ElasticSearch PersistRequest not implemented")
}

// UpdateRequest updates a request in ES. NOTE: Not implemented!
func (s *Storage) UpdateRequest(requestID string, pipelineID string, progress string) error {
	return errors.New("ElasticSearch UpdateRequest not implemented")
}

// PersistResultMetadata persists the result metadata to ES. NOTE: Not implemented!
func (s *Storage) PersistResultMetadata(requestID string, resultUUID string, resultURI string, progress string) error {
	return errors.New("ElasticSearch PersistResultMetadata not implemented")
}

// FetchRequests pulls session request information from ES. NOTE: Not implemented!
func (s *Storage) FetchRequests(sessionID string) ([]*model.Request, error) {
	return nil, errors.New("ElasticSearch FetchRequests not implemented")
}

// FetchResultMetadata pulls request result information from ES. NOTE: Not implemented!
func (s *Storage) FetchResultMetadata(requestID string) ([]*model.Result, error) {
	return nil, errors.New("ElasticSearch FetchResultMetadata not implemented")
}
