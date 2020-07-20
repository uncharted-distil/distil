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
	"path"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
)

// D3M captures the needed information for a D3M dataset.
type D3M struct {
	DatasetName string
	DatasetPath string
}

// NewD3MDataset creates a new d3m dataset from a dataset folder.
func NewD3MDataset(datasetName string, datasetPath string) (*D3M, error) {

	return &D3M{
		DatasetName: datasetName,
		DatasetPath: datasetPath,
	}, nil
}

// CreateDataset processes the D3M dataset and updates it as needed to meet distil needs.
func (d *D3M) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*api.RawDataset, error) {
	if datasetName == "" {
		datasetName = d.DatasetName
	}

	// read the metadata
	meta, err := metadata.LoadMetadataFromOriginalSchema(path.Join(d.DatasetPath, compute.D3MDataSchema), false)
	if err != nil {
		return nil, err
	}

	// update the id & name & storage name
	datasetID := model.NormalizeDatasetID(datasetName)
	meta.Name = datasetName
	meta.ID = datasetName
	meta.StorageName = datasetID

	// read the data
	csvData, err := util.ReadCSVFile(path.Join(d.DatasetPath, meta.GetMainDataResource().ResPath), false)
	if err != nil {
		return nil, err
	}

	return &api.RawDataset{
		ID:       datasetID,
		Name:     datasetName,
		Data:     csvData,
		Metadata: meta,
	}, nil
}
