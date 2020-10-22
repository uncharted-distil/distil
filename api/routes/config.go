//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package routes

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil/api/env"
)

// ConfigHandler returns the compiled version number, timestamp and initial config.
func ConfigHandler(config env.Config, version string, timestamp string, problemPath string, datasetDocPath string, ta2Version string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		target := "unknown"
		dataset := "unknown"
		var metrics []string

		// marshal version
		err := handleJSON(w, map[string]interface{}{
			"version":    version,
			"timestamp":  timestamp,
			"dataset":    dataset,
			"target":     target,
			"metrics":    metrics,
			"help":       config.HelpURL,
			"ta2version": ta2Version,
			"subdomains": config.Subdomains,
			"mapAPIKey": config.MapAPIKey,
			"tileRequestURL": config.TileRequestURL,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal version into JSON and write response"))
			return
		}
	}
}
