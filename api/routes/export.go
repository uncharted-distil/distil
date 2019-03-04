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
	"os"

	"goji.io/pat"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/model"
	"github.com/unchartedsoftware/plog"
)

// ExportHandler exports the caller supplied solution by calling through to the compute
// server export functionality.
func ExportHandler(solutionCtor model.SolutionStorageCtor, metaCtor model.MetadataStorageCtor, client *compute.Client, exportPath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		solutionID := pat.Param(r, "solution-id")

		// get the solution target
		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// get the initial target
		sol, err := solution.FetchSolution(solutionID)
		if err != nil {
			handleError(w, err)
			return
		}

		// export relies on a fitted model
		fittedSolutionID := sol.Result.FittedSolutionID
		if fittedSolutionID == "" {
			handleError(w, errors.Errorf("export failed - no fitted solution found for solution %s", solutionID))
			return
		}

		err = client.ExportSolution(context.Background(), fittedSolutionID)
		if err != nil {
			log.Infof("Failed solution export request for %s", fittedSolutionID)
			os.Exit(1)
		} else {
			log.Infof("Completed export request for %s", fittedSolutionID)
			os.Exit(0)
		}
		return
	}
}
