package routes

import (
	"net/http"
)

// FileHandler provides a static file lookup route using the OS file system
func FileHandler(rootDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(rootDir)).ServeHTTP(w, r)
	}
}
