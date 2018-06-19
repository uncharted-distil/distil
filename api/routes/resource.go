package routes

import (
	"net/http"
)

// ResourceHandler provides a static file lookup route using simple directory mapping.
func ResourceHandler(resourceDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(resourceDir)).ServeHTTP(w, r)
	}
}
