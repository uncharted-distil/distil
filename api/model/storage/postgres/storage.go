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

package postgres

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	log "github.com/unchartedsoftware/plog"

	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/postgres"
)

const (
	// ProvenanceStorage label for storage valid types
	ProvenanceStorage = "storage-valid"
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

func (s *Storage) updateStats(storageName string) {
	_, err := s.client.Exec(fmt.Sprintf("ANALYZE \"%s\"", storageName))
	if err != nil {
		log.Warnf("error updating postgres stats for %s: %+v", storageName, err)
	}
}

// VerifyData checks each column in the table against every supported type, then updates what types are valid in the SuggestedType
func (s *Storage) VerifyData(datasetID string, tableName string) error {
	tableName = getBaseTableName(tableName)
	validTypes := postgres.GetValidTypes()
	ds, err := s.metadata.FetchDataset(datasetID, true, true, false)
	if err != nil {
		return err
	}
	//removing double and geometry for now
	double := "double precision"
	geometry := "geometry"
	mainValidTypes := []string{}
	for _, i := range validTypes {
		if i != double && i != geometry {
			mainValidTypes = append(mainValidTypes, i)
		}
	}

	// if view can succeed on column add potential type to column
	for _, v := range ds.Variables {
		// ignore index columns
		if model.IsIndexRole(v.SelectedRole) {
			continue
		}
		suggestedMap := make(map[string]bool)
		for _, t := range v.SuggestedTypes {
			suggestedMap[t.Type] = true
		}
		for _, j := range mainValidTypes {
			if postgres.IsColumnType(s.client, tableName, v, j) {
				d3mTypes, err := postgres.MapPostgresTypeToD3MType(j)
				if err != nil {
					continue
				}
				for _, k := range d3mTypes {
					// this could be moved up to an exit case above but a lot of the upconversion from pg to d3m types involves multiple results
					if suggestedMap[k] {
						continue
					}
					suggestedType := model.SuggestedType{Probability: 0, Type: k, Provenance: ProvenanceStorage}
					v.SuggestedTypes = append(v.SuggestedTypes, &suggestedType)
				}
			}
		}
	}
	// save changes err convert
	log.Infof("update metadata with complete list of suggested types")
	err = s.metadata.UpdateDataset(ds)
	if err != nil {
		return err
	}
	return nil
}
