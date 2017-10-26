package elastic

import (
	es "github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
	elastic "gopkg.in/olivere/elastic.v5"
)

// Storage accesses the underlying ES instance.
type Storage struct {
	client *elastic.Client
}

// NewMetadataStorage returns a constructor for a metadata storage.
func NewMetadataStorage(clientCtor es.ClientCtor) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		esClient, err := clientCtor()
		if err != nil {
			return nil, err
		}

		return &Storage{
			client: esClient,
		}, nil
	}
}
