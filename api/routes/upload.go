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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil/api/dataset"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/util"
)

// UploadHandler uploads a file to the local file system and then imports it.
func UploadHandler(config *env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetName := pat.Param(r, "dataset")

		// type cant be a post param since the upload is the actual data
		queryValues := r.URL.Query()
		typ := queryValues.Get("type")

		// read the file from the request
		data, err := receiveFile(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to receive file from request"))
			return
		}
		// Figure out what type of dataset we've got
		var outputPath string
		if typ == "table" {
			tmpPath := env.GetTmpPath()
			csvFilename := path.Join(tmpPath, fmt.Sprintf("%s_raw.csv", datasetName))
			outputPath = util.GetUniqueName(csvFilename)
			err = util.WriteFileWithDirs(csvFilename, data, os.ModePerm)
		} else if typ == "media" {
			// Expand the data into temp storage
			outputPath, err = dataset.StoreZipDataset(datasetName, data)
		} else if typ == "" {
			handleError(w, errors.Errorf("upload type parameter not specified"))
			return
		} else {
			handleError(w, errors.Errorf("unrecognized upload type"))
			return
		}

		if err != nil {
			handleError(w, errors.Wrap(err, "unable to create raw dataset"))
			return
		}

		log.Infof("uploaded new dataset %s at %s", datasetName, outputPath)
		// marshal data and sent the response back
		err = handleJSON(w, map[string]interface{}{"dataset": datasetName, "location": outputPath, "result": "uploaded"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}

func receiveFile(r *http.Request) ([]byte, error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get file from request")
	}
	defer file.Close()

	// Copy the file data to the buffer
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return nil, errors.Wrap(err, "unable to copy file")
	}

	return buf.Bytes(), nil
}
