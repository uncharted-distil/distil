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

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// ConfidenceSummaryHandler bins predicted result confidence data for consumption in a downstream summary view.
func ConfidenceSummaryHandler(metaCtor api.MetadataStorageCtor, solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		dataset := pat.Param(r, "dataset")

		// get variable summary mode
		mode, err := api.SummaryModeFromString(pat.Param(r, "mode"))
		if err != nil {
			handleError(w, err)
			return
		}

		resultUUID, err := url.PathUnescape(pat.Param(r, "results-uuid"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape results uuid"))
			return
		}

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

		// get the result URI. Error ignored to make it ES compatible.
		res, err := solution.FetchSolutionResultByUUID(resultUUID)
		if err != nil {
			handleError(w, err)
			return
		}
		if res == nil {
			err = handleJSON(w, SummaryResult{})
			if err != nil {
				handleError(w, errors.Wrap(err, "unable marshal nil histogram into JSON"))
			}
			return
		}
		// if the variable is a geobounds and there is a band column, add a filter
		// to only consider the first band.
		hasBand := false
		isGeobounds := false
		for _, v := range ds.Variables {
			if v.DisplayName == "band" {
				hasBand = true
			} else if model.IsGeoBounds(v.Type) {
				isGeobounds = true
			}
		}
		if hasBand && isGeobounds {
			boundsFilter := model.NewCategoricalFilter("band", model.IncludeFilter, []string{"01"})
			boundsFilter.IsBaselineFilter = true
			filterParams.AddFilter(boundsFilter)
		}
		// fetch summary histogram
		summary, err := data.FetchConfidenceSummary(dataset, storageName, res.ResultURI, filterParams, api.SummaryMode(mode))
		if err != nil {
			handleError(w, err)
			return
		}

		if summary["confidence"] != nil {
			confSummary := summary["confidence"]
			if confSummary.Baseline.IsEmpty() {
				summary["confidence"] = nil
			} else {
				confSummary.Key = api.GetConfidenceKey(res.ResultUUID)
				confSummary.Label = "Confidence"
			}
		}
		if summary["rank"] != nil {
			rankSummary := summary["rank"]
			if rankSummary.Baseline.IsEmpty() {
				summary["rank"] = nil
			} else {
				rankSummary.Key = api.GetRankKey(res.ResultUUID)
				rankSummary.Label = "Rank"
			}
		}

		// marshal data and sent the response back
		err = handleJSON(w, summary)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
