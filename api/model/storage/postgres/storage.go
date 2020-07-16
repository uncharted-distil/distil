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
	"fmt"

	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/model"

	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/postgres"
)

// Storage accesses the underlying postgres database.
type Storage struct {
	client      postgres.DatabaseDriver
	batchClient postgres.DatabaseDriver
	metadata    api.MetadataStorage
}

// NewDataStorage returns a constructor for a data storage.
func NewDataStorage(clientCtor postgres.ClientCtor, batchClientCtor postgres.ClientCtor, metadataCtor api.MetadataStorageCtor) api.DataStorageCtor {
	return func() (api.DataStorage, error) {
		return newStorage(clientCtor, batchClientCtor, metadataCtor)
	}
}

// NewSolutionStorage returns a constructor for a solution storage.
func NewSolutionStorage(clientCtor postgres.ClientCtor, metadataCtor api.MetadataStorageCtor) api.SolutionStorageCtor {
	return func() (api.SolutionStorage, error) {
		return newStorage(clientCtor, nil, metadataCtor)
	}
}

func newStorage(clientCtor postgres.ClientCtor, batchClientCtor postgres.ClientCtor, metadataCtor api.MetadataStorageCtor) (*Storage, error) {
	client, err := clientCtor()
	if err != nil {
		return nil, err
	}

	var batchClient postgres.DatabaseDriver
	if batchClientCtor != nil {
		batchClient, err = batchClientCtor()
		if err != nil {
			return nil, err
		}
	}

	metadata, err := metadataCtor()
	if err != nil {
		return nil, err
	}

	return &Storage{
		client:      client,
		batchClient: batchClient,
		metadata:    metadata,
	}, nil
}

// GetStorageName returns a valid unique name to use for a given dataset name.
func (s *Storage) GetStorageName(dataset string) (string, error) {
	// format normalize the dataset
	storageName := model.NormalizeDatasetID(dataset)

	// get all database tables
	existingTables := map[string]bool{}
	rows, err := s.client.Query("select table_name from information_schema.tables;")
	if err != nil {
		return "", errors.Wrapf(err, "unable to get list of existing tables")
	}
	for rows.Next() {
		var existingName string
		err = rows.Scan(&existingName)
		if err != nil {
			return "", errors.Wrapf(err, "unable to scan table name")
		}
		existingTables[existingName] = true
	}
	err = rows.Err()
	if err != nil {
		return "", errors.Wrapf(err, "unable to read list of existing tables")
	}

	// get a unique value
	currentName := storageName
	for i := 1; existingTables[currentName]; i++ {
		currentName = fmt.Sprintf("%s_%d", storageName, i)
	}

	return currentName, nil
}
