package datamart

import (
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/rest"
	"github.com/unchartedsoftware/distil/api/task"
)

const (
	nyuSearchFunction = "search"
	isiSearchFunction = "new/search_data"
)

type parseSearchResult func(responseRaw []byte) ([]*model.Dataset, error)

// Storage accesses the underlying datamart instance.
type Storage struct {
	client         *rest.Client
	outputPath     string
	searchFunction string
	config         *task.IngestTaskConfig
	parser         parseSearchResult
}

// NewNYUMetadataStorage returns a constructor for an NYU datamart.
func NewNYUMetadataStorage(outputPath string, config *task.IngestTaskConfig, clientCtor rest.ClientCtor) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		return &Storage{
			client:         clientCtor(),
			outputPath:     outputPath,
			searchFunction: nyuSearchFunction,
			config:         config,
			parser:         parseNYUSearchResult,
		}, nil
	}
}

// NewISIMetadataStorage returns a constructor for an ISI datamart.
func NewISIMetadataStorage(outputPath string, config *task.IngestTaskConfig, clientCtor rest.ClientCtor) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		return &Storage{
			client:         clientCtor(),
			outputPath:     outputPath,
			searchFunction: isiSearchFunction,
			config:         config,
			parser:         parseISISearchResult,
		}, nil
	}
}
