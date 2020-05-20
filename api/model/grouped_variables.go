package model

import (
	"fmt"

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
		if variable.Grouping != nil && variable.Grouping.GetHidden() != nil {
			hidden = append(hidden, variable.Grouping.GetHidden()...)
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
				variable.Grouping != nil {
				cluserCol, ok := GetClusterColFromGrouping(variable.Grouping)
				if ok && HasClusterData(dataset, cluserCol, metaStore) {
					updatedFilterParams.Highlight.Key = cluserCol
					break
				}
			}
		}
	}

	// create variable lookup
	varMap := make(map[string]*model.Variable)
	for _, variable := range variables {
		varMap[variable.Name] = variable
	}

	updatedFilterParams.Variables = []string{}
	for _, filterVar := range filterParams.Variables {
		if varMap[filterVar] != nil {
			variable := varMap[filterVar]
			if variable.Grouping != nil {
				componentVars := []string{}

				// Include X and Y col when not dealing with time series - time series data is fetched subsequently
				if !model.IsTimeSeries(variable.Type) {
					timeseriesGrouping := variable.Grouping.(*model.TimeseriesGrouping)
					componentVars = append(componentVars, timeseriesGrouping.XCol, timeseriesGrouping.YCol)
				}

				// include the grouping ID if present
				if variable.Grouping.GetIDCol() != "" {
					componentVars = append(componentVars, variable.Grouping.GetIDCol())
				}

				// include the grouping sub-ids if the ID is created from mutliple columns
				if len(variable.Grouping.GetSubIDs()) > 0 {
					componentVars = append(componentVars, variable.Grouping.GetSubIDs()...)
				}

				// filter out any hidden variables for timeseries
				for _, componentVarName := range componentVars {
					updatedFilterParams.AddVariable(componentVarName)
				}
			} else if model.IsImage(variable.Type) {
				// add the variable, and if it exists the cluster{
				updatedFilterParams.AddVariable(variable.Name)
				clusterField := fmt.Sprintf("_cluster_%s", variable.Name)
				if varMap[clusterField] != nil {
					updatedFilterParams.AddVariable(varMap[clusterField].Name)
				}
			} else {
				updatedFilterParams.AddVariable(variable.Name)
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

// GetClusterColFromGrouping return the cluster column if it exists in the group.
func GetClusterColFromGrouping(group model.BaseGrouping) (string, bool) {
	clustered, ok := group.(*model.TimeseriesGrouping)
	if ok {
		return clustered.ClusterCol, true
	}

	return "", false
}
