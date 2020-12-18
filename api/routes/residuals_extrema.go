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
	"math"
	"net/http"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// ResidualsExtrema contains a residual extrema response.
type ResidualsExtrema struct {
	Extrema *api.Extrema `json:"extrema"`
}

func fetchSolutionResidualExtrema(meta api.MetadataStorage, data api.DataStorage, solution api.SolutionStorage, dataset string, storageName string, target string, solutionID string) (*api.Extrema, error) {
	// check target var type
	variable, err := meta.FetchVariable(dataset, target)
	if err != nil {
		return nil, err
	}

	if !model.IsNumerical(variable.Type) && variable.Type != model.TimeSeriesType {
		return nil, nil
	}

	// we need to get extrema min and max for all solutions sharing dataset and target
	solutions, err := solution.FetchSolutionsByDatasetTarget(dataset, target)
	if err != nil {
		return nil, err
	}

	// get extrema
	min := math.MaxFloat64
	max := -math.MaxFloat64
	for _, sol := range solutions {
		if len(sol.Results) > 0 && !sol.IsBad {
			// result uri
			resultURI := sol.Results[0].ResultURI
			// predicted extrema
			residualExtrema, err := data.FetchResidualsExtremaByURI(dataset, storageName, resultURI)
			if err != nil {
				return nil, err
			}
			max = math.Max(max, residualExtrema.Max)
			min = math.Min(min, residualExtrema.Min)
		}
	}

	// make symmetrical
	extremum := math.Max(math.Abs(min), math.Abs(max))

	return api.NewExtrema(-extremum, extremum)
}

// ResidualsExtremaHandler returns the extremas for a residual summary.
func ResidualsExtremaHandler(metaCtor api.MetadataStorageCtor, solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		target := pat.Param(r, "target")

		meta, err := metaCtor()
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

		ds, err := meta.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName

		// extract extrema for solution
		extrema, err := fetchSolutionResidualExtrema(meta, data, solution, dataset, storageName, target, "")
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data and sent the response back
		err = handleJSON(w, ResidualsExtrema{
			Extrema: extrema,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
