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
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util/json"
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

			if grouping.IDCol != "" {
				// Create a new variable and column for the time series key.
				if err := task.CreateComposedVariable(meta, data, dataset, grouping.IDCol, grouping.Properties.YCol, grouping.SubIDs); err != nil {
					handleError(w, errors.Wrapf(err, "unable to create new variable %s", grouping.IDCol))
					return
				}

				// Set the name of the expected cluster column - it doesn't necessarily exist.
				grouping.Properties.ClusterCol = model.ClusterVarPrefix + grouping.IDCol
			}

			// Create a new grouped variable for the time series.
			groupingVarName := strings.Join([]string{grouping.Properties.XCol, grouping.Properties.YCol}, task.DefaultSeparator)
			err = meta.AddGroupedVariable(dataset, groupingVarName, grouping.Properties.YCol, model.TimeSeriesType, model.VarDistilRoleGrouping, grouping)
			if err != nil {
				handleError(w, err)
				return
			}
		} else if model.IsGeoCoordinate(grouping.Type) {
			// No key required in this case.
			groupingVarName := strings.Join([]string{grouping.Properties.XCol, grouping.Properties.YCol}, task.DefaultSeparator)
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
