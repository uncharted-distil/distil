package datamart

import (
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/rest"
	"github.com/unchartedsoftware/distil/api/task"
)

const (
	nyuSearchFunction = "search"
	nyuGetFunction    = "download"
	isiSearchFunction = "new/search_data"
	isiGetFunction    = "new/materialize_data"
)

type parseSearchResult func(responseRaw []byte) ([]*model.Dataset, error)
type downloadDataset func(datamart *Storage, id string, uri string) (string, error)

// Storage accesses the underlying datamart instance.
type Storage struct {
	client         *rest.Client
	outputPath     string
	searchFunction string
	getFunction    string
	config         *task.IngestTaskConfig
	parser         parseSearchResult
	download       downloadDataset
}

// NewNYUMetadataStorage returns a constructor for an NYU datamart.
func NewNYUMetadataStorage(outputPath string, config *task.IngestTaskConfig, clientCtor rest.ClientCtor) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		return &Storage{
			client:         clientCtor(),
			outputPath:     outputPath,
			searchFunction: nyuSearchFunction,
			getFunction:    nyuGetFunction,
			config:         config,
			parser:         parseNYUSearchResult,
			download:       materializeNYUDataset,
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
			getFunction:    isiGetFunction,
			config:         config,
			parser:         parseISISearchResult,
			download:       materializeISIDataset,
		}, nil
	}
}
