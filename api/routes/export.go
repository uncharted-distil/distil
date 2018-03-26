package routes

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/pipeline"
	"github.com/unchartedsoftware/plog"
)

// ExportHandler exports the caller supplied pipeline by calling through to the compute
// server export functionality.
func ExportHandler(storageCtor model.PipelineStorageCtor, metaStorageCtor model.MetadataStorageCtor, client *pipeline.Client, exportPath string, problem *pipeline.Problem) func(http.ResponseWriter, *http.Request) {

	// ** jan eval only
	var targetVar string
	var targetTask string

	if problem != nil {
		targetVar = problem.Inputs.Data[0].Targets[0].ColName
		targetTask = problem.Properties.TaskType
	}
	// ** end jan eval only

	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		pipelineID := pat.Param(r, "pipeline-id")
		sessionID := pat.Param(r, "session")

		// get the pipeline target
		pipelineStorage, err := storageCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		res, err := pipelineStorage.FetchResultMetadataByPipelineID(pipelineID)
		if err != nil {
			handleError(w, err)
			return
		}

		pipelineTarget := ""
		for _, f := range res.Features {
			if f.FeatureType == model.FeatureTypeTarget {
				pipelineTarget = f.FeatureName
			}
		}

		// get the initial target
		request, err := pipelineStorage.FetchRequest(res.RequestID)
		if err != nil {
			handleError(w, err)
			return
		}

		metaStorage, err := metaStorageCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		variable, err := metaStorage.FetchVariable(request.Dataset, "datasets", pipelineTarget)
		if err != nil {
			handleError(w, err)
			return
		}

		// ** jan eval only

		// fail if the pipeline target was not the expected dataset target for the problem
		if targetVar != "" && strings.ToUpper(variable.Name) != strings.ToUpper(targetVar) {
			log.Warnf("Target %s is not the expected target variable %s", variable.Name, targetVar)
			http.Error(w, fmt.Sprintf("The selected target `%s` does not match the required target variable `%s`.", variable.Name, targetVar), http.StatusBadRequest)
			return
		}

		// fail if the pipeline task does not match the task for the problem - this isn't not currently stored
		// in the request, so we hack it
		if targetTask != "" {
			if model.IsNumerical(variable.Type) && strings.ToUpper(targetTask) != "REGRESSION" {
				http.Error(w, fmt.Sprintf("Target type `%s` expected to be one of `Categorical`, `Ordinal` or `Boolean` for this task.", variable.Type), http.StatusBadRequest)
				log.Warnf("Target var type of %s is not of the expected type for problem task `%s`", variable.Type, targetTask)
				return
			} else if model.IsCategorical(variable.Type) && strings.ToUpper(targetTask) != "CLASSIFICATION" {
				http.Error(w, fmt.Sprintf("Target type `%s` expected to be one `Integer` or `Float` for this task.", variable.Type), http.StatusBadRequest)
				log.Warnf("Target var type of %s is not of the expected type for problem task `%s`", variable.Type, targetTask)
				return
			} else if model.IsText(variable.Type) {
				http.Error(w, fmt.Sprintf("Target type of `%s` unsupported for this task", variable.Type), http.StatusBadRequest)
				log.Warnf("Target var type of %s is not an allowed target type", variable.Type)
				return
			}
		}
		// ** end jan eval only

		exportPath := path.Join(exportPath, pipelineID+".d3m")
		exportURI := fmt.Sprintf("file://%s", exportPath)
		log.Infof("Exporting to %s", exportURI)

		err = client.ExportPipeline(context.Background(), sessionID, pipelineID, exportURI)
		if err == nil {
			log.Infof("Failed pipeline export request to %s", exportURI)
			os.Exit(1)
		} else {
			log.Infof("Completed export request to %s", exportURI)
			os.Exit(0)
		}
		return
	}
}
