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

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
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

		if grouping.Type == "timeseries" {
			// ensure properties are typed correctly

			storageName := model.NormalizeDatasetID(dataset)

			// ensure id is timeseries
			err = meta.SetDataType(dataset, grouping.IDCol, model.TimeSeriesType)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to update the data type in storage"))
				return
			}
			err = data.SetDataType(dataset, storageName, grouping.IDCol, model.TimeSeriesType)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to update the data type in storage"))
				return
			}

			if grouping.Properties.ClusterCol != "" {
				// ensure cluster is categorical
				err = meta.SetDataType(dataset, grouping.Properties.ClusterCol, model.CategoricalType)
				if err != nil {
					handleError(w, errors.Wrap(err, "unable to update the data type in storage"))
					return
				}
				err = data.SetDataType(dataset, storageName, grouping.Properties.ClusterCol, model.CategoricalType)
				if err != nil {
					handleError(w, errors.Wrap(err, "unable to update the data type in storage"))
					return
				}
			}

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
