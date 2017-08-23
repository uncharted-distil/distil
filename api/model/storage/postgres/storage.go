package postgres

import (
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/postgres"
)

// Storage accesses the underlying postgres database.
type Storage struct {
	client postgres.DatabaseDriver
}

// NewStorage returns a constructor for a Storage.
func NewStorage(clientCtor postgres.ClientCtor) model.StorageCtor {
	return func() (model.Storage, error) {
		client, err := clientCtor()
		if err != nil {
			return nil, err
		}

		return &Storage{
			client: client,
		}, nil
	}
}
