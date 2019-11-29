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
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil/api/task"
)

// PredictionsHandler receives a file and produces results using the specified
// fitted solution id
func PredictionsHandler(outputPath string, config *task.IngestTaskConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		fittedSolutionID := pat.Param(r, "fitted-solution-id")

		// read the file from the request
		data, err := receiveFile(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to receive file from request"))
			return
		}
		log.Infof("received data to use for predictions for dataset %s solution %s", dataset, fittedSolutionID)

		_, err = task.Predict(dataset, fittedSolutionID, data)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to generate predictions"))
			return
		}

		// marshal data and sent the response back
		err = handleJSON(w, map[string]interface{}{"result": "uploaded"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
