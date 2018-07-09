package routes

import (
	"github.com/pkg/errors"
	"net/http"

	"github.com/unchartedsoftware/distil/api/env"
)

// ConfigHandler returns the compiled version number, timestamp and initial config.
func ConfigHandler(config env.Config, version string, timestamp string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		//
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read problem file")
		}

		problemInfo := &compute.ProblemPersist{}
		err = json.Unmarshal(b, problemInfo)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to unmarshal classification response")
		}
		//

		func parseMetrics(filename string) ([]string, error) {

		parseMetrics(problemFile)
		.Inputs.Data.DatasetID
		.Inputs.Data.Targets[i].ColName


		// marshall version
		err := handleJSON(w, map[string]interface{}{
			"version":   version,
			"timestamp": timestamp,
			"discovery": config.IsProblemDiscovery,
			"dataset": ,
			"target"

		})




		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal version into JSON and write response"))
			return
		}

		return
	}
}
