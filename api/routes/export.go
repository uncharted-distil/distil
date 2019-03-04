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
