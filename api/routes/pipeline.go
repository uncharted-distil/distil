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

		results, err := pipeline.FetchPipelineResultByDatasetTarget(dataset, target, pipelineID)
		if err != nil {
			handleError(w, err)
			return
		}

		// Blank the result URI.
		for _, res := range results {
			res.ResultURI = ""
		}

		// marshall data and sent the response back
		err = handleJSON(w, Pipeline{
			Pipelines: results,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session pipelines into JSON"))
			return
		}

		return
	}
}
