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

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"

	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
)

// VariableRankingHandler generates a route handler that allows ranking variables of a dataset relative to the importance
// of a selected variable.  This ranking is potentially long running.
func VariableRankingHandler(metaCtor api.MetadataStorageCtor, solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variable name
		target := pat.Param(r, "target")

		// get storage client
		storage, err := metaCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to connect to ES"))
			return
		}

		d, err := storage.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to fetch dataset"))
			return
		}

		var rankings map[string]float64
		summaryVariables, err := api.FetchSummaryVariables(dataset, storage)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to expand grouped variables"))
			return
		}
		rankings, err = targetRank(dataset, target, d.Folder, summaryVariables, d.Source)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable get variable ranking"))
			return
		}

		res := make(map[string]interface{})
		for _, variable := range d.Variables {
			rank, ok := rankings[variable.StorageName]
			if ok {
				res[variable.StorageName] = rank
			} else {
				res[variable.StorageName] = nil
			}
		}

		// marshal output into JSON
		err = handleJSON(w, res)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
	}
}

func targetRank(dataset string, target string, folder string, variables []*model.Variable, source metadata.DatasetSource) (map[string]float64, error) {
	// compute rankings
	rankings, err := task.TargetRank(folder, target, variables, source)
	if err != nil {
		return nil, err
	}
	return rankings, nil
}
