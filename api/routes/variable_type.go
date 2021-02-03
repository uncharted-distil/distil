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
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
)

// VariableTypeHandler generates a route handler that facilitates the update
// of a variable type.
func VariableTypeHandler(storageCtor api.DataStorageCtor, metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}
		field, ok := json.String(params, "field")
		if !ok {
			handleError(w, errors.Wrap(err, "Unable to parse `field` parameter"))
			return
		}
		typ, ok := json.String(params, "type")
		if !ok {
			handleError(w, errors.Wrap(err, "Unable to parse `type` parameter"))
			return
		}
		dataset := pat.Param(r, "dataset")

		// get clients
		storage, err := storageCtor()
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

		// check the variable type to make sure it is valid
		isValid, err := storage.IsValidDataType(dataset, storageName, field, typ)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to verify the data type in storage"))
			return
		}
		if !isValid {
			handleErrorType(
				w,
				fmt.Errorf("unable to verify the data type in storage"),
				http.StatusBadRequest)
			return
		}

		// update the variable type in the storage
		err = setDataType(meta, storage, dataset, storageName, field, typ)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to update the data type in storage"))
			return
		}

		// update the extremas stored in ES
		err = api.UpdateExtremas(dataset, field, meta, storage)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to update the extremas in metadata"))
			return
		}

		variables, err := api.FetchSummaryVariables(dataset, meta)
		if err != nil {
			handleError(w, err)
			return
		}

		for _, v := range variables {
			if model.IsNumerical(v.Type) || model.IsDateTime(v.Type) {
				extrema, err := storage.FetchExtrema(ds.ID, ds.StorageName, v)
				if err != nil {
					log.Warnf("defaulting extrema values due to error fetching extrema for '%s': %+v", v.Key, err)
					extrema = getDefaultExtrema(v)
				}
				v.Min = extrema.Min
				v.Max = extrema.Max
			}
		}
		// marshal data
		err = handleJSON(w, VariablesResult{
			Variables: variables,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}

func getPostParameters(r *http.Request) (map[string]interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse POST request")
	}

	return json.Unmarshal(body)
}
