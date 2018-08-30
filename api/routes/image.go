package routes

import (
	"net/http"
	"path"

	"goji.io/pat"
)

const (
	imageFolder = "media"
)

// ImageHandler provides a static file lookup route using simple directory mapping.
func ImageHandler(resourceDir string, proxyServer string, proxy map[string]bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// resources can either be local or remote
		dataset := pat.Param(r, "dataset")
		file := pat.Param(r, "file")
		path := path.Join(imageFolder, file)

		bytes, err := fetchResourceBytes(resourceDir, proxyServer, proxy, dataset, path)
		if err != nil {
			handleError(w, err)
			return
		}
		w.Write(bytes)
	}
}
