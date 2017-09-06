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
func (s *Storage) PersistResult(dataset string, pipelineID string, resultURI string) error {
	return errors.New("ElasticSearch PersistResult not implemented")
}

// FetchResults pulls the pipeline result from ES. NOTE: Not implemented!
func (s *Storage) FetchResults(pipelineURI string, resultURI string, index string, dataset string, targetName string) (*model.FilteredData, error) {
	return nil, errors.New("ElasticSearch FetchResults not implemented")
}

// FetchResultsSummary pulls the pipeline result summary from ES. NOTE: Not implemented!
func (s *Storage) FetchResultsSummary(pipelineURI string, resultURI string, index string, dataset string, targetName string) (*model.Histogram, error) {
	return nil, errors.New("ElasticSearch FetchResultsSummary not implemented")
}
