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
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"goji.io/v3/pat"
)

// OutlierOutput represents a outlier response for a variable.
type OutlierOutput struct {
	Success bool `json:"success"`
}

const (
	outlierVarName     = "_outlier"
	outlierDisplayName = "Outlier"
)

// OutlierDetectionHandler generates a route handler that enables outlier detection for either
// remote sensing or tabular data. Return a boolean if the detection was successful.
func OutlierDetectionHandler(metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		config, err := env.LoadConfig()
		if err != nil {
			handleError(w, err)
			return
		}
		if !config.OutlierDetectionEnabled {
			err := handleJSON(w, OutlierOutput{
				Success: false,
			})
			if err != nil {
				handleError(w, errors.Wrap(err, "unable marshal outlier variable name into JSON"))
				return
			}
			return
		}
		dataset := pat.Param(r, "dataset")
		variable := pat.Param(r, "variable")

		// get storage clients
		metaStorage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// get the metadata
		datasetMeta, err := metaStorage.FetchDataset(dataset, true, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// find the outliers in the dataset
		points, err := task.OutlierDetection(datasetMeta, variable)
		if err != nil {
			handleError(w, err)
			return
		}

		// create an output
		output := OutlierOutput{
			Success: points != nil,
		}

		// marshal output into JSON
		err = handleJSON(w, output)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to marshal outlier detection results into a JSON"))
			return
		}
	}
}

// OutlierResultsHandler generates a route handler that enables outlier detection for either
// remote sensing or tabular data. Return a boolean and add the data to the dataset.
func OutlierResultsHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		variable := pat.Param(r, "variable")

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

		// get the metadata
		datasetMeta, err := metaStorage.FetchDataset(dataset, true, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := datasetMeta.StorageName

		// find the outliers in the dataset
		points, err := task.OutlierDetection(datasetMeta, variable)
		if err != nil {
			handleError(w, err)
			return
		}

		// check if the outlier variable exist in the metadata
		outlierVarMetaExist, err := metaStorage.DoesVariableExist(dataset, outlierVarName)
		if err != nil {
			handleError(w, err)
			return
		}

		// check if the outlier variable exist in the database
		outlierVarExistData, err := dataStorage.DoesVariableExist(dataset, storageName, outlierVarName)
		if err != nil {
			handleError(w, err)
			return
		}

		// create an output
		output := OutlierOutput{
			Success: false,
		}

		if !(outlierVarMetaExist && outlierVarExistData) {

			// add Variable to MetaData
			err = metaStorage.AddVariable(dataset, outlierVarName, outlierDisplayName, model.CategoricalType, []string{model.VarDistilRoleAugmented})
			if err != nil {
				handleError(w, err)
				return
			}
			// fetch the new outlier variable
			outlierVar, err := metaStorage.FetchVariable(dataset, outlierVarName)
			if err != nil {
				handleError(w, err)
				return
			}
			// add the values to the new outlier variable
			outlierVar.Values = []string{task.OutlierAnomaly, task.OutlierRegular}
			// update outlier variable
			err = metaStorage.UpdateVariable(dataset, outlierVarName, outlierVar)
			if err != nil {
				handleError(w, err)
				return
			}
			// add Variable to Database
			err = dataStorage.AddVariable(dataset, storageName, outlierVarName, model.CategoricalType, "")
			if err != nil {
				handleError(w, err)
				return
			}

			// build the data for batching
			outlierBatch := make(map[string]string)
			for _, point := range points {
				outlierBatch[point.D3MIndex] = point.Label
			}

			// update the batches
			err = dataStorage.UpdateVariableBatch(storageName, outlierVarName, outlierBatch)
			if err != nil {
				handleError(w, err)
				return
			}

			output = OutlierOutput{
				Success: true,
			}
		}

		// marshal output into JSON
		err = handleJSON(w, output)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to marshal outlier detection results into a JSON"))
			return
		}
	}
}
