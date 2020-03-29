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

// PredictionResponse represents a result from a produce call on a prediction dataset.
type PredictionResponse struct {
	RequestID        string    `json:"requestId"`
	FittedSolutionID string    `json:"fittedSolutionId"`
	Feature          string    `json:"feature"`
	Dataset          string    `json:"dataset"`
	Progress         string    `json:"progress"`
	Timestamp        time.Time `json:"timestamp"`
	ResultID         string    `json:"resultId"`
	PredictedKey     string    `json:"predictedKey"`
}

// PredictionsHandler fetches predictions associated with a given dataset and target.
func PredictionsHandler(solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		fittedSolutionID := pat.Param(r, "fitted-solution-id")

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		requests, err := solution.FetchPredictionsByFittedSolutionID(fittedSolutionID)
		if err != nil {
			handleError(w, err)
			return
		}

		predictions := make([]*PredictionResponse, 0)
		for _, req := range requests {
			// gather predictions
			solRes, err := solution.FetchSolutionResultsByFittedSolutionID(req.RequestID)
			if err != nil {
				handleError(w, err)
				return
			}
			for _, sol := range solRes {
				predictionResponse := &PredictionResponse{
					// request
					RequestID:        req.RequestID,
					Dataset:          req.Dataset,
					Feature:          req.Target,
					FittedSolutionID: req.FittedSolutionID,
					// solution
					Timestamp: sol.CreatedTime,
					Progress:  sol.Progress,
					// keys
					PredictedKey: model.GetPredictedKey(sol.SolutionID),
					ResultID:     sol.ResultUUID,
				}
				predictions = append(predictions, predictionResponse)
			}
		}

		// marshal data and sent the response back
		err = handleJSON(w, predictions)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal prediction results into JSON"))
			return
		}
	}
}

// PredictionHandler fetches a prediction by its ID.
func PredictionHandler(solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		requestID := pat.Param(r, "request-id")

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		solRes, err := solution.FetchSolutionResultByProduceRequestID(requestID)
		if err != nil {
			handleError(w, err)
			return
		}

		sol, err := solution.FetchSolution(solRes.SolutionID)
		if err != nil {
			handleError(w, err)
			return
		}

		req, err := solution.FetchPrediction(requestID)
		if err != nil {
			handleError(w, err)
			return
		}

		predictionResponse := &PredictionResponse{
			// request
			RequestID:        req.RequestID,
			Dataset:          req.Dataset,
			Feature:          req.Target,
			FittedSolutionID: req.FittedSolutionID,
			// solution
			Timestamp: sol.CreatedTime,
			Progress:  solRes.Progress,
			// keys
			PredictedKey: model.GetPredictedKey(sol.SolutionID),
			ResultID:     solRes.ResultUUID,
		}

		// marshal data and sent the response back
		err = handleJSON(w, predictionResponse)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal predictions into JSON"))
			return
		}
	}
}
