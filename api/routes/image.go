package routes

import (
	"net/http"
	"path"

	"goji.io/pat"

	api "github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil/api/env"
	"github.com/unchartedsoftware/distil/api/model"
)

const (
	imageFolder = "media"
)

// ImageHandler provides a static file lookup route using simple directory mapping.
func ImageHandler(ctor model.MetadataStorageCtor, config *env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// resources can either be local or remote
		dataset := pat.Param(r, "dataset")
		source := pat.Param(r, "source")
		file := pat.Param(r, "file")
		path := path.Join(imageFolder, file)

		// get metadata client
		storage, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}

		res, err := storage.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		resolver := createResolverForResource(api.DatasetSource(source), res.Folder, config)

		bytes, err := fetchResourceBytes(resolver.ResolveInputAbsolute(""), dataset, path)
		if err != nil {
			handleError(w, err)
			return
		}
		w.Write(bytes)
	}
}
