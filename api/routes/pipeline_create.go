package routes

import (
	"net/http"

	"goji.io/pat"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"strconv"

	"github.com/unchartedsoftware/distil/api/pipeline"
	log "github.com/unchartedsoftware/plog"
)

// PipelineCreateHandler is a thing that doesn't has the capacity to feel love
func PipelineCreateHandler(client *pipeline.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		sessionID := pat.Param(r, "session-id")
		task := pat.Param(r, "task")
		output := pat.Param(r, "output")
		metric := pat.Param(r, "metric")
		maxPipelines := pat.Param(r, "max-pipelines")

		pipelineCnt, err := strconv.Atoi(maxPipelines)
		if err != nil {
			pipelineCnt = 1
		}

		// generate the request
		createReq := pipeline.PipelineCreateRequest{
			Context:      &pipeline.SessionContext{SessionId: sessionID},
			Task:         pipeline.Task(pipeline.Task_value[task]),
			Output:       pipeline.Output(pipeline.Output_value[output]),
			Metric:       []pipeline.Metric{pipeline.Metric(pipeline.Metric_value[metric])},
			MaxPipelines: int32(pipelineCnt),
		}
		req := pipeline.GeneratePipelineCreateRequest(&createReq)

		session, err := client.GetSession(sessionID)
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to issue CreatePipelineRequest"))
			return
		}

		// gets an existing request or dispatchs a new one
		proxy, err := session.GetOrDispatch(context.Background(), req)
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to issue CreatePipelineRequest"))
			return
		}

		// process the result proxy, which is replicated for completed, pending requests
		for {
			select {
			case result := <-proxy.Results:
				res := (*result).(*pipeline.PipelineCreateResult)
				log.Infof("RESULT %v", res.String())
				w.Write([]byte(res.String()))
			case err := <-proxy.Errors:
				log.Info("ERROR")
				handleError(w, errors.Wrap(err, "failed to issue CreatePipelineRequest"))
				return
			case <-proxy.Done:
				log.Info("DONE")
				return
			}
		}
	}
}
