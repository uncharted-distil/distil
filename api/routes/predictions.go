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
	"time"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// PredictionResponse represents a result from a produce call on a prediction dataset.
type PredictionResponse struct {
	RequestID        string         `json:"requestId"`
	FittedSolutionID string         `json:"fittedSolutionId"`
	Feature          string         `json:"feature"`
	Features         []*api.Feature `json:"features"`
	Dataset          string         `json:"dataset"`
	Progress         string         `json:"progress"`
	Timestamp        time.Time      `json:"timestamp"`
	ResultID         string         `json:"resultId"`
	PredictedKey     string         `json:"predictedKey"`
}

// PredictionsHandler fetches predictions associated with a given dataset and target.
func PredictionsHandler(solutionCtor api.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
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

		// recover the original feature set
		solutionReq, err := solution.FetchRequestByFittedSolutionID(fittedSolutionID)
		if err != nil {
			handleError(w, err)
			return
		}
		features := make([]*api.Feature, len(solutionReq.Features)-1)
		for i, v := range solutionReq.Features {
			if v.FeatureType != model.FeatureTypeTarget {
				features[i] = v
			}
		}

		predictions := []*PredictionResponse{}
		for _, predictionReq := range requests {
			// check to see if we have results ready and the ID if so
			predictionResults, err := solution.FetchPredictionResultByProduceRequestID(predictionReq.RequestID)
			if err != nil {
				handleError(w, err)
				return
			}
			var resultID string
			if predictionResults != nil {
				resultID = predictionResults.ResultUUID
			}

			predictionResponse := &PredictionResponse{
				// request
				RequestID:        predictionReq.RequestID,
				Dataset:          predictionReq.Dataset,
				Feature:          predictionReq.Target,
				Features:         features,
				FittedSolutionID: predictionReq.FittedSolutionID,
				// solution
				Timestamp: predictionReq.CreatedTime,
				Progress:  predictionReq.Progress,
				// keys
				PredictedKey: api.GetPredictedKey(predictionReq.RequestID),
				ResultID:     resultID,
			}
			predictions = append(predictions, predictionResponse)
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
func PredictionHandler(solutionCtor api.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		requestID := pat.Param(r, "request-id")

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		predictionReq, err := solution.FetchPrediction(requestID)
		if err != nil {
			handleError(w, err)
			return
		}

		// recover the original feature set
		solutionReq, err := solution.FetchRequestByFittedSolutionID(predictionReq.FittedSolutionID)
		if err != nil {
			handleError(w, err)
			return
		}
		features := make([]*api.Feature, len(solutionReq.Features)-1)
		for i, v := range solutionReq.Features {
			if v.FeatureType != model.FeatureTypeTarget {
				features[i] = v
			}
		}

		// check to see if we have results ready and the ID if so
		predictionResults, err := solution.FetchPredictionResultByProduceRequestID(predictionReq.RequestID)
		if err != nil {
			handleError(w, err)
			return
		}
		var resultID string
		if predictionResults != nil {
			resultID = predictionResults.ResultUUID
		}

		predictionResponse := &PredictionResponse{
			// request
			RequestID:        predictionReq.RequestID,
			Dataset:          predictionReq.Dataset,
			Feature:          predictionReq.Target, // target name
			Features:         features,             // features used in making predictions
			FittedSolutionID: predictionReq.FittedSolutionID,
			// solution
			Timestamp: predictionReq.CreatedTime,
			Progress:  predictionReq.Progress,
			//keys
			PredictedKey: api.GetPredictedKey(predictionReq.RequestID),
			ResultID:     resultID,
		}

		// marshal data and sent the response back
		err = handleJSON(w, predictionResponse)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal predictions into JSON"))
			return
		}
	}
}
