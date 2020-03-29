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
	"net/url"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// PredictionResultsHandler fetches solution test result values, or model predictions and returns them to the client
// in a JSON structure
func PredictionResultsHandler(solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor, metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
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
		storageName := model.NormalizeDatasetID(dataset)

		produceRequestID, err := url.PathUnescape(pat.Param(r, "produce-request-id"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape produce request id"))
			return
		}

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
		filterParams.Merge(req.Filters)

		// Expand any grouped variables defined in filters into their subcomponents
		updatedFilterParams, err := api.ExpandFilterParams(dataset, filterParams, meta)
		if err != nil {
			handleError(w, err)
			return
		}

		// get the result URI
		res, err := solution.FetchSolutionResultByProduceRequestID(produceRequestID)
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

		results, err := data.FetchResults(dataset, storageName, res.ResultURI, solutionID, updatedFilterParams, true)
		if err != nil {
			handleError(w, err)
			return
		}

		// replace any NaN values with an empty string
		results = api.ReplaceNaNs(results, api.EmptyString)

		outputs := &api.PredictionResult{
			FilteredData:     results,
			FittedSolutionID: res.FittedSolutionID,
			ProduceRequestID: res.ProduceRequestID,
		}
		// marshal data and sent the response back
		err = handleJSON(w, outputs)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal solution result into JSON"))
			return
		}
	}
}
