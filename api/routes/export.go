package routes

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/pipeline"
	"github.com/unchartedsoftware/plog"
)

// ExportHandler exports the caller supplied pipeline by calling through to the compute
// server export functionality.
func ExportHandler(client *pipeline.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		pipelineID := pat.Param(r, "pipeline-id")
		sessionID := pat.Param(r, "session")
		exportURI := fmt.Sprintf("file:///%s.d3m", pipelineID)

		err := client.ExportPipeline(context.Background(), sessionID, pipelineID, exportURI)
		if err != nil {
			log.Info("Failed pipeline export request to %s", exportURI)
		} else {
			log.Info("Completed export request to %s", exportURI)
		}
		os.Exit(0)
		return
	}
}
