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

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	api "github.com/uncharted-distil/distil/api/model"
)

// SolutionVariableRankingHandler generates a route handler that returns the importances associated with
// with a solution's features.  If a given pipeline produced feature level importances, we'll use those.
// As a fallback we will rank by the feature/target combination.
func SolutionVariableRankingHandler(metaCtor api.MetadataStorageCtor, solutionCtor api.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		solutionID := pat.Param(r, "solution-id")

		// get storage client
		storage, err := metaCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to connect to ES"))
			return
		}

		solutionStorage, err := solutionCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to connect to pg"))
		}
		request, err := solutionStorage.FetchRequestBySolutionID(solutionID)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to fetch request"))
		}

		dataset, err := storage.FetchDataset(request.Dataset, false, false, false)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to fetch dataset"))
			return
		}

		// first attempt to generate the rankings using the proper solution rankings that
		// are derived from the model
		var rankings map[string]float64
		if solutionID != "" {
			rankings, err = solutionRank(solutionID, solutionCtor)
		}
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to fetch solution ranks"))
			return
		}

		// if no ranking were generated, fall back on the dataset/target rank estimates
		if len(rankings) == 0 {
			summaryVariables, err := api.FetchSummaryVariables(dataset.ID, storage)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to expand grouped variables"))
				return
			}
			rankings, err = targetRank(dataset.ID, request.TargetFeature(), dataset.Folder, summaryVariables, dataset.Source)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable get variable ranking"))
				return
			}
		}

		res := make(map[string]interface{})
		for _, variable := range dataset.Variables {
			rank, ok := rankings[variable.Key]
			if ok {
				res[variable.Key] = rank
			} else {
				res[variable.Key] = nil
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

func solutionRank(solutionID string, solutionCtor api.SolutionStorageCtor) (map[string]float64, error) {
	// get storage client
	storage, err := solutionCtor()
	if err != nil {
		return nil, err
	}

	// Fetch the per-feature weights that were returned by the model.  Not all models support
	// this functionality, so this can potentially return an empty map.
	weights, err := storage.FetchSolutionWeights(solutionID)
	if err != nil {
		return nil, err
	}

	ranks := make(map[string]float64)
	for _, fw := range weights {
		ranks[fw.FeatureName] = fw.Weight
	}

	return ranks, nil
}
