package routes

import (
	"github.com/pkg/errors"
	"net/http"

	"github.com/unchartedsoftware/distil/api/compute"
	"github.com/unchartedsoftware/distil/api/env"
)

// ConfigHandler returns the compiled version number, timestamp and initial config.
func ConfigHandler(config env.Config, version string, timestamp string, problemPath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		target := "unknown"
		dataset := "unknown"
		// load problem file
		problem, err := compute.LoadProblemSchemaFromFile(problemPath)
		if err == nil {
			if len(problem.Inputs.Data) > 0 {
				dataset = problem.Inputs.Data[0].DatasetID
				if len(problem.Inputs.Data[0].Targets) > 0 {
					target = problem.Inputs.Data[0].Targets[0].ColName
				}
			}
		}

		// marshall version
		err = handleJSON(w, map[string]interface{}{
			"version":   version,
			"timestamp": timestamp,
			"discovery": config.IsProblemDiscovery,
			"dataset":   dataset,
			"target":    target,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal version into JSON and write response"))
			return
		}

		return
	}
}
