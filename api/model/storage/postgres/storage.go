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

package postgres

import (
	"github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/postgres"
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

// NewSolutionStorage returns a constructor for a solution storage.
func NewSolutionStorage(clientCtor postgres.ClientCtor, metadataCtor model.MetadataStorageCtor) model.SolutionStorageCtor {
	return func() (model.SolutionStorage, error) {
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
