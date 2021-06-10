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
	"goji.io/v3/pat"

	api "github.com/uncharted-distil/distil/api/model"
)

// ResultsHandler fetches predicted solution values and returns them to the client
// in a JSON structure
func ResultsHandler(solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor, metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
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

		dataset := pat.Param(r, "dataset")

		solutionID, err := url.PathUnescape(pat.Param(r, "solution-id"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape solution id"))
			return
		}

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		data, err := dataCtor()
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

		// get the filters
		req, err := solution.FetchRequestBySolutionID(solutionID)
		if err != nil {
			handleError(w, err)
			return
		}
		if req == nil {
			handleError(w, errors.Errorf("solution id `%s` cannot be mapped to result URI", solutionID))
			return
		}

		// merge provided filterParams with those of the request
		filterParams.Variables = req.Filters.Variables
		// Expand any grouped variables defined in filters into their subcomponents
		updatedFilterParams, err := api.ExpandFilterParams(dataset, filterParams, false, meta)
		if err != nil {
			handleError(w, err)
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

		results, err := data.FetchResults(dataset, storageName, res[0].ResultURI, res[0].ResultUUID, updatedFilterParams, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// replace any NaN values with an empty string
		resultsTransformed := transformDataForClient(results, api.EmptyString)

		outputs := &PredictionResult{
			FilteredDataClient: resultsTransformed,
			FittedSolutionID:   res[0].FittedSolutionID,
			ProduceRequestID:   res[0].ProduceRequestID,
		}
		// marshal data and sent the response back
		err = handleJSON(w, outputs)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result rows into JSON"))
			return
		}
	}
}
