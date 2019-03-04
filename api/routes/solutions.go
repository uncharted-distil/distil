package routes

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/uncharted-distil/distil/api/model"
)

// Solution represents a pipeline solution.
type Solution struct {
	RequestID    string                 `json:"requestId"`
	Feature      string                 `json:"feature"`
	SolutionID   string                 `json:"solutionId"`
	ResultUUID   string                 `json:"resultId"`
	Progress     string                 `json:"progress"`
	Scores       []*model.SolutionScore `json:"scores"`
	Timestamp    time.Time              `json:"timestamp"`
	Dataset      string                 `json:"dataset"`
	Features     []*model.Feature       `json:"features"`
	Filters      *model.FilterParams    `json:"filters"`
	PredictedKey string                 `json:"predictedKey"`
	ErrorKey     string                 `json:"errorKey"`
}

// RequestResponse represents a request response.
type RequestResponse struct {
	RequestID string      `json:"requestId"`
	Dataset   string      `json:"dataset"`
	Feature   string      `json:"feature"`
	Progress  string      `json:"progress"`
	Timestamp time.Time   `json:"timestamp"`
	Solutions []*Solution `json:"solutions"`
}

// SolutionHandler fetches existing solutions.
func SolutionHandler(solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		dataset := pat.Param(r, "dataset")
		target := pat.Param(r, "target")
		solutionID := pat.Param(r, "solution-id")

		if solutionID == "null" {
			solutionID = ""
		}
		if dataset == "null" {
			dataset = ""
		}
		if target == "null" {
			target = ""
		}

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		requests, err := solution.FetchRequestByDatasetTarget(dataset, target, solutionID)
		if err != nil {
			handleError(w, err)
			return
		}

		response := make([]*RequestResponse, 0)

		for _, req := range requests {

			// gather solutions
			solutions := make([]*Solution, 0)
			for _, sol := range req.Solutions {

				solution := &Solution{
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
					Progress:   sol.Progress,
					// keys
					PredictedKey: model.GetPredictedKey(req.TargetFeature(), sol.SolutionID),
					ErrorKey:     model.GetErrorKey(req.TargetFeature(), sol.SolutionID),
				}
				if sol.Result != nil {
					// result
					solution.ResultUUID = sol.Result.ResultUUID
				}
				solutions = append(solutions, solution)
			}

			response = append(response, &RequestResponse{
				RequestID: req.RequestID,
				Dataset:   req.Dataset,
				Feature:   req.TargetFeature(),
				Progress:  req.Progress,
				Timestamp: req.CreatedTime,
				Solutions: solutions,
			})
		}

		// marshal data and sent the response back
		err = handleJSON(w, response)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session solutions into JSON"))
			return
		}
	}
}
