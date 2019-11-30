//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package routes

import (
	"net/http"
	"path"

	"goji.io/v3/pat"

	api "github.com/uncharted-distil/distil-ingest/pkg/metadata"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/model"
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

		sourcePath := env.ResolvePath(api.DatasetSource(source), res.Folder)

		bytes, err := fetchResourceBytes(sourcePath, dataset, path)
		if err != nil {
			handleError(w, err)
			return
		}
		w.Write(bytes)
	}
}
