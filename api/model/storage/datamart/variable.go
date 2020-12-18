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

package datamart

import (
	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/model"
)

// DoesVariableExist returns whether or not a variable exists.
func (s *Storage) DoesVariableExist(dataset string, varName string) (bool, error) {
	return false, errors.Errorf("Not implemented")
}

// FetchVariable returns the variable for the provided dataset, and variable.
func (s *Storage) FetchVariable(dataset string, varName string) (*model.Variable, error) {
	return nil, errors.Errorf("Not implemented")
}

// FetchVariableDisplay returns the display variable for the provided dataset, and variable.
func (s *Storage) FetchVariableDisplay(dataset string, varName string) (*model.Variable, error) {
	return nil, errors.Errorf("Not implemented")
}

// FetchVariables returns all the variables for the provided dataset.
func (s *Storage) FetchVariables(dataset string, includeIndex bool, includeMeta bool, includeSystemData bool) ([]*model.Variable, error) {
	return nil, errors.Errorf("Not implemented")
}

// FetchVariablesDisplay returns all the display variables for the provided dataset.
func (s *Storage) FetchVariablesDisplay(dataset string) ([]*model.Variable, error) {
	return nil, errors.Errorf("Not implemented")
}

// FetchVariablesByName returns all the variables for the provided dataset and names.
func (s *Storage) FetchVariablesByName(dataset string, variables []string, includeIndex bool, includeMeta bool, includeSystemData bool) ([]*model.Variable, error) {
	return nil, errors.Errorf("Not implemented")
}
