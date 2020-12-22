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
	"io/ioutil"
	"net/http"
	"path"

	"goji.io/v3/pat"

	"github.com/pkg/errors"
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
		file := pat.Param(r, "file")

		// get metadata client
		storage, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}

		res, err := storage.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		sourcePath := env.ResolvePath(res.Source, res.Folder)

		bytes, err := ioutil.ReadFile(path.Join(sourcePath, imageFolder, file))
		if err != nil {
			handleError(w, err)
			return
		}
		_, err = w.Write(bytes)
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to write image resource bytes"))
			return
		}
	}
}
