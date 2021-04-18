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

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
	"goji.io/v3/pat"
)

const (
	orderBy = "orderBy"
)

// FilteredDataClient is the structure the client requires when fetching data.
type FilteredDataClient struct {
	NumRows         int                        `json:"numRows"`
	NumRowsFiltered int                        `json:"numRowsFiltered"`
	Columns         []*api.Column              `json:"columns"`
	Values          [][]*api.FilteredDataValue `json:"values"`
}

// DataHandler creates a route that fetches filtered data from backing storage instance.
func DataHandler(storageCtor api.DataStorageCtor, metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		// get variable names and ranges out of the params
		filterParams, err := api.ParseFilterParamsFromJSON(params)
		if err != nil {
			handleError(w, err)
			return
		}

		dataset := pat.Param(r, "dataset")
		includeGroupingCol := pat.Param(r, "include-grouping-col")
		includeGroupingColBool := parseBoolParam(includeGroupingCol)
		var orderByVar *model.Variable = nil
		// get filter client
		storage, err := storageCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		metaStore, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		vars, err := metaStore.FetchVariables(dataset, true, true, true)
		if err != nil {
			handleError(w, err)
			return
		}
		if params[orderBy] != nil {
			for _, v := range vars {
				if v.HeaderName == params[orderBy] {
					orderByVar = v
					break
				}
			}
		}
		ds, err := metaStore.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName

		// replace any grouped variables in filter params with the group's
		expandedFilterParams, err := api.ExpandFilterParams(dataset, api.NewFilterParamsFromRaw(filterParams), false, metaStore)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to expand filter params"))
			return
		}

		// fetch filtered data based on the supplied search parameters
		data, err := storage.FetchData(dataset, storageName, expandedFilterParams, includeGroupingColBool, orderByVar)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable fetch filtered data"))
			return
		}

		// replace NaNs with an empty string to make them JSON encodable
		dataTransformed := transformDataForClient(data, api.EmptyString)

		// marshal output into JSON
		bytes, err := json.Marshal(dataTransformed)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal filtered data result into JSON"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(bytes)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal version into JSON and write response"))
			return
		}
	}
}

func transformDataForClient(data *api.FilteredData, replacementType api.NaNReplacement) *FilteredDataClient {
	data = api.ReplaceNaNs(data, replacementType)
	dataTransformed := &FilteredDataClient{
		NumRows:         data.NumRows,
		NumRowsFiltered: data.NumRowsFiltered,
		Values:          data.Values,
		Columns:         make([]*api.Column, len(data.Columns)),
	}

	for _, c := range data.Columns {
		dataTransformed.Columns[c.Index] = c
	}

	return dataTransformed
}
