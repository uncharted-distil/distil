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

		ds, err := meta.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName

		groupingType := g["type"].(string)
		err = createGrouping(dataset, storageName, groupingType, g, meta, data)
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

		// If there was a new variable created that specifically has the grouping role, then delete it.
		groupingVarExists, err := meta.DoesVariableExist(dataset, variable.Grouping.GetIDCol())
		if err != nil {
			handleError(w, err)
			return
		}
		if groupingVarExists {
			idColVariable, err := meta.FetchVariable(dataset, variable.Grouping.GetIDCol())
			if err != nil {
				handleError(w, err)
				return
			}
			if idColVariable.DistilRole == model.VarDistilRoleGrouping {
				err = meta.DeleteVariable(dataset, variable.Grouping.GetIDCol())
				if err != nil {
					handleError(w, err)
					return
				}
			}
		}

		// Delete the grouping variable itself
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

func parseMultiBandImageGrouping(rawGrouping map[string]interface{}) (*model.MultiBandImageGrouping, error) {
	grouping := &model.MultiBandImageGrouping{}
	err := json.MapToStruct(grouping, rawGrouping)
	if err != nil {
		return nil, err
	}

	return grouping, nil
}

func createGrouping(dataset string, storageName string, groupingType string, rawGrouping map[string]interface{}, meta api.MetadataStorage, data api.DataStorage) error {
	if model.IsTimeSeries(groupingType) {
		tsg, err := parseTimeseriesGrouping(rawGrouping)
		if err != nil {
			return err
		}

		// Create a new variable and column for the time series key.
		if err := task.CreateComposedVariable(meta, data, dataset, storageName, tsg.IDCol, tsg.IDCol, tsg.SubIDs); err != nil {
			return errors.Wrapf(err, "unable to create new variable %s", tsg.IDCol)
		}

		// Set the name of the expected cluster column - it doesn't necessarily exist.
		tsg.ClusterCol = model.ClusterVarPrefix + tsg.IDCol

		// Create a new grouped variable for the time series.
		groupingVarName := strings.Join([]string{tsg.XCol, tsg.YCol}, task.DefaultSeparator)
		err = meta.AddGroupedVariable(dataset, groupingVarName, tsg.YCol, model.TimeSeriesType, model.VarDistilRoleData, tsg)
		if err != nil {
			return err
		}
	} else if model.IsGeoCoordinate(groupingType) {
		gcg, err := parseGeoCoordinateGrouping(rawGrouping)
		if err != nil {
			return err
		}

		// No key required in this case.
		groupingVarName := strings.Join([]string{gcg.XCol, gcg.YCol}, task.DefaultSeparator)
		err = meta.AddGroupedVariable(dataset, groupingVarName, "Geocoordinate", model.BivariateFilter, model.VarDistilRoleGrouping, gcg)
		if err != nil {
			return err
		}
	} else if model.IsMultiBandImage(groupingType) {
		rsg, err := parseMultiBandImageGrouping(rawGrouping)
		if err != nil {
			return err
		}

		// Set the name of the expected cluster column - it doesn't necessarily exist.
		varName := rsg.IDCol + "_group"
		rsg.ClusterCol = model.ClusterVarPrefix + rsg.IDCol
		err = meta.AddGroupedVariable(dataset, varName, "Tile", model.MultiBandImageType, model.VarDistilRoleGrouping, rsg)
		if err != nil {
			return err
		}
	} else {
		return errors.Errorf("unhandled group type %s", groupingType)
	}

	return nil
}
