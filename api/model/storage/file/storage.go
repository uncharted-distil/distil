package file

import (
	"github.com/uncharted-distil/distil/api/model"
)

// Storage accesses the underlying datamart instance.
type Storage struct {
	folder string
}

// NewMetadataStorage returns a constructor for a metadata storage.
func NewMetadataStorage(folder string) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		return &Storage{
			folder: folder,
		}, nil
	}
}
