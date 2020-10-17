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
	"time"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil/api/model"
)

// SolutionResponse represents a pipeline solution.
type SolutionResponse struct {
	RequestID        string                 `json:"requestId"`
	Feature          string                 `json:"feature"`
	Dataset          string                 `json:"dataset"`
	Features         []*model.Feature       `json:"features"`
	Filters          *model.FilterParams    `json:"filters"`
	SolutionID       string                 `json:"solutionId"`
	FittedSolutionID string                 `json:"fittedSolutionId"`
	ResultID         string                 `json:"resultId"`
	Progress         string                 `json:"progress"`
	Scores           []*model.SolutionScore `json:"scores"`
	Timestamp        time.Time              `json:"timestamp"`
	PredictedKey     string                 `json:"predictedKey"`
	ErrorKey         string                 `json:"errorKey"`
	ConfidenceKey    string                 `json:"confidenceKey"`
}

// SolutionsHandler fetches solutions associated with a given dataset and target.
func SolutionsHandler(solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		dataset := handleNullParameter(pat.Param(r, "dataset"))
		target := handleNullParameter(pat.Param(r, "target"))

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		requests, err := solution.FetchRequestByDatasetTarget(dataset, target)
		if err != nil {
			handleError(w, err)
			return
		}

		solutions := make([]*SolutionResponse, 0)
		for _, req := range requests {
			// gather solutions
			reqSolutions, err := solution.FetchSolutionsByRequestID(req.RequestID)
			if err != nil {
				handleError(w, err)
				return
			}
			for _, sol := range reqSolutions {
				solution := &SolutionResponse{
					// request
					RequestID: req.RequestID,
					Dataset:   req.Dataset,
					Feature:   req.TargetFeature(),
					Features:  req.Features,
					Filters:   req.Filters,
					// solution
					SolutionID: sol.SolutionID,
					Scores:     sol.Scores,
					Timestamp:  sol.CreatedTime,
					Progress:   sol.State.Progress,
					// keys
					PredictedKey:  model.GetPredictedKey(sol.SolutionID),
					ErrorKey:      model.GetErrorKey(sol.SolutionID),
					ConfidenceKey: model.GetConfidenceKey(sol.SolutionID),
				}
				if len(sol.Results) > 0 {
					// result
					solution.ResultID = sol.Results[0].ResultUUID
					solution.FittedSolutionID = sol.Results[0].FittedSolutionID
				}
				solutions = append(solutions, solution)
			}
		}

		// marshal data and sent the response back
		err = handleJSON(w, solutions)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session solutions into JSON"))
			return
		}
	}
}

// SolutionHandler fetches a solution by its ID.
func SolutionHandler(solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		solutionID := pat.Param(r, "solution-id")

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		sol, err := solution.FetchSolution(solutionID)
		if err != nil {
			handleError(w, err)
			return
		}

		req, err := solution.FetchRequest(sol.RequestID)
		if err != nil {
			handleError(w, err)
			return
		}

		resultID := ""
		fittedSolutionID := ""
		if len(sol.Results) > 0 {
			resultID = sol.Results[0].ResultUUID
			fittedSolutionID = sol.Results[0].FittedSolutionID
		}

		solutionResponse := SolutionResponse{
			// request
			RequestID: req.RequestID,
			Dataset:   req.Dataset,
			Feature:   req.TargetFeature(),
			Features:  req.Features,
			Filters:   req.Filters,
			// solution
			SolutionID:       sol.SolutionID,
			Scores:           sol.Scores,
			Timestamp:        sol.CreatedTime,
			Progress:         sol.State.Progress,
			ResultID:         resultID,
			FittedSolutionID: fittedSolutionID,
			// keys
			PredictedKey:  model.GetPredictedKey(sol.SolutionID),
			ErrorKey:      model.GetErrorKey(sol.SolutionID),
			ConfidenceKey: model.GetConfidenceKey(sol.SolutionID),
		}

		// marshal data and sent the response back
		err = handleJSON(w, solutionResponse)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session solutions into JSON"))
			return
		}
	}
}
