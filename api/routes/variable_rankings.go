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
	"goji.io/pat"

	//"github.com/uncharted-distil/distil-compute/model"

	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
)

// RankingResult represents a ranking response for a target variable.
type RankingResult struct {
	Rankings map[string]interface{} `json:"rankings"`
}

// VariableRankingHandler generates a route handler that allows to ranking
// variables of a dataset relative to the importance of a selected variable.
func VariableRankingHandler(metaCtor api.MetadataStorageCtor, solutionCtor api.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variable name
		target := pat.Param(r, "target")
		// get solution id (optional param)
		queryValues := r.URL.Query()
		solutionID := queryValues.Get("solution")

		var rankings map[string]interface{}
		var err error
		if solutionID == "" {
			rankings, err = targetRank(dataset, target, metaCtor)
		} else {
			rankings, err = solutionRank(solutionID, solutionCtor)
		}

		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}

		// marshal output into JSON
		err = handleJSON(w, RankingResult{
			Rankings: rankings,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
	}
}

func solutionRank(solutionID string, solutionCtor api.SolutionStorageCtor) (map[string]interface{}, error) {
	// get storage client
	storage, err := solutionCtor()
	if err != nil {
		return nil, err
	}

	weights, err := storage.FetchSolutionFeatureWeights(solutionID)
	if err != nil {
		return nil, err
	}

	ranks := make(map[string]interface{})
	for _, fw := range weights {
		ranks[fw.FeatureName] = fw.Weight
	}

	return ranks, nil
}

func targetRank(dataset string, target string, metaCtor api.MetadataStorageCtor) (map[string]interface{}, error) {
	// get storage client
	storage, err := metaCtor()
	if err != nil {
		return nil, err
	}

	d, err := storage.FetchDataset(dataset, false, false)
	if err != nil {
		return nil, err
	}

	// compute rankings
	rankings, err := task.TargetRank(d.Folder, target, d.Variables, d.Source)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for _, variable := range d.Variables {
		rank, ok := rankings[variable.Name]
		if ok {
			res[variable.Name] = rank
		} else {
			res[variable.Name] = 0
		}
	}

	return res, nil
}
