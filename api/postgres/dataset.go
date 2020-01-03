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
	"github.com/uncharted-distil/distil-compute/model"
)

// Dataset is a struct containing the metadata of a dataset being processed.
type Dataset struct {
	ID              string
	Name            string
	Description     string
	Variables       []*model.Variable
	variablesLookup map[string]bool
	insertBatch     []string
	insertArgs      []interface{}
}

// NewDataset creates a new dataset instance.
func NewDataset(id, name, description string, meta *model.Metadata) *Dataset {
	ds := &Dataset{
		ID:              id,
		Name:            name,
		Description:     description,
		variablesLookup: make(map[string]bool),
	}
	// NOTE: Can only support data in a single data resource for now.
	if meta != nil {
		ds.Variables = meta.DataResources[0].Variables
	}

	ds.ResetBatch()

	return ds
}

// ResetBatch clears the batch contents.
func (ds *Dataset) ResetBatch() {
	ds.insertBatch = make([]string, 0)
	ds.insertArgs = make([]interface{}, 0)
}

// HasVariable checks to see if a variable is already contained in the dataset.
func (ds *Dataset) HasVariable(variable *model.Variable) bool {
	return ds.variablesLookup[variable.Name]
}

// AddVariable adds a variable to the dataset.
func (ds *Dataset) AddVariable(variable *model.Variable) {
	ds.Variables = append(ds.Variables, variable)
	ds.variablesLookup[variable.Name] = true
}

// AddInsert adds an insert statement and parameters to the batch.
func (ds *Dataset) AddInsert(statement string, args []interface{}) {
	ds.insertBatch = append(ds.insertBatch, statement)
	ds.insertArgs = append(ds.insertArgs, args...)
}

// GetBatch returns the insert statement batch.
func (ds *Dataset) GetBatch() []string {
	return ds.insertBatch
}

// GetBatchSize gets the insert batch count.
func (ds *Dataset) GetBatchSize() int {
	return len(ds.insertBatch)
}

// GetBatchArgs returns the insert batch arguments.
func (ds *Dataset) GetBatchArgs() []interface{} {
	return ds.insertArgs
}
