package routes

import (
	"github.com/pkg/errors"
	"net/http"
)

// VersionHandler returns the compiled version number and timestamp.
func VersionHandler(version string, timestamp string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// marshall version
		err := handleJSON(w, map[string]string{
			"version":   version,
			"timestamp": timestamp,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal version into JSON and write response"))
			return
		}

		return
	}
}
