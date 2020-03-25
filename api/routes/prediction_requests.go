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

// PredictionRequest represents a information used to make a prediction request.
type PredictionRequest struct {
	RequestID        string    `json:"requestId"`
	FittedSolutionID string    `json:"fittedSolutionId"`
	Feature          string    `json:"feature"`
	Dataset          string    `json:"dataset"`
	Progress         string    `json:"progress"`
	Timestamp        time.Time `json:"timestamp"`
}

// PredictionRequestsHandler fetches prediction request for a given dataset and target.
func PredictionRequestsHandler(solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		fittedSolutionID := pat.Param(r, "fitted-solution-id")

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		preds, err := solution.FetchPredictionsByFittedSolutionID(fittedSolutionID)
		if err != nil {
			handleError(w, err)
			return
		}
		resp := make([]*PredictionRequest, len(preds))
		for i, p := range preds {
			resp[i] = &PredictionRequest{
				RequestID:        p.RequestID,
				Dataset:          p.Dataset,
				FittedSolutionID: p.FittedSolutionID,
				Feature:          p.Target,
				Timestamp:        p.CreatedTime,
				Progress:         p.Progress,
			}
		}

		// marshal data and sent the response back
		err = handleJSON(w, resp)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session solutions into JSON"))
			return
		}
	}
}

// PredictionRequestHandler fetches a prediction request by its ID.
func PredictionRequestHandler(solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		requestID := pat.Param(r, "request-id")

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		pred, err := solution.FetchPrediction(requestID)
		if err != nil {
			handleError(w, err)
			return
		}

		resp := &PredictionRequest{
			RequestID:        pred.RequestID,
			Dataset:          pred.Dataset,
			FittedSolutionID: pred.FittedSolutionID,
			Feature:          pred.Target,
			Timestamp:        pred.CreatedTime,
			Progress:         pred.Progress,
		}

		// marshal data and sent the response back
		err = handleJSON(w, resp)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session solutions into JSON"))
			return
		}
	}
}

func handleNullParameter(value string) string {
	if value == "null" {
		return ""
	}

	return value
}
