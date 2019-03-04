package file

import (
	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/model"
)

// DoesVariableExist returns whether or not a variable exists.
func (s *Storage) DoesVariableExist(dataset string, varName string) (bool, error) {
	return false, errors.Errorf("Not implemented")
}

// FetchVariable returns the variable for the provided index, dataset, and variable.
func (s *Storage) FetchVariable(dataset string, varName string) (*model.Variable, error) {
	return nil, errors.Errorf("Not implemented")
}

// FetchVariableDisplay returns the display variable for the provided index, dataset, and variable.
func (s *Storage) FetchVariableDisplay(dataset string, varName string) (*model.Variable, error) {
	return nil, errors.Errorf("Not implemented")
}

// FetchVariables returns all the variables for the provided index and dataset.
func (s *Storage) FetchVariables(dataset string, includeIndex bool, includeMeta bool) ([]*model.Variable, error) {
	return nil, errors.Errorf("Not implemented")
}

// FetchVariablesDisplay returns all the display variables for the provided index and dataset.
func (s *Storage) FetchVariablesDisplay(dataset string) ([]*model.Variable, error) {
	return nil, errors.Errorf("Not implemented")
}
