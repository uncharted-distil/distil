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
			if err := composeVariable(meta, data, dataset, grouping.IDCol, grouping.SubIDs); err != nil {
				handleError(w, errors.Wrapf(err, "unable to create new variable %s", grouping.IDCol))
				return
			}

			// ensure properties are typed correctly
			storageName := model.NormalizeDatasetID(dataset)

			// ensure id is timeseries
			err = setDataType(meta, data, dataset, storageName, grouping.IDCol, model.TimeSeriesType)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to update the data type in storage"))
				return
			}

			// For set the name of the expected cluster column - it doesn't necessarily exist.
			grouping.Properties.ClusterCol = fmt.Sprintf("%s%s", model.ClusterVarPrefix, grouping.IDCol)
		} else if model.IsGeoCoordinate(grouping.Type) {
			// make the lat column the id col for now since id col is what holds the info.
			grouping.IDCol = grouping.Properties.XCol
		}

		err = meta.AddGrouping(dataset, grouping)
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

		grouping := model.Grouping{}
		err = json.MapToStruct(&grouping, g)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse grouping parameter"))
			return
		}

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		err = meta.RemoveGrouping(dataset, grouping)
		if err != nil {
			handleError(w, err)
			return
		}

		// if there was a cluster var associated with this group and it has been created, remove it now
		// TODO: Should this be done explicitly through the client through some type of a
		// delete route?
		if grouping.Properties.ClusterCol != "" {
			clusterVarExist, err := meta.DoesVariableExist(dataset, grouping.Properties.ClusterCol)
			if err != nil {
				handleError(w, err)
				return
			}
			if clusterVarExist {
				err = meta.DeleteVariable(dataset, grouping.Properties.ClusterCol)
				if err != nil {
					handleError(w, err)
					return
				}
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

func composeVariable(metaStorage api.MetadataStorage, dataStorage api.DataStorage, dataset string, composedVarName string, sourceVarNames []string) error {
	composeExists, err := metaStorage.DoesVariableExist(dataset, composedVarName)
	if err != nil {
		return err
	}

	storageName := model.NormalizeDatasetID(dataset)
	if !composeExists {
		// create the new field
		err = metaStorage.AddVariable(dataset, composedVarName, "key", model.StringType, "grouping")
		if err != nil {
			return err
		}
		err = dataStorage.AddVariable(dataset, storageName, composedVarName, model.StringType)
		if err != nil {
			return err
		}
	}

	// Fetch data using the source names as the filter
	filter := &api.FilterParams{
		Variables: sourceVarNames,
	}
	rawData, err := dataStorage.FetchData(dataset, storageName, filter, false)
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
	err = dataStorage.UpdateVariableBatch(storageName, composedVarName, composedData)
	if err != nil {
		return err
	}

	return nil
}

//
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
