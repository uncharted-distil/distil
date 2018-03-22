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

// ExportHandler exports the caller supplied pipeline by calling through to the compute
// server export functionality.
func ExportHandler(pipelineCtor model.PipelineStorageCtor, metaCtor model.MetadataStorageCtor, client *pipeline.Client, exportPath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		pipelineID := pat.Param(r, "pipeline-id")
		sessionID := pat.Param(r, "session")

		// get the pipeline target
		pipeline, err := pipelineCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		res, err := pipeline.FetchResultMetadataByPipelineID(pipelineID)
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
		request, err := pipeline.FetchRequest(res.RequestID)
		if err != nil {
			handleError(w, err)
			return
		}

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		variable, err := meta.FetchVariable(request.Dataset, "datasets", pipelineTarget)
		if err != nil {
			handleError(w, err)
			return
		}

		// fail if the pipeline target was not the original dataset target
		if variable.Role != "suggestedTarget" {
			log.Warnf("Target %s is not the expected target variable", variable.Name)
			http.Error(w, fmt.Sprintf("The selected target `%s` does not match the required target variable.", variable.Name), http.StatusBadRequest)
			return
		}

		exportPath := path.Join(exportPath, pipelineID+".d3m")
		exportURI := fmt.Sprintf("file://%s", exportPath)
		log.Infof("Exporting to %s", exportURI)

		err = client.ExportPipeline(context.Background(), sessionID, pipelineID, exportURI)
		if err != nil {
			log.Infof("Failed pipeline export request to %s", exportURI)
			os.Exit(1)
		} else {
			log.Infof("Completed export request to %s", exportURI)
			os.Exit(0)
		}
		return
	}
}
