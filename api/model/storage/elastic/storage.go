//
//   Copyright © 2019 Uncharted Software Inc.
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

package elastic

import (
	es "github.com/uncharted-distil/distil/api/elastic"
	"github.com/uncharted-distil/distil/api/model"
	elastic "gopkg.in/olivere/elastic.v5"
)

// Storage accesses the underlying ES instance.
type Storage struct {
	client       *elastic.Client
	datasetIndex string
	modelIndex   string
}

// NewMetadataStorage returns a constructor for a metadata storage.
func NewMetadataStorage(datasetIndex string, clientCtor es.ClientCtor) model.MetadataStorageCtor {
	return func() (model.MetadataStorage, error) {
		esClient, err := clientCtor()
		if err != nil {
			return nil, err
		}

		return &Storage{
			client:       esClient,
			datasetIndex: datasetIndex,
		}, nil
	}
}

// NewExportedModelStorage returns a constructor for an exported model storage.
func NewExportedModelStorage(modelIndex string, clientCtor es.ClientCtor) model.ExportedModelStorageCtor {
	return func() (model.ExportedModelStorage, error) {
		esClient, err := clientCtor()
		if err != nil {
			return nil, err
		}

		return &Storage{
			client:     esClient,
			modelIndex: modelIndex,
		}, nil
	}
}
