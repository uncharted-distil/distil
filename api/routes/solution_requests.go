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

	"github.com/uncharted-distil/distil/api/model"
)

// SolutionRequest represents a information used to make a pipeline generation request.
type SolutionRequest struct {
	RequestID string              `json:"requestId"`
	Feature   string              `json:"feature"`
	Dataset   string              `json:"dataset"`
	Features  []*model.Feature    `json:"features"`
	Filters   *model.FilterParams `json:"filters"`
	Progress  string              `json:"progress"`
	Timestamp time.Time           `json:"timestamp"`
}

// SolutionRequestsHandler fetches search request for a given dataset and target.
func SolutionRequestsHandler(solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
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
		resp := make([]*SolutionRequest, len(requests))
		for i, req := range requests {
			resp[i] = &SolutionRequest{
				// request
				RequestID: req.RequestID,
				Dataset:   req.Dataset,
				Feature:   req.TargetFeature(),
				Features:  req.Features,
				Filters:   req.Filters,
				Timestamp: req.CreatedTime,
				Progress:  req.Progress,
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

// SolutionRequestHandler fetches a solution request by its ID.
func SolutionRequestHandler(solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		requestID := pat.Param(r, "request-id")

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		req, err := solution.FetchRequest(requestID)
		if err != nil {
			handleError(w, err)
			return
		}
		resp := SolutionRequest{
			// request
			RequestID: req.RequestID,
			Dataset:   req.Dataset,
			Feature:   req.TargetFeature(),
			Features:  req.Features,
			Filters:   req.Filters,
			Timestamp: req.CreatedTime,
			Progress:  req.Progress,
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
