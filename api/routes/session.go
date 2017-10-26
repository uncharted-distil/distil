package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// Session represents a session response
type Session struct {
	Pipelines []*model.Request `json:"pipelines"`
}

// SessionHandler fetches existing pipelines for a session.
func SessionHandler(storageCtor model.PipelineStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		sessionID := pat.Param(r, "session")

		client, err := storageCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		results, err := client.FetchRequests(sessionID)
		if err != nil {
			handleError(w, err)
			return
		}

		// Blank the result URI.
		for _, req := range results {
			for _, res := range req.Results {
				res.ResultURI = ""
			}
		}

		// marshall data and sent the response back
		err = handleJSON(w, Session{
			Pipelines: results,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session pipelines into JSON"))
			return
		}

		return
	}
}
