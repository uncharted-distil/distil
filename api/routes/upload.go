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
	"io"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"goji.io/pat"

	"github.com/uncharted-distil/distil/api/task"
)

// UploadHandler uploads a file to the local file system and then imports it.
func UploadHandler(outputPath string, config *task.IngestTaskConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		typ := pat.Param(r, "type")

		var err error
		formattedPath := ""
		if typ == "table" {
			formattedPath, err = uploadTableDataset(dataset, outputPath, config, r)
		} else if typ == "image" {
			formattedPath, err = uploadImageDataset(dataset, outputPath, config, r)
		}

		if err != nil {
			handleError(w, errors.Wrap(err, "unable to upload dataset"))
			return
		}

		log.Infof("uploaded new dataset %s at %s", dataset, formattedPath)
		// marshal data and sent the response back
		err = handleJSON(w, map[string]interface{}{"result": "uploaded"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}

func uploadTableDataset(dataset string, outputPath string, config *task.IngestTaskConfig, r *http.Request) (string, error) {
	// read the file from the request
	bytes, err := receiveFile(r)
	if err != nil {
		return "", errors.Wrap(err, "unable to receive file from request")
	}

	// create the raw dataset schema doc
	formattedPath, err := task.CreateDataset(dataset, bytes, outputPath, config)
	if err != nil {
		return "", errors.Wrap(err, "unable to create dataset")
	}

	return formattedPath, nil
}

func uploadImageDataset(dataset string, outputPath string, config *task.IngestTaskConfig, r *http.Request) (string, error) {
	// parse POST params
	params, err := getPostParameters(r)
	if err != nil {
		return "", errors.Wrap(err, "Unable to parse post parameters")
	}

	foldersRaw, ok := params["folders"].([]interface{})
	if !ok {
		return "", errors.Errorf("unable to parse 'folders' parameter")
	}

	typ, ok := params["type"].(string)
	if !ok {
		return "", errors.Errorf("unable to parse 'type' parameter")
	}
	folders := make([]string, 0)
	for _, f := range foldersRaw {
		folders = append(folders, f.(string))
	}

	// create the raw dataset schema doc
	formattedPath, err := task.CreateImageDataset(dataset, folders, typ, outputPath, config)
	if err != nil {
		return "", errors.Wrap(err, "unable to create dataset")
	}

	return formattedPath, nil
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
