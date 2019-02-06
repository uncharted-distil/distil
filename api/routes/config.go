package routes

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/unchartedsoftware/distil/api/compute"
	"github.com/unchartedsoftware/distil/api/env"
)

// ConfigHandler returns the compiled version number, timestamp and initial config.
func ConfigHandler(config env.Config, version string, timestamp string, problemPath string, datasetDocPath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		target := "unknown"
		dataset := "unknown"
		taskType := "unknown"
		taskSubType := "unknown"
		var metrics []string

		if config.IsTask1 {
			// load dataset file
			dataDoc, err := compute.LoadDatasetSchemaFromFile(datasetDocPath)
			if err == nil {
				dataset = "d_" + dataDoc.About.DatasetID
			}
		}

		if config.IsTask2 {
			// load problem file
			problem, err := compute.LoadProblemSchemaFromFile(problemPath)
			if err == nil {
				// get inputs
				if problem.Inputs != nil {
					if len(problem.Inputs.Data) > 0 {
						// get dataset
						dataset = problem.Inputs.Data[0].DatasetID
						// get targets
						if len(problem.Inputs.Data[0].Targets) > 0 {
							target = problem.Inputs.Data[0].Targets[0].ColName
						}
					}
					// get metrics
					if problem.Inputs.PerformanceMetrics != nil {
						for _, metric := range problem.Inputs.PerformanceMetrics {
							metrics = append(metrics, metric.Metric)
						}
					}
				}
				// get task types
				if problem.About != nil {
					taskType = problem.About.TaskType
					taskSubType = problem.About.TaskSubType
				}
			}
		}

		// marshal version
		err := handleJSON(w, map[string]interface{}{
			"version":     version,
			"timestamp":   timestamp,
			"isTask1":     config.IsTask1,
			"isTask2":     config.IsTask2,
			"dataset":     dataset,
			"target":      target,
			"taskType":    taskType,
			"taskSubType": taskSubType,
			"metrics":     metrics,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal version into JSON and write response"))
			return
		}
	}
}
