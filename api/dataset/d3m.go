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

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
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

	// read the dataset
	ds, err := serialization.ReadDataset(path.Join(d.DatasetPath, compute.D3MDataSchema))
	if err != nil {
		return nil, err
	}

	// update the id & name & storage name
	datasetID := model.NormalizeDatasetID(datasetName)
	ds.Metadata.Name = datasetName
	ds.Metadata.ID = datasetName
	ds.Metadata.StorageName = datasetID

	// update the non main data resources to be absolute paths
	mainDR := ds.Metadata.GetMainDataResource()
	for _, dr := range ds.Metadata.DataResources {
		if dr != mainDR {
			dr.ResPath = model.GetResourcePathFromFolder(d.DatasetPath, dr)
		}
	}

	ds.DefinitiveTypes = d.isFullySpecified(ds)

	return ds, nil
}

func (d *D3M) isFullySpecified(ds *api.RawDataset) bool {
	// fully specified means all variables are in the metadata, there are no
	// unknown types and there is at least one non string, non index type
	// (to avoid to case where everything was initialized to text)

	mainDR := ds.Metadata.GetMainDataResource()
	if len(ds.Data[0]) != len(mainDR.Variables) {
		return false
	}

	// find one non string and non index, and make sure no unknowns exist
	foundComplexType := false
	varMapIndex := map[int]*model.Variable{}
	for _, v := range mainDR.Variables {
		if v.Type == model.UnknownSchemaType {
			return true
		} else if !foundComplexType && !model.IsText(v.Type) && !model.IsIndexRole(v.SelectedRole) {
			foundComplexType = true
		}
		varMapIndex[v.Index] = v
	}
	if !foundComplexType {
		return false
	}

	// check the variable list against the header in the data
	for i, h := range ds.Data[0] {
		if varMapIndex[i] == nil || varMapIndex[i].HeaderName != h {
			return false
		}
	}

	return true
}
