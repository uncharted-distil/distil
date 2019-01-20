package datamart

import (
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/rest"
	"github.com/unchartedsoftware/distil/api/task"
)

// Storage accesses the underlying datamart instance.
type Storage struct {
	client     *rest.Client
	outputPath string
	config     *task.IngestTaskConfig
}

// NewMetadataStorage returns a constructor for a metadata storage.
func NewMetadataStorage(outputPath string, config *task.IngestTaskConfig, clientCtor rest.ClientCtor) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		return &Storage{
			client:     clientCtor(),
			outputPath: outputPath,
			config:     config,
		}, nil
	}
}
