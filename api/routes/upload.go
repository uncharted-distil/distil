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
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil/api/dataset"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
)

// UploadHandler uploads a file to the local file system and then imports it.
func UploadHandler(outputPath string, config *env.Config) func(http.ResponseWriter, *http.Request) {
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

		var ds task.DatasetConstructor
		if typ == "table" {
			ds, err = uploadTableDataset(datasetName, outputPath, config, data)
		} else if typ == "image" {
			imageType, ok := queryValues["image"]
			if !ok {
				handleError(w, errors.Errorf("unable to parse 'image' parameter"))
				return
			}

			ds, err = uploadImageDataset(datasetName, outputPath, config, data, imageType[0])
		} else if typ == "remote" {
			imageType, ok := queryValues["image"]
			if !ok {
				handleError(w, errors.Errorf("unable to parse 'image' parameter"))
				return
			}

			ds, err = uploadRemoteSensingDataset(datasetName, outputPath, config, data, imageType[0])
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

		// create the raw dataset schema doc
		formattedPath, err := task.CreateDataset(datasetName, ds, outputPath, api.DatasetTypeModelling, config)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to create d3m dataset"))
			return
		}

		if err != nil {
			handleError(w, errors.Wrap(err, "unable to upload dataset"))
			return
		}

		log.Infof("uploaded new dataset %s at %s", datasetName, formattedPath)
		// marshal data and sent the response back
		err = handleJSON(w, map[string]interface{}{"result": "uploaded"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}

func uploadTableDataset(datasetName string, outputPath string, config *env.Config, data []byte) (task.DatasetConstructor, error) {
	ds, err := dataset.NewTableDataset(datasetName, data, config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create raw dataset")
	}

	return ds, nil
}

func uploadImageDataset(datasetName string, outputPath string, config *env.Config, data []byte, imageType string) (task.DatasetConstructor, error) {
	ds, err := dataset.NewImageDataset(datasetName, imageType, data, config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create raw dataset")
	}

	return ds, nil
}

func uploadRemoteSensingDataset(datasetName string, outputPath string, config *env.Config, data []byte, imageType string) (task.DatasetConstructor, error) {
	ds, err := dataset.NewSatelliteDataset(datasetName, imageType, data, config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create raw dataset")
	}

	return ds, nil
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
