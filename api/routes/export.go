package routes

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"

	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/pipeline"
	"github.com/unchartedsoftware/plog"
)

// ExportHandler exports the caller supplied solution by calling through to the compute
// server export functionality.
func ExportHandler(solutionCtor model.SolutionStorageCtor, metaCtor model.MetadataStorageCtor, client *pipeline.Client, exportPath string) func(http.ResponseWriter, *http.Request) {
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
		pip, err := solution.FetchSolution(solutionID)
		if err != nil {
			handleError(w, err)
			return
		}

		m, err := solution.FetchRequest(pip.RequestID)
		if err != nil {
			handleError(w, err)
			return
		}

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		variable, err := meta.FetchVariable(m.Dataset, "datasets", solutionTarget)
		if err != nil {
			handleError(w, err)
			return
		}

		// fail if the solution target was not the original dataset target
		if variable.Role != "suggestedTarget" {
			log.Warnf("Target %s is not the expected target variable", variable.Name)
			http.Error(w, fmt.Sprintf("The selected target `%s` does not match the required target variable.", variable.Name), http.StatusBadRequest)
			return
		}

		exportPath := path.Join(exportPath, solutionID+".d3m")
		exportURI := fmt.Sprintf("file://%s", exportPath)
		log.Infof("Exporting to %s", exportURI)

		err = client.ExportSolution(context.Background(), solutionID, exportURI)
		if err != nil {
			log.Infof("Failed solution export request to %s", exportURI)
			os.Exit(1)
		} else {
			log.Infof("Completed export request to %s", exportURI)
			os.Exit(0)
		}
		return
	}
}
