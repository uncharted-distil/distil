package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/unchartedsoftware/distil/api/pipeline"
)

// SessionResult represents the result of a session start request.
type SessionResult struct {
	ID string `json:"id"`
}

// PipelineSessionHandler creates a route to handle pipeline create requests.
func PipelineSessionHandler(client *pipeline.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// start the session
		session, err := client.StartSession(context.Background())
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to fetch session ID"))
			return
		}
		// write the session
		err = handleJSON(w, SessionResult{
			ID: session.ID,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
