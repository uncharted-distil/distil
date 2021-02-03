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
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
)

// OutlierResult represents a outlier response for a variable.
type OutlierResult struct {
	OutlierField string `json:"outlier"`
}

// OutlierDetectionHandler generates a route handler that enables outlier detection
// for either remote sensing or tabular data.
// Generates a variable and the creation of the new column to hold the cluster label.
// Return the name of the variable if the detection has run successfully.
func OutlierDetectionHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
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
		datasetMeta, err := metaStorage.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := datasetMeta.StorageName

		// find the outliers in the dataset
		outlierData, err := task.OutlierDetection(datasetMeta, variable)
		if err != nil {
			handleError(w, err)
			return
		}

		// create a name for the outlier variable
		outlierVarName := "_outlier"

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

		if !(outlierVarMetaExist && outlierVarExistData) {

			// add Variable to MetaData
			err = metaStorage.AddVariable(dataset, outlierVarName, "Outlier", model.CategoricalType, "data")
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
			for _, outlier := range outlierData {
				outlierBatch[outlier.D3MIndex] = outlier.Label
			}

			// update the batches
			err = dataStorage.UpdateVariableBatch(storageName, outlierVarName, outlierBatch)
			if err != nil {
				handleError(w, err)
				return
			}
		}

		// marshal output into JSON
		err = handleJSON(w, OutlierResult{
			OutlierField: outlierVarName,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal outlier variable name into JSON"))
			return
		}
	}
}
