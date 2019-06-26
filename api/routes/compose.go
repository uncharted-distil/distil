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

		// check if the compose var exists already
		composeExists, err := metaStorage.DoesVariableExist(dataset, varName)
		if err != nil {
			handleError(w, err)
			return
		}

		if !composeExists {
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
		}

		// read the data from storage
		filter := &api.FilterParams{
			Variables: variables,
		}
		rawData, err := dataStorage.FetchData(dataset, storageName, filter, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// map fields
		d3mIndexFieldindex := -1
		mappedFields := make(map[string]int)
		for i, c := range rawData.Columns {
			if c.Label == model.D3MIndexName {
				d3mIndexFieldindex = i
			} else {
				mappedFields[c.Label] = i
			}
		}

		// cycle through all the data
		hashData := make(map[string]string)
		for _, r := range rawData.Values {
			// create the hash from the specified columns
			hash, err := createFieldHash(r, variables, mappedFields)
			if err != nil {
				handleError(w, err)
				return
			}
			hashData[fmt.Sprintf("%v", r[d3mIndexFieldindex])] = hash
		}

		// save the new column
		err = dataStorage.UpdateVariableBatch(storageName, varName, hashData)
		if err != nil {
			handleError(w, err)
			return
		}
	}
}

func createFieldHash(data []interface{}, fields []string, mappedFields map[string]int) (string, error) {
	// pull the fields to hash
	dataToHash := make([]interface{}, 0)
	for i := 0; i < len(fields); i++ {
		fieldIndex := mappedFields[fields[i]]
		dataToHash = append(dataToHash, data[fieldIndex])
	}

	// hash the desired fields
	hashInt, err := hashstructure.Hash(dataToHash, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", hashInt), nil
}
