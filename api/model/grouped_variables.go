package model

import (
	"github.com/uncharted-distil/distil-compute/model"
	log "github.com/unchartedsoftware/plog"
)

// FetchSummaryVariables fetches the variable list from the data store and filters/expands it to contain only
// those variables that should be displayed to a client.
func FetchSummaryVariables(dataset string, metaStore MetadataStorage) ([]*model.Variable, error) {
	// fetch all variables from the dataset
	variables, err := metaStore.FetchVariables(dataset, false, true)
	if err != nil {
		return nil, err
	}

	// get the hidden list from any grouped variables
	hidden := []string{}
	for _, variable := range variables {
		if variable.Grouping != nil && variable.Grouping.Hidden != nil {
			hidden = append(hidden, variable.Grouping.Hidden...)
		}
	}

	// loop through the var list and drop any that should be hidden
	visibleVars := []*model.Variable{}
	for _, variable := range variables {
		if !isHidden(variable.Name, hidden) {
			visibleVars = append(visibleVars, variable)
		}
	}
	return visibleVars, nil
}

// FetchDatasetVariables fetches the variable list for a dataset, and removes any variables that don't have
// corresponding dataset values (grouped variables).
func FetchDatasetVariables(dataset string, metaStore MetadataStorage) ([]*model.Variable, error) {
	// fetch all variables from the dataset
	variables, err := metaStore.FetchVariables(dataset, true, true)
	if err != nil {
		return nil, err
	}

	// drop grouped variables - only their components are stored in DB
	retainedVariables := []*model.Variable{}
	for _, variable := range variables {
		if variable.Grouping == nil {
			retainedVariables = append(retainedVariables, variable)
		}
	}
	return retainedVariables, nil
}

// ExpandFilterParams examines filter parameters for grouped variables, and replaces them with their constituent components
// as necessary.
func ExpandFilterParams(dataset string, filterParams *FilterParams, metaStore MetadataStorage) (*FilterParams, error) {
	if filterParams == nil {
		return nil, nil
	}

	// Fetch all variables from the dataset
	variables, err := metaStore.FetchVariables(dataset, true, true)
	if err != nil {
		return nil, err
	}

	updatedFilterParams := filterParams.Clone()

	// Check if the highlight variable is a group variable, and if it has associated cluster data.
	// If it does, update the filter key to use the highlight column.
	if updatedFilterParams.Highlight != nil {
		for _, variable := range variables {
			if variable.Name == updatedFilterParams.Highlight.Key &&
				variable.Grouping != nil && HasClusterData(dataset, variable.Grouping.Properties.ClusterCol, metaStore) {
				updatedFilterParams.Highlight.Key = variable.Grouping.Properties.ClusterCol
				break
			}
		}
	}

	updatedFilterParams.Variables = []string{}
	for _, variable := range variables {
		for _, filterVar := range filterParams.Variables {
			if filterVar == variable.Name {
				if variable.Grouping != nil {
					componentVars := []string{}

					// Include X and Y col when not dealing with time series - time series data is fetched subsequently
					if !model.IsTimeSeries(variable.Type) {
						componentVars = append(componentVars, variable.Grouping.Properties.XCol, variable.Grouping.Properties.YCol)
					}

					// include the grouping ID if present
					if variable.Grouping.IDCol != "" {
						componentVars = append(componentVars, variable.Grouping.IDCol)
					}

					// include the grouping sub-ids if the ID is created from mutliple columns
					if variable.Grouping.SubIDs != nil && len(variable.Grouping.SubIDs) > 0 {
						componentVars = append(componentVars, variable.Grouping.SubIDs...)
					}

					// filter out any hidden variables for timeseries
					for _, componentVarName := range componentVars {
						updatedFilterParams.AddVariable(componentVarName)
					}
				} else {
					updatedFilterParams.AddVariable(variable.Name)
				}
			}
		}
	}

	return updatedFilterParams, nil
}

func isHidden(variableName string, hidden []string) bool {
	for _, hiddenVarName := range hidden {
		if variableName == hiddenVarName {
			return true
		}
	}
	return false
}

// HasClusterData checks to see if a grouped variable has associated cluster data available.  If the cluster
// info has not yet been computed (it can be a long running task) then this willl return false.
func HasClusterData(datasetName string, variableName string, metaStore MetadataStorage) bool {
	result, err := metaStore.DoesVariableExist(datasetName, variableName)
	if err != nil {
		log.Warn(err)
	}
	return result
}
