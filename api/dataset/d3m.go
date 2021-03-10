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

package dataset

import (
	"path"

	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
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
func (d *D3M) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*serialization.RawDataset, error) {
	log.Infof("creating dataset from d3m dataset source")
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

func (d *D3M) isFullySpecified(ds *serialization.RawDataset) bool {
	// fully specified means all variables are in the metadata, there are no
	// unknown types and there is at least one non string, non index type
	// (to avoid to case where everything was initialized to text)

	mainDR := ds.Metadata.GetMainDataResource()
	if len(ds.Data[0]) != len(mainDR.Variables) {
		log.Infof("not every variable is specified in the metadata")
		return false
	}

	// find one non string and non index, and make sure no unknowns exist
	foundComplexType := false
	varMapIndex := map[int]*model.Variable{}
	for _, v := range mainDR.Variables {
		if v.Type == model.UnknownSchemaType {
			log.Infof("at least one variable is unknown type")
			return false
		} else if !foundComplexType && d.variableIsTyped(v) {
			foundComplexType = true
		}
		varMapIndex[v.Index] = v
	}
	if !foundComplexType {
		log.Infof("all variables are either an index or a string")
		return false
	}

	// check the variable list against the header in the data
	for i, h := range ds.Data[0] {
		if varMapIndex[i] == nil || varMapIndex[i].HeaderName != h {
			log.Infof("header in data file does not match metadata variable list (%s differs from %s at position %d)", h, varMapIndex[i].HeaderName, i)
			return false
		}
	}

	log.Infof("metadata is fully specified")
	return true
}

func (d *D3M) variableIsTyped(variable *model.Variable) bool {
	// a variable is typed if:
	// 						it isnt a string and not an index
	//						it is a string but references another resource
	if !model.IsText(variable.Type) && !model.IsIndexRole(variable.SelectedRole) {
		return true
	}

	return variable.RefersTo != nil
}

// GetDefinitiveTypes returns an empty list as definitive types.
func (d *D3M) GetDefinitiveTypes() []*model.Variable {
	return []*model.Variable{}
}
