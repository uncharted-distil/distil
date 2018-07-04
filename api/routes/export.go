package routes

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"goji.io/pat"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/compute"
	"github.com/unchartedsoftware/distil/api/model"
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
		req, err := solution.FetchRequestBySolutionID(solutionID)
		if err != nil {
			handleError(w, err)
			return
		}

		solutionTarget := req.TargetFeature()

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

		m, err := solution.FetchRequest(sol.RequestID)
		if err != nil {
			handleError(w, err)
			return
		}

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		variable, err := meta.FetchVariable(m.Dataset, solutionTarget)
		if err != nil {
			handleError(w, err)
			return
		}

		// fail if the solution target was not the original dataset target
		if variable.Role != "suggestedTarget" {
			log.Warnf("Target %s is not the expected target variable", variable.Key)
			http.Error(w, fmt.Sprintf("The selected target `%s` does not match the required target variable.", variable.Key), http.StatusBadRequest)
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
