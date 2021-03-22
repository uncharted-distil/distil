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

package serialization

import (
	"path"

	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
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
	ReadDataset(uri string) (*RawDataset, error)
	WriteDataset(uri string, data *RawDataset) error
	ReadData(uri string) ([][]string, error)
	WriteData(uri string, data [][]string) error
	ReadMetadata(uri string) (*model.Metadata, error)
	WriteMetadata(uri string, metadata *model.Metadata, extended bool, update bool) error
	ReadRawVariables(uri string) ([]string, error)
}

// RawDataset contains basic information about the structure of the dataset as well
// as the raw learning data.
type RawDataset struct {
	ID              string
	Name            string
	Metadata        *model.Metadata
	Data            [][]string
	DefinitiveTypes bool
}

// SyncMetadata updates the key metadata properties to match a given metadata.
// This is often use to update the metadata for prediction or prefeaturization purposes.
func (d *RawDataset) SyncMetadata(metaToSync *model.Metadata) {
	d.Metadata.ID = metaToSync.ID
	d.Metadata.Name = metaToSync.Name
	d.Metadata.StorageName = metaToSync.StorageName
}

// AddField adds a field to the dataset, updating both the data and the metadata.
func (d *RawDataset) AddField(variable *model.Variable) error {
	if d.FieldExists(variable) {
		return errors.Errorf("field '%s' already exists in the raw dataset", variable.Key)
	}
	clone := variable.Clone()
	clone.Index = len(d.Metadata.GetMainDataResource().Variables)
	d.Metadata.GetMainDataResource().Variables = append(d.Metadata.GetMainDataResource().Variables, clone)

	// the first row is the header row
	d.Data[0] = append(d.Data[0], variable.HeaderName)
	for i, row := range d.Data[1:] {
		d.Data[i+1] = append(row, "")
	}

	return nil
}

// FieldExists returns true if a field is already part of the metadata.
func (d *RawDataset) FieldExists(variable *model.Variable) bool {
	for _, v := range d.Metadata.GetMainDataResource().Variables {
		if v.Key == variable.Key {
			return true
		}
	}

	return false
}

// GetVariableIndex returns the index of the variable as found in the header
// or -1 if not found in the header.
func (d *RawDataset) GetVariableIndex(variableHeaderName string) int {
	for i, f := range d.Data[0] {
		if f == variableHeaderName {
			return i
		}
	}

	return -1
}

// GetVariableIndices returns the mapping of variable header name to header index.
// It will error if a field is not found in the header.
func (d *RawDataset) GetVariableIndices(variableHeaderNames []string) (map[string]int, error) {
	indices := map[string]int{}
	for _, v := range variableHeaderNames {
		varIndex := d.GetVariableIndex(v)
		if varIndex == -1 {
			return nil, errors.Errorf("variable '%s' does not exist in header", v)
		}
		indices[v] = varIndex
	}

	return indices, nil
}

// FilterDataset updates the dataset to only keep the rows that have the specified
// column in the filter map set to true.
func (d *RawDataset) FilterDataset(filter map[string]bool) {
	if len(filter) == 0 {
		// clear the dataset since nothing is in filter set
		d.Data = [][]string{d.Data[0]}
		return
	}

	d3mIndexIndex := d.GetVariableIndex(model.D3MIndexFieldName)

	// start with the header
	filteredData := [][]string{d.Data[0]}
	for i := 1; i < len(d.Data); i++ {
		if filter[d.Data[i][d3mIndexIndex]] {
			filteredData = append(filteredData, d.Data[i])
		}
	}
	d.Data = filteredData
}

// UpdateDataset updates a dataset with the value specified in the updates dictionary.
// If the specified column value is not found in the dictionary, then it is left unchanged.
// Updates are specified by column index value.
func (d *RawDataset) UpdateDataset(updates map[int]map[string]string) {
	d3mIndexIndex := d.GetVariableIndex(model.D3MIndexFieldName)
	for i := 1; i < len(d.Data); i++ {
		d3mIndexValue := d.Data[i][d3mIndexIndex]
		for columnIndex, colUpdates := range updates {
			updateValue, ok := colUpdates[d3mIndexValue]
			if ok {
				d.Data[i][columnIndex] = updateValue
			}
		}
	}
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

// WriteMetadata writes the metadata to disk.
func WriteMetadata(uri string, metadata *model.Metadata) error {
	store := GetStorage(metadata.GetMainDataResource().ResPath)
	return store.WriteMetadata(uri, metadata, true, true)
}

// GetCSVStorage returns the instantiated csv storage.
func GetCSVStorage() Storage {
	return csvStorage
}

// GetParquetStorage returns the instantiated parquet storage.
func GetParquetStorage() Storage {
	return parquetStorage
}

// ReadDataset reads the metadata to find the main data reference, then reads that.
func ReadDataset(schemaPath string) (*RawDataset, error) {
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaPath, false)
	if err != nil {
		return nil, err
	}

	dataPath := model.GetResourcePath(schemaPath, meta.GetMainDataResource())
	return GetStorage(dataPath).ReadDataset(schemaPath)
}

// WriteDataset determines which storage engine to use and then writes out the
// metadata and the data using it.
func WriteDataset(folderPath string, dataset *RawDataset) error {
	// use the main data resource to determine the storage engine
	storage := GetStorage(dataset.Metadata.GetMainDataResource().ResPath)

	return storage.WriteDataset(folderPath, dataset)
}

// ReadMetadata reads the metadata in the specified path.
func ReadMetadata(schemaPath string) (*model.Metadata, error) {
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaPath, false)
	if err != nil {
		return nil, err
	}

	dataPath := model.GetResourcePath(schemaPath, meta.GetMainDataResource())
	return GetStorage(dataPath).ReadMetadata(schemaPath)
}
