package elastic

import (
	"errors"
	"time"

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
	return errors.New("ElasticSearch Pe	rsisSession not implemented")
}

// PersistRequest persists a request to ES. NOTE: Not implemented!
func (s *Storage) PersistRequest(sessionID string, requestID string, dataset string, progress string, createdTime time.Time) error {
	return errors.New("ElasticSearch PersistRequest not implemented")
}

// PersistRequestFeature persists request feature information to ES. NOTE: Not implemented!
func (s *Storage) PersistRequestFeature(requestID string, featureName string, featureType string) error {
	return errors.New("ElasticSearch PersistRequestFeature not implemented")
}

// UpdateRequest updates a request in ES. NOTE: Not implemented!
func (s *Storage) UpdateRequest(requestID string, progress string, updatedTime time.Time) error {
	return errors.New("ElasticSearch UpdateRequest not implemented")
}

// PersistResultMetadata persists the result metadata to ES. NOTE: Not implemented!
func (s *Storage) PersistResultMetadata(requestID string, pipelineID string, resultUUID string, resultURI string, progress string, outputType string, createdTime time.Time) error {
	return errors.New("ElasticSearch PersistResultMetadata not implemented")
}

// PersistResultScore persists the result score to ES. NOTE: Not implemented!
func (s *Storage) PersistResultScore(pipelineID string, metric string, score float64) error {
	return errors.New("ElasticSearch PersistResultScore not implemented")
}

// FetchRequests pulls session request information from ES. NOTE: Not implemented!
func (s *Storage) FetchRequests(sessionID string) ([]*model.Request, error) {
	return nil, errors.New("ElasticSearch FetchRequests not implemented")
}

// FetchResultMetadata pulls request result information from ES. NOTE: Not implemented!
func (s *Storage) FetchResultMetadata(requestID string) ([]*model.Result, error) {
	return nil, errors.New("ElasticSearch FetchResultMetadata not implemented")
}

// FetchResultMetadataByUUID pulls request result information from ES. NOTE: Not implemented!
func (s *Storage) FetchResultMetadataByUUID(resultUUID string) (*model.Result, error) {
	return nil, errors.New("ElasticSearch FetchResultMetadataByUUID not implemented")
}

// FetchResultScore pulls request result score from ES. NOTE: Not implemented!
func (s *Storage) FetchResultScore(pipelineID string) ([]*model.ResultScore, error) {
	return nil, errors.New("ElasticSearch FetchResultScore not implemented")
}

// FetchRequestFeature pulls request feature information from ES. NOTE: Not implemented!
func (s *Storage) FetchRequestFeature(requestID string) ([]*model.RequestFeature, error) {
	return nil, errors.New("ElasticSearch FetchRequestFeature not implemented")
}

// SetDataType updates the data type of the field in ES. NOTE: Not implemented!
func (s *Storage) SetDataType(dataset string, index string, field string, fieldType string) error {
	return errors.New("ElasticSearch SetDataType not implemented")
}
