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
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
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
		field := params["field"].(string)
		typ := params["type"].(string)
		dataset := pat.Param(r, "dataset")
		storageName := model.NormalizeDatasetID(dataset)

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

		// update the variable type in the storage
		err = storage.SetDataType(dataset, storageName, field, typ)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to update the data type in storage"))
			return
		}

		// update the variable type in the metadata
		err = meta.SetDataType(dataset, field, typ)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to update the data type in metadata"))
			return
		}

		// update the extremas stored in ES
		err = api.UpdateExtremas(dataset, field, meta, storage)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to update the extremas in metadata"))
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

func getPostParameters(r *http.Request) (map[string]interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse POST request")
	}

	params := make(map[string]interface{})
	err = json.Unmarshal(body, &params)

	return params, err
}
