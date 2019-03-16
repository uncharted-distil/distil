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
	"goji.io/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
)

// GeocodingResult represents a geocoding response for a variable.
type GeocodingResult struct {
	LatitudeField  string `json:"latitude"`
	LongitudeField string `json:"longitude"`
}

// GeocodingHandler generates a route handler that enables geocoding
// of a variable and the creation of two new columns to hold the lat and lon.
func GeocodingHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, sourceFolder string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		storageName := model.NormalizeDatasetID(dataset)
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
		latVarName := fmt.Sprintf("_lat_%s", variable)
		lonVarName := fmt.Sprintf("_lon_%s", variable)

		// check if the lat and lon variables exist
		latVarExist, err := metaStorage.DoesVariableExist(dataset, latVarName)
		if err != nil {
			handleError(w, err)
			return
		}
		lonVarExist, err := metaStorage.DoesVariableExist(dataset, lonVarName)
		if err != nil {
			handleError(w, err)
			return
		}

		// create the new metadata and database variables
		if !latVarExist {
			err = metaStorage.AddVariable(dataset, latVarName, model.LatitudeType, "geocoding")
			if err != nil {
				handleError(w, err)
				return
			}
			err = dataStorage.AddVariable(dataset, storageName, latVarName, model.LatitudeType)
			if err != nil {
				handleError(w, err)
				return
			}
		}
		if !lonVarExist {
			err = metaStorage.AddVariable(dataset, lonVarName, model.LongitudeType, "geocoding")
			if err != nil {
				handleError(w, err)
				return
			}
			err = dataStorage.AddVariable(dataset, storageName, lonVarName, model.LongitudeType)
			if err != nil {
				handleError(w, err)
				return
			}
		}

		// geocode data
		geocoded, err := task.GeocodeForward(sourceFolder, dataset, variable)
		if err != nil {
			handleError(w, err)
			return
		}

		// build the data for batching
		latData := make(map[string]string)
		lonData := make(map[string]string)
		for _, point := range geocoded {
			latData[point.D3MIndex] = fmt.Sprintf("%s", point.Latitude)
			lonData[point.D3MIndex] = fmt.Sprintf("%s", point.Longitude)
		}

		// update the batches
		err = dataStorage.UpdateVariableBatch(storageName, latVarName, latData)
		if err != nil {
			handleError(w, err)
			return
		}
		err = dataStorage.UpdateVariableBatch(storageName, lonVarName, lonData)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal output into JSON
		err = handleJSON(w, GeocodingResult{
			LatitudeField:  latVarName,
			LongitudeField: lonVarName,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal geocoded result into JSON"))
			return
		}
	}
}
