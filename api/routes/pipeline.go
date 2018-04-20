package routes

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// PipelineInfo represents the pipeline information relevant to the client.
type PipelineInfo struct {
	RequestID   string                 `json:"requestId"`
	Feature     string                 `json:"feature"`
	PipelineID  string                 `json:"pipelineId"`
	ResultUUID  string                 `json:"resultId"`
	Progress    string                 `json:"progress"`
	Scores      []*model.PipelineScore `json:"scores"`
	CreatedTime time.Time              `json:"timestamp"`
	Dataset     string                 `json:"dataset"`
	Features    []*model.Feature       `json:"features"`
	Filters     *model.FilterParams    `json:"filters"`
}

// PipelineResponse represents a request response
type PipelineResponse struct {
	Pipelines []*PipelineInfo `json:"pipelines"`
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

		requests, err := pipeline.FetchPipelineResultByDatasetTarget(dataset, target, pipelineID)
		if err != nil {
			handleError(w, err)
			return
		}

		// flatten the results
		pipelines := make([]*PipelineInfo, 0)
		for _, req := range requests {

			for _, pip := range req.Pipelines {
				pipeline := &PipelineInfo{
					// request
					RequestID: req.RequestID,
					Dataset:   req.Dataset,
					Feature:   req.TargetFeature(),
					Features:  req.Features,
					Filters:   req.Filters,
					// pipeline
					PipelineID:  pip.PipelineID,
					Scores:      pip.Scores,
					CreatedTime: pip.CreatedTime,
					Progress:    pip.Progress,
				}
				for _, res := range pip.Results {
					// result
					pipeline.CreatedTime = res.CreatedTime
					pipeline.ResultUUID = res.ResultUUID
					pipeline.Progress = res.Progress
				}

				pipelines = append(pipelines, pipeline)
			}
		}

		// marshall data and sent the response back
		err = handleJSON(w, &PipelineResponse{
			Pipelines: pipelines,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session pipelines into JSON"))
			return
		}

		return
	}
}
