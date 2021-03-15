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
	"bytes"
	"net/http"

	"io/ioutil"
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"goji.io/v3/pat"
)

// ExtractHandler extracts a dataset from storage and writes it to disk.
func ExtractHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, config env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
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
		// get storage clients
		metaStorage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		dataStorage, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		// replace any grouped variables in filter params with the group's
		expandedFilterParams, err := api.ExpandFilterParams(dataset, filterParams, false, metaStorage)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to expand filter params"))
			return
		}
		_, datasetPath, err := task.ExportDataset(dataset, metaStorage, dataStorage, expandedFilterParams)
		if err != nil {
			handleError(w, err)
			return
		}
		streamCSV, err := ioutil.ReadFile(path.Join(datasetPath, compute.D3MDataFolder, compute.D3MLearningData))
		if err != nil {
			handleError(w, err)
			return
		}
		buffer := bytes.NewBuffer(streamCSV)

		w.Header().Set("Content-type", "application/csv")
		_, err = buffer.WriteTo(w)
		if err != nil {
			handleError(w, err)
			return
		}
	}
}
