package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// Session represents a session response
type Session struct {
	Pipelines []*model.Result `json:"pipelines"`
}

// SessionHandler fetches existing pipelines for a session.
func SessionHandler(storageCtor model.PipelineStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		sessionID := pat.Param(r, "session")
		dataset := pat.Param(r, "dataset")
		target := pat.Param(r, "target")
		pipelineID := pat.Param(r, "pipeline-id")

		client, err := storageCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		requests, err := client.FetchRequests(sessionID)
		if err != nil {
			handleError(w, err)
			return
		}

		// Blank the result URI.
		for _, req := range requests {
			for _, res := range req.Results {
				res.ResultURI = ""
			}
		}

		// TODO: FILTER BY DATASET / TARGET FEATURE IN SQL!
		var filtered []*model.Request
		for _, req := range requests {
			if dataset == "null" || req.Dataset == dataset {
				if target == "null" {
					filtered = append(filtered, req)
				} else {
					for _, feature := range req.Features {
						if feature.FeatureType == model.FeatureTypeTarget &&
							feature.FeatureName == target {
							filtered = append(filtered, req)
						}
					}
				}
			}
		}

		// TODO: FILTER BY LATEST RESULT IN SQL!
		latest := make(map[string]*model.Result)
		for _, req := range filtered {
			for _, res := range req.Results {
				// TODO: ADD DATASET TO TABLE
				res.Dataset = req.Dataset
				current, ok := latest[res.PipelineID]
				if !ok {
					latest[res.PipelineID] = res
				} else {
					if current.CreatedTime.Before(res.CreatedTime) {
						latest[res.PipelineID] = res
					}
				}
			}
		}

		// TODO: FILTER BY PIPELINE-ID IN SQL!
		var final []*model.Result
		for _, res := range latest {
			if pipelineID == "null" || res.PipelineID == pipelineID {
				final = append(final, res)
			}
		}

		// marshall data and sent the response back
		err = handleJSON(w, Session{
			Pipelines: final,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session pipelines into JSON"))
			return
		}

		return
	}
}
