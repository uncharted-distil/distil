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
		//grouping, err := parseGrouping(g)
		//grouping := model.Grouping{}
		//err = json.MapToStruct(&grouping, g)
		//if err != nil {
		//	handleError(w, errors.Wrap(err, "Unable to parse grouping parameter"))
		//	return
		//}

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

		groupingType := g["type"].(string)
		if model.IsTimeSeries(groupingType) {
			tsg, err := parseTimeseriesGrouping(g)
			if err != nil {
				handleError(w, err)
				return
			}

			if tsg.IDCol != "" {
				// Create a new variable and column for the time series key.
				if err := task.CreateComposedVariable(meta, data, dataset, tsg.IDCol, tsg.YCol, tsg.SubIDs); err != nil {
					handleError(w, errors.Wrapf(err, "unable to create new variable %s", tsg.IDCol))
					return
				}

				// Set the name of the expected cluster column - it doesn't necessarily exist.
				tsg.ClusterCol = model.ClusterVarPrefix + tsg.IDCol
			}

			// Create a new grouped variable for the time series.
			groupingVarName := strings.Join([]string{tsg.XCol, tsg.YCol}, task.DefaultSeparator)
			err = meta.AddGroupedVariable(dataset, groupingVarName, tsg.YCol, model.TimeSeriesType, model.VarDistilRoleGrouping, tsg)
			if err != nil {
				handleError(w, err)
				return
			}
		} else if model.IsGeoCoordinate(groupingType) {
			gcg, err := parseGeoCoordinateGrouping(g)
			if err != nil {
				handleError(w, err)
				return
			}

			// No key required in this case.
			groupingVarName := strings.Join([]string{gcg.XCol, gcg.YCol}, task.DefaultSeparator)
			err = meta.AddGroupedVariable(dataset, groupingVarName, "Geocoordinate", model.GeoCoordinateType, model.VarDistilRoleGrouping, gcg)
			if err != nil {
				handleError(w, err)
				return
			}
		} else if model.IsRemoteSensing(groupingType) {
			rsg, err := parseRemoteSensingGrouping(g)
			if err != nil {
				handleError(w, err)
				return
			}

			err = meta.AddGroupedVariable(dataset, rsg.IDCol+"_group", "Geocoordinate", model.RemoteSensingType, model.VarDistilRoleGrouping, rsg)
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

		if !variable.IsGrouping() {
			handleError(w, errors.Errorf("variable %s is not grouped", variableName))
			return
		}

		// If there was a cluster var associated with this group and it has been created, remove it now
		cg, ok := variable.Grouping.(model.ClusteredGrouping)
		if ok && cg.GetClusterCol() != "" {
			clusterVarExist, err := meta.DoesVariableExist(dataset, cg.GetClusterCol())
			if err != nil {
				handleError(w, err)
				return
			}
			if clusterVarExist {
				err = meta.DeleteVariable(dataset, cg.GetClusterCol())
				if err != nil {
					handleError(w, err)
					return
				}
			}
		}

		// If there was an ID col associated with this group that was built from SubIDs, delete it now
		if variable.Grouping.GetIDCol() != "" && variable.Grouping.GetSubIDs() != nil {
			err = meta.DeleteVariable(dataset, variable.Grouping.GetIDCol())
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

func parseTimeseriesGrouping(rawGrouping map[string]interface{}) (*model.TimeseriesGrouping, error) {
	grouping := &model.TimeseriesGrouping{}
	err := json.MapToStruct(grouping, rawGrouping)
	if err != nil {
		return nil, err
	}

	return grouping, nil
}

func parseGeoCoordinateGrouping(rawGrouping map[string]interface{}) (*model.GeoCoordinateGrouping, error) {
	grouping := &model.GeoCoordinateGrouping{}
	err := json.MapToStruct(grouping, rawGrouping)
	if err != nil {
		return nil, err
	}

	return grouping, nil
}

func parseRemoteSensingGrouping(rawGrouping map[string]interface{}) (*model.RemoteSensingGrouping, error) {
	grouping := &model.RemoteSensingGrouping{}
	err := json.MapToStruct(grouping, rawGrouping)
	if err != nil {
		return nil, err
	}

	return grouping, nil
}
