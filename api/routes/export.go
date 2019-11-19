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

	"goji.io/v3/pat"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

// ExportHandler exports the caller supplied solution by calling through to the compute
// server export functionality.
func ExportHandler(client *compute.Client, exportPath string, solutionCtor model.SolutionStorageCtor, logger *env.DiscoveryLogger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		solutionID := pat.Param(r, "solution-id")

		// need to map to the initial search solution id since that is the pipeline
		// we need to export
		solutionStorage, err := solutionCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to connect to solution storage"))
			return
		}
		solution, err := solutionStorage.FetchSolution(solutionID)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to get stored solution information"))
			return
		}

		err = client.ExportSolution(context.Background(), solution.InitialSearchSolutionID)
		if err != nil {
			log.Infof("Failed solution export request for %s", solution.InitialSearchSolutionID)
		} else {
			log.Infof("Completed export request for %s", solution.InitialSearchSolutionID)
		}

		_, err = logger.InitializeLog("event-" + util.GenerateTimeFileNameStr() + ".csv")
		if err != nil {
			log.Infof("error initializing log after export: %v", err)
		}
	}
}
