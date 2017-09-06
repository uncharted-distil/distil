package postgres

import (
	es "github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/postgres"
	elastic "gopkg.in/olivere/elastic.v5"
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
