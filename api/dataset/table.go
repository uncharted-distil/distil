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

package dataset

import (
	"bytes"
	"encoding/csv"
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"

	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
)

// Table represents a basic table dataset.
type Table struct {
	Dataset string     `json:"dataset"`
	CSVData [][]string `json:"csvData"`
}

// NewTableDataset creates a new table dataset from raw byte data, assuming csv.
func NewTableDataset(dataset string, rawData []byte, config *env.Config) (*Table, error) {
	reader := csv.NewReader(bytes.NewReader(rawData))
	csvData, err := reader.ReadAll()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read csv data")
	}
	return &Table{
		Dataset: dataset,
		CSVData: csvData,
	}, nil
}

// CreateDataset structures a raw csv file into a valid D3M dataset.
func (t *Table) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*api.RawDataset, error) {
	if datasetName == "" {
		datasetName = t.Dataset
	}
	dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)

	// create the raw dataset schema doc
	datasetID := model.NormalizeDatasetID(datasetName)
	meta := model.NewMetadata(datasetName, datasetName, "", datasetID)
	dr := model.NewDataResource(compute.DefaultResourceID, model.ResTypeRaw, map[string][]string{compute.D3MResourceFormat: {"csv"}})
	dr.ResPath = dataFilePath
	meta.DataResources = []*model.DataResource{dr}

	return &api.RawDataset{
		ID:       datasetID,
		Name:     datasetName,
		Data:     t.CSVData,
		Metadata: meta,
	}, nil
}
