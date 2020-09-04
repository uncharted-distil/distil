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

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	log "github.com/unchartedsoftware/plog"
)

// ClusteringResult represents a clustering response for a variable.
type ClusteringResult struct {
	ClusterField string `json:"cluster"`
}

// ClusteringHandler generates a route handler that enables clustering
// of a variable and the creation of the new column to hold the cluster label.
func ClusteringHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, config env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variable name
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

		ds, err := metaStorage.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName

		clusterVarName := fmt.Sprintf("%s%s", model.ClusterVarPrefix, variable)

		// check if the cluster variables exist
		clusterVarExist, err := metaStorage.DoesVariableExist(dataset, clusterVarName)
		if err != nil {
			handleError(w, err)
			return
		}

		// get the source dataset folder
		datasetMeta, err := metaStorage.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// create the new metadata and database variables
		if !clusterVarExist {
			// add data variable if needed
			clusterVarInStorage, err := dataStorage.DoesVariableExist(dataset, storageName, clusterVarName)
			if err != nil {
				log.Warnf("unable to check if cluster variable already exists: %v", err)
			}
			if !clusterVarInStorage {
				err = dataStorage.AddVariable(dataset, storageName, clusterVarName, model.CategoricalType)
				if err != nil {
					handleError(w, err)
					return
				}
			}

			// cluster data
			addMeta, clustered, err := task.Cluster(datasetMeta, variable, config.ClusteringKMeans)
			if err != nil {
				handleError(w, err)
				return
			}

			if addMeta {
				err = metaStorage.AddVariable(dataset, clusterVarName, "Pattern", model.CategoricalType, "metadata")
				if err != nil {
					handleError(w, err)
					return
				}
			}

			// build the data for batching
			clusteredData := make(map[string]string)
			for _, cluster := range clustered {
				clusteredData[cluster.D3MIndex] = cluster.Label
			}

			// update the batches
			err = dataStorage.UpdateVariableBatch(storageName, clusterVarName, clusteredData)
			if err != nil {
				handleError(w, err)
				return
			}
		}

		// marshal output into JSON
		err = handleJSON(w, ClusteringResult{
			ClusterField: clusterVarName,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal clustering result into JSON"))
			return
		}
	}
}
