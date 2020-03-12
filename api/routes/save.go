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
	"context"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/primitive/compute"
	log "github.com/unchartedsoftware/plog"
)

func parseBoolParam(value string) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		parsed = false
	}

	return parsed
}

// SaveHandler exports the caller supplied solution by calling through to the compute
// server export functionality.
func SaveHandler(client *compute.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		solutionID := pat.Param(r, "solution-id")
		fitted := pat.Param(r, "fitted")
		fittedBool := parseBoolParam(fitted)

		var err error
		if fittedBool {
			err = client.SaveFittedSolution(context.Background(), solutionID)
		} else {
			err = client.SaveSolution(context.Background(), solutionID)
		}

		if err != nil {
			handleError(w, errors.Wrap(err, "failed solution export request"))
			return
		} else {
			log.Infof("Completed export request for %s", solutionID)
		}

		err = handleJSON(w, map[string]interface{}{"solution-id": solutionID, "result": "saved"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal save result into JSON"))
			return
		}
	}
}
