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
	"net/url"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
	"goji.io/v3/pat"
)

const (
	orderBy = "orderBy"
)

// FilteredDataClient is the structure the client requires when fetching data.
type FilteredDataClient struct {
	NumRows          int                        `json:"numRows"`
	NumRowsFiltered  int                        `json:"numRowsFiltered"`
	Columns          []*api.Column              `json:"columns"`
	Values           [][]*api.FilteredDataValue `json:"values"`
	FittedSolutionID string                     `json:"fittedSolutionId"`
	ProduceRequestID string                     `json:"produceRequestId"`
}

// DataHandler creates a route that fetches filtered data from backing storage instance.
func DataHandler(storageCtor api.DataStorageCtor, metaCtor api.MetadataStorageCtor, solutionCtor api.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
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
		var data *api.FilteredData

		if len(filterParams.Filters) < 1 && filterParams.Invert {
			// inverted empty filter means return no data
			err = handleJSON(w, api.EmptyFilterData())
			if err != nil {
				handleError(w, errors.Wrap(err, "unable marshal result rows into JSON"))
			}
			return
		}

		dataset := pat.Param(r, "dataset")
		includeGroupingCol := params["include-grouping-col"]
		includeGroupingColBool := parseBoolParam(includeGroupingCol.(string))

		solutionID, err := url.PathUnescape(params["solution-id"].(string))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape solution id"))
			return
		}

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
		vars, err := metaStore.FetchVariables(dataset, true, true, true)
		if err != nil {
			handleError(w, err)
			return
		}
		ds, err := metaStore.FetchDataset(dataset, false, false, false)
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

		fittedSolutionID := ""
		produceRequestID := ""
		if solutionID != "" {
			// get results using the solution id
			solution, err := solutionCtor()
			if err != nil {
				handleError(w, err)
				return
			}

			req, err := solution.FetchRequestBySolutionID(solutionID)
			if err != nil {
				handleError(w, err)
				return
			}
			if req == nil {
				handleError(w, errors.Errorf("solution id `%s` cannot be mapped to result URI", solutionID))
				return
			}

			// get the result URI
			res, err := solution.FetchSolutionResults(solutionID)
			if err != nil {
				handleError(w, err)
				return
			}

			// if no result, return an empty map
			if res == nil {
				err = handleJSON(w, make(map[string]interface{}))
				if err != nil {
					handleError(w, errors.Wrap(err, "unable marshal version into JSON and write response"))
				}
				return
			}

			data, err = storage.FetchResults(dataset, storageName, res[0].ResultURI, res[0].ResultUUID, expandedFilterParams, false)
			if err != nil {
				handleError(w, err)
				return
			}
			fittedSolutionID = res[0].FittedSolutionID
			produceRequestID = res[0].ProduceRequestID
		} else {
			var orderByVar *model.Variable
			if params[orderBy] != nil {
				for _, v := range vars {
					if v.HeaderName == params[orderBy] {
						orderByVar = v
						break
					}
				}
			}

			// fetch filtered data based on the supplied search parameters
			data, err = storage.FetchData(dataset, storageName, expandedFilterParams, includeGroupingColBool, orderByVar)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable fetch filtered data"))
				return
			}
		}

		// replace NaNs with an empty string to make them JSON encodable
		dataTransformed := transformDataForClient(data, api.EmptyString)
		dataTransformed.FittedSolutionID = fittedSolutionID
		dataTransformed.ProduceRequestID = produceRequestID

		// marshal output into JSON
		bytes, err := json.Marshal(dataTransformed)
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

func transformDataForClient(data *api.FilteredData, replacementType api.NaNReplacement) *FilteredDataClient {
	data = api.ReplaceNaNs(data, replacementType)
	dataTransformed := &FilteredDataClient{
		NumRows:         data.NumRows,
		NumRowsFiltered: data.NumRowsFiltered,
		Values:          data.Values,
		Columns:         make([]*api.Column, len(data.Columns)),
	}

	for _, c := range data.Columns {
		dataTransformed.Columns[c.Index] = c
	}

	return dataTransformed
}
