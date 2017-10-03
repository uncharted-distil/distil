package routes

import (
	"net/http"
	"os"

	"github.com/unchartedsoftware/plog"
)

// AbortHandler terminates the server.  Yes, this is intentional.  Its part of the
// eval protocol.  Don't look at me like that.
func AbortHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Received abort request - shutting down")
		os.Exit(0)
		return
	}
}
