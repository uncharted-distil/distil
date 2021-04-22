//
//   Copyright © 2021 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package datamart

import (
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/rest"
	"github.com/uncharted-distil/distil/api/task"
)

const (
	nyuSearchFunction       = "api/v1/search"
	nyuGetFunction          = "api/v1/download"
	isiSearchFunction       = "search"
	isiSearchFunctionNoData = "search_without_data"
	isiGetFunction          = "new/materialize_data"
)

type searchQuery func(datamart *Storage, query *SearchQuery, baseDataPath string) ([]byte, error)
type parseSearchResult func(responseRaw []byte, baseDataset *model.Dataset) ([]*model.Dataset, error)
type downloadDataset func(datamart *Storage, id string, uri string) (string, error)

// Storage accesses the underlying datamart instance.
type Storage struct {
	client         *rest.Client
	outputPath     string
	searchFunction string
	ingestConfig   *task.IngestTaskConfig
	config         *env.Config
	search         searchQuery
	parse          parseSearchResult
	download       downloadDataset
}

// NewNYUMetadataStorage returns a constructor for an NYU datamart.
func NewNYUMetadataStorage(outputPath string, config *env.Config, ingestConfig *task.IngestTaskConfig, clientCtor rest.ClientCtor) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		return &Storage{
			client:         clientCtor(),
			outputPath:     outputPath,
			searchFunction: nyuSearchFunction,
			config:         config,
			ingestConfig:   ingestConfig,
			search:         nyuSearch,
			parse:          parseNYUSearchResult,
			download:       materializeNYUDataset,
		}, nil
	}
}

// NewISIMetadataStorage returns a constructor for an ISI datamart.
func NewISIMetadataStorage(outputPath string, config *env.Config, ingestConfig *task.IngestTaskConfig, clientCtor rest.ClientCtor) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		return &Storage{
			client:         clientCtor(),
			outputPath:     outputPath,
			searchFunction: isiSearchFunction,
			config:         config,
			ingestConfig:   ingestConfig,
			search:         isiSearch,
			parse:          parseISISearchResult,
			download:       materializeISIDataset,
		}, nil
	}
}
