package routes

import (
	"fmt"
	"net/http"

	"github.com/unchartedsoftware/plog"
	"goji.io/pat"
)

// EchoHandler generates a route a simple echo route handler for testing
// purposes.
func EchoHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Processing echo request")
		fmt.Fprintf(w, "Distil - %s", pat.Param(r, "echo"))
	}
}
