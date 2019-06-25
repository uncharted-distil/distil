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

	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
)

// ComposeHandler inserts a new field based on existing fields.
func ComposeHandler(dataCtor api.DataStorageCtor, esMetaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		storageName := model.NormalizeDatasetID(dataset)

		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}
		varName, _ := params["varName"].(string)
		variables, _ := json.StringArray(params, "variables")

		// initialize the storage
		metaStorage, err := esMetaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		dataStorage, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// create the new field
		err = metaStorage.AddVariable(dataset, varName, model.TextType, "grouping")
		if err != nil {
			handleError(w, err)
			return
		}
		err = dataStorage.AddVariable(dataset, storageName, varName, model.TextType)
		if err != nil {
			handleError(w, err)
			return
		}

		// map fields for the hash building
		fields := make([]*model.Variable, 0)
		for i := 0; i < len(variables); i++ {
			v, err := metaStorage.FetchVariable(dataset, variables[i])
			if err != nil {
				handleError(w, err)
				return
			}
			fields = append(fields, v)
		}
		d3mIndexField, err := metaStorage.FetchVariable(dataset, model.D3MIndexName)
		if err != nil {
			handleError(w, err)
			return
		}

		// read the data from storage
		rawData, err := dataStorage.FetchData(dataset, storageName, &api.FilterParams{}, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// cycle through all the data
		hashData := make(map[string]string)
		for _, r := range rawData.Values {
			// create the hash from the specified columns
			hash, err := createFieldHash(r, fields)
			if err != nil {
				handleError(w, err)
				return
			}
			hashData[r[d3mIndexField.Index].(string)] = hash
		}

		// save the new column
		err = dataStorage.UpdateVariableBatch(storageName, varName, hashData)
		if err != nil {
			handleError(w, err)
			return
		}
	}
}

func createFieldHash(data []interface{}, fields []*model.Variable) (string, error) {
	// pull the fields to hash
	dataToHash := make([]interface{}, 0)
	for i := 0; i < len(fields); i++ {
		dataToHash = append(dataToHash, data[i])
	}

	// hash the desired fields
	hashInt, err := hashstructure.Hash(dataToHash, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", hashInt), nil
}
