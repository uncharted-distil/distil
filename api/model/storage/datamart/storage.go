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

type searchQuery func(datamart *Storage, query *SearchQuery) ([]byte, error)
type parseSearchResult func(responseRaw []byte) ([]*model.Dataset, error)
type downloadDataset func(datamart *Storage, id string, uri string) (string, error)

// Storage accesses the underlying datamart instance.
type Storage struct {
	client      *rest.Client
	outputPath  string
	getFunction string
	config      *task.IngestTaskConfig
	search      searchQuery
	parse       parseSearchResult
	download    downloadDataset
}

// NewNYUMetadataStorage returns a constructor for an NYU datamart.
func NewNYUMetadataStorage(outputPath string, config *task.IngestTaskConfig, clientCtor rest.ClientCtor) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		return &Storage{
			client:      clientCtor(),
			outputPath:  outputPath,
			getFunction: nyuGetFunction,
			config:      config,
			search:      nyuSearch,
			parse:       parseNYUSearchResult,
			download:    materializeNYUDataset,
		}, nil
	}
}

// NewISIMetadataStorage returns a constructor for an ISI datamart.
func NewISIMetadataStorage(outputPath string, config *task.IngestTaskConfig, clientCtor rest.ClientCtor) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		return &Storage{
			client:      clientCtor(),
			outputPath:  outputPath,
			getFunction: isiGetFunction,
			config:      config,
			search:      isiSearch,
			parse:       parseISISearchResult,
			download:    materializeISIDataset,
		}, nil
	}
}
