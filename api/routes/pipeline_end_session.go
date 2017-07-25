package routes

import (
	"net/http"

	"goji.io/pat"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/unchartedsoftware/distil/api/pipeline"
)

// PipelineEndSessionHandler creates a route to handle pipeline session end
// requests.
func PipelineEndSessionHandler(client *pipeline.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse session id
		sessionID := pat.Param(r, "session-id")
		// end the session
		err := client.EndSession(context.Background(), sessionID)
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to fetch session ID"))
			return
		}
		// write the session
		err = handleJSON(w, SessionResult{
			ID: sessionID,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session result into JSON"))
			return
		}
	}
}
