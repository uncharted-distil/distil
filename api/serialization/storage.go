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

package serialization

import (
	"path"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

const (
	schemaVersion = "4.0.0"
	license       = "Unknown"
)

var (
	csvStorage     = NewCSV()
	parquetStorage = NewParquet()
)

// Storage defines the base functions needed to store datasets to a backing
// storage for interactions with an auto ml server.
type Storage interface {
	ReadDataset(uri string) (*api.RawDataset, error)
	WriteDataset(uri string, data *api.RawDataset) error
	ReadData(uri string) ([][]string, error)
	WriteData(uri string, data [][]string) error
	ReadMetadata(uri string) (*model.Metadata, error)
	WriteMetadata(uri string, metadata *model.Metadata, extended bool, update bool) error
	ReadRawVariables(uri string) ([]string, error)
}

// GetStorage returns the storage to use based on URI.
func GetStorage(uri string) Storage {
	if path.Ext(uri) == ".parquet" {
		return parquetStorage
	}

	return csvStorage
}

// WriteData writes data to storage using the specified URI.
func WriteData(uri string, data [][]string) error {
	store := GetStorage(uri)
	return store.WriteData(uri, data)
}

// GetCSVStorage returns the instantiated csv storage.
func GetCSVStorage() Storage {
	return csvStorage
}

// GetParquetStorage returns the instantiated parquet storage.
func GetParquetStorage() Storage {
	return parquetStorage
}

// ReadData reads the metadata to find the main data reference, then reads that.
func ReadData(schemaPath string) ([][]string, error) {
	// metadata can be read by CSV storage
	meta, err := csvStorage.ReadMetadata(schemaPath)
	if err != nil {
		return nil, err
	}

	dataPath := model.GetResourcePath(schemaPath, meta.GetMainDataResource())
	storage := GetStorage(dataPath)
	data, err := storage.ReadData(dataPath)
	if err != nil {
		return nil, err
	}

	return data, nil
}
