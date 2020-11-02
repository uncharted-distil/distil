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
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
	"goji.io/v3/pat"
)

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
		invert := pat.Param(r, "invert")
		invertBool := parseBoolParam(invert)

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

		ds, err := metaStore.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName

		// replace any grouped variables in filter params with the group's
		expandedFilterParams, err := api.ExpandFilterParams(dataset, filterParams, false, metaStore)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to expand filter params"))
			return
		}

		// fetch filtered data based on the supplied search parameters
		data, err := storage.FetchData(dataset, storageName, expandedFilterParams, invertBool)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable fetch filtered data"))
			return
		}

		// replace NaNs with an empty string to make them JSON encodable
		data = api.ReplaceNaNs(data, api.EmptyString)
		// marshal output into JSON
		bytes, err := json.Marshal(data)
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
