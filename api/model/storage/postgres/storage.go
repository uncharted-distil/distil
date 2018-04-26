package postgres

import (
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/postgres"
)

const (
	requestTableName        = "request"
	pipelineTableName       = "pipeline"
	pipelineResultTableName = "pipeline_result"
	pipelineScoreTableName  = "pipeline_score"
	featureTableName        = "request_feature"
	filterTableName         = "request_filter"
	wordStemTableName       = "word_stem"

	// Database data types
	dataTypeText    = "TEXT"
	dataTypeDouble  = "double precision"
	dataTypeFloat   = "FLOAT8"
	dataTypeInteger = "INTEGER"
)

// Storage accesses the underlying postgres database.
type Storage struct {
	client   postgres.DatabaseDriver
	metadata model.MetadataStorage
}

// NewDataStorage returns a constructor for a data storage.
func NewDataStorage(clientCtor postgres.ClientCtor, metadataCtor model.MetadataStorageCtor) model.DataStorageCtor {
	return func() (model.DataStorage, error) {
		return newStorage(clientCtor, metadataCtor)
	}
}

// NewPipelineStorage returns a constructor for a pipeline storage.
func NewPipelineStorage(clientCtor postgres.ClientCtor, metadataCtor model.MetadataStorageCtor) model.PipelineStorageCtor {
	return func() (model.PipelineStorage, error) {
		return newStorage(clientCtor, metadataCtor)
	}
}

func newStorage(clientCtor postgres.ClientCtor, metadataCtor model.MetadataStorageCtor) (*Storage, error) {
	client, err := clientCtor()
	if err != nil {
		return nil, err
	}

	metadata, err := metadataCtor()
	if err != nil {
		return nil, err
	}

	return &Storage{
		client:   client,
		metadata: metadata,
	}, nil
}
