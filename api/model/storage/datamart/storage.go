package datamart

import (
	"github.com/unchartedsoftware/distil/api/model"
)

// Storage accesses the underlying datamart instance.
type Storage struct {
	uri string
}

// NewMetadataStorage returns a constructor for a metadata storage.
func NewMetadataStorage(uri string) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		return &Storage{
			uri: uri,
		}, nil
	}
}
