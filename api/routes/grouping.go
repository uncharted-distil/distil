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

package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
)

const (
	defaultSeparator = "_"
)

// GroupingHandler generates a route handler that adds a grouping.
func GroupingHandler(dataCtor api.DataStorageCtor, metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		dataset := pat.Param(r, "dataset")

		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		g, ok := json.Get(params, "grouping")
		if !ok {
			handleError(w, errors.Wrap(err, "Unable to parse grouping parameter"))
			return
		}

		// extract the grouping info
		grouping := model.Grouping{}
		err = json.MapToStruct(&grouping, g)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse grouping parameter"))
			return
		}

		data, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		if model.IsTimeSeries(grouping.Type) {
			// Create a new variable and column for the time series key.
			if err := createComposedVariable(meta, data, dataset, grouping.IDCol, grouping.SubIDs); err != nil {
				handleError(w, errors.Wrapf(err, "unable to create new variable %s", grouping.IDCol))
				return
			}

			// Set the name of the expected cluster column - it doesn't necessarily exist.
			grouping.Properties.ClusterCol = model.ClusterVarPrefix + grouping.IDCol

			// Create a new grouped variable for the time series.
			groupingVarName := strings.Join([]string{grouping.Properties.XCol, grouping.Properties.YCol}, defaultSeparator)
			err = meta.AddGroupedVariable(dataset, groupingVarName, "Count", model.TimeSeriesType, model.VarDistilRoleGrouping, grouping)
			if err != nil {
				handleError(w, err)
				return
			}
		} else if model.IsGeoCoordinate(grouping.Type) {
			// No key required in this case.
			groupingVarName := strings.Join([]string{grouping.Properties.XCol, grouping.Properties.YCol}, defaultSeparator)
			err = meta.AddGroupedVariable(dataset, groupingVarName, "Geocoordinate", model.GeoCoordinateType, model.VarDistilRoleGrouping, grouping)
			if err != nil {
				handleError(w, err)
				return
			}
		}

		// marshal data
		err = handleJSON(w, map[string]interface{}{
			"success": true,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal response into JSON"))
			return
		}
	}
}

// RemoveGroupingHandler generates a route handler that removes a grouping.
func RemoveGroupingHandler(dataCtor api.DataStorageCtor, metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// extract route parameters
		dataset := pat.Param(r, "dataset")
		variableName := pat.Param(r, "variable")

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		variable, err := meta.FetchVariable(dataset, variableName)
		if err != nil {
			handleError(w, err)
			return
		}

		if variable.Grouping == nil {
			handleError(w, errors.Errorf("variable %s is not grouped", variableName))
			return
		}

		// If there was a cluster var associated with this group and it has been created, remove it now
		if variable.Grouping.Properties.ClusterCol != "" {
			clusterVarExist, err := meta.DoesVariableExist(dataset, variable.Grouping.Properties.ClusterCol)
			if err != nil {
				handleError(w, err)
				return
			}
			if clusterVarExist {
				err = meta.DeleteVariable(dataset, variable.Grouping.Properties.ClusterCol)
				if err != nil {
					handleError(w, err)
					return
				}
			}
		}

		// If there was an ID col associated with this group that was built from SubIDs, delete it now
		if variable.Grouping.IDCol != "" && variable.Grouping.SubIDs != nil {
			err = meta.DeleteVariable(dataset, variable.Grouping.IDCol)
			if err != nil {
				handleError(w, err)
				return
			}
		}

		// Delete the gropuing variable itself
		err = meta.DeleteVariable(dataset, variableName)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data
		err = handleJSON(w, map[string]interface{}{
			"success": true,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal response into JSON"))
			return
		}
	}
}

func createComposedVariable(metaStorage api.MetadataStorage, dataStorage api.DataStorage, dataset string, composedVarName string, sourceVarNames []string) error {

	// create the variable metadata entry
	err := metaStorage.AddVariable(dataset, composedVarName, composedVarName, model.StringType, model.VarDistilRoleGrouping)
	if err != nil {
		return err
	}

	// create the variable data store entry
	datasetStorageName := model.NormalizeDatasetID(dataset)
	err = dataStorage.AddVariable(dataset, datasetStorageName, composedVarName, model.StringType)
	if err != nil {
		return err
	}

	// Fetch data using the source names as the filter
	filter := &api.FilterParams{
		Variables: sourceVarNames,
	}
	rawData, err := dataStorage.FetchData(dataset, datasetStorageName, filter, false)
	if err != nil {
		return err
	}

	// Create a map of the retreived fields to column number.  Store d3mIndex since it needs to be directly referenced
	// further along.
	d3mIndexFieldindex := -1
	colNameToIdx := make(map[string]int)
	for i, c := range rawData.Columns {
		if c.Label == model.D3MIndexName {
			d3mIndexFieldindex = i
		} else {
			colNameToIdx[c.Label] = i
		}
	}

	// Loop over the fetched data, composing each column value into a single new column value using the
	// separator.
	composedData := make(map[string]string)
	for _, r := range rawData.Values {
		// create the hash from the specified columns
		composed := createComposedFields(r, sourceVarNames, colNameToIdx, defaultSeparator)
		composedData[fmt.Sprintf("%v", r[d3mIndexFieldindex].Value)] = composed
	}

	// Save the new column
	err = dataStorage.UpdateVariableBatch(datasetStorageName, composedVarName, composedData)
	if err != nil {
		return err
	}

	return nil
}

func createComposedFields(data []*api.FilteredDataValue, fields []string, mappedFields map[string]int, separator string) string {
	dataToJoin := make([]string, len(fields))
	for i, field := range fields {
		dataToJoin[i] = fmt.Sprintf("%v", data[mappedFields[field]].Value)
	}
	return strings.Join(dataToJoin, separator)
}

func setDataType(esStorage api.MetadataStorage, pgStorage api.DataStorage,
	dataset string, storageName string, field string, typ string) error {
	err := esStorage.SetDataType(dataset, field, typ)
	if err != nil {
		return err
	}
	err = pgStorage.SetDataType(dataset, storageName, field, typ)
	if err != nil {
		return err
	}

	return nil
}
