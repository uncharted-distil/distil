package routes

import (
	"net/http"
	"strconv"

	"goji.io/pat"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

// PipelineCreateHandler creates a route to handle pipeline create requests.
func PipelineCreateHandler(pipelineService *pipeline.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		task := pat.Param(r, "task")
		metric := pat.Param(r, "metric")
		output := pat.Param(r, "output")
		feature := pat.Param(r, "feature")
		numPipelines := pat.Param(r, "num")

		request := pipeline.PipelineCreateRequest{}
		request.TrainDatasetUris[0] = dataset

		// task
		if val, ok := pipeline.Task_value[task]; ok {
			request.Task = pipeline.Task(val)
		} else {
			handleError(w, errors.Errorf("unhandled task type %s", task))
			return
		}

		// metric
		if val, ok := pipeline.Metric_value[metric]; ok {
			request.Metric[0] = pipeline.Metric(val)
		} else {
			handleError(w, errors.Errorf("unhandled metric type %s", metric))
		}

		// output
		if val, ok := pipeline.Output_value[output]; ok {
			request.Output = pipeline.Output(val)
		} else {
			handleError(w, errors.Errorf("unhandled output type %s", output))
		}

		// feature name
		request.TargetFeatures[0] = feature

		// num pipelines to request
		if val, err := strconv.Atoi(numPipelines); err != nil {
			request.MaxPipelines = int32(val)
		} else {
			handleError(w, errors.Wrapf(err, "num pipelines value %s is not an int", numPipelines))
		}
	}
}
