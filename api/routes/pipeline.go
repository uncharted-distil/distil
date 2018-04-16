package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// Pipeline represents a session response
type Pipeline struct {
	Pipelines []*model.PipelineResult `json:"pipelines"`
	Features  []*model.Feature        `json:"features"`
	Filters   *model.FilterParams     `json:"filters"`
}

// PipelineHandler fetches existing pipelines.
func PipelineHandler(pipelineCtor model.PipelineStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		dataset := pat.Param(r, "dataset")
		target := pat.Param(r, "target")
		pipelineID := pat.Param(r, "pipeline-id")

		if pipelineID == "null" {
			pipelineID = ""
		}
		if dataset == "null" {
			dataset = ""
		}
		if target == "null" {
			target = ""
		}

		pipeline, err := pipelineCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		requests, err := pipeline.FetchRequestPipelineResultByDatasetTarget(dataset, target, pipelineID)
		if err != nil {
			handleError(w, err)
			return
		}

		// Blank the result URI & flatten the results.
		pipelines := make(map[string]*Pipeline)
		for _, req := range requests {
			for _, pip := range req.Pipelines {
				for _, res := range pip.Results {
					res.ResultURI = ""
					if pipelines[res.PipelineID] == nil {
						pipelines[res.PipelineID] = &Pipeline{
							Pipelines: make([]*model.PipelineResult, 0),
							Features:  req.Features,
							Filters:   req.Filters,
						}
					}
					pipelines[res.PipelineID].Pipelines = append(pipelines[res.PipelineID].Pipelines, res)
				}
			}
		}

		// marshall data and sent the response back
		err = handleJSON(w, pipelines)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session pipelines into JSON"))
			return
		}

		return
	}
}
