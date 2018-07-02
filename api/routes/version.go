package routes

import (
	"github.com/pkg/errors"
	"net/http"

	"github.com/unchartedsoftware/distil/api/env"
)

// ConfigHandler returns the compiled version number, timestamp and initial config.
func ConfigHandler(config env.Config, version string, timestamp string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// marshall version
		err := handleJSON(w, map[string]interface{}{
			"version":   version,
			"timestamp": timestamp,
			"discovery": config.IsProblemDiscovery,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal version into JSON and write response"))
			return
		}

		return
	}
}
