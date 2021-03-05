//
//   Copyright © 2021 Uncharted Software Inc.
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
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil/api/dataset"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/rest"
	"github.com/uncharted-distil/distil/api/util"
)

// UploadHandler uploads a file to the local file system and then imports it.
func UploadHandler(config *env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetName := pat.Param(r, "dataset")

		// type cant be a post param since the upload is the actual data
		queryValues := r.URL.Query()
		typ := queryValues.Get("type")

		var outputPath string
		var err error
		if typ == "datamart" {
			var params map[string]interface{}
			params, err = getPostParameters(r)
			if err != nil {
				handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
				return
			}

			urlString := params["url"].(string)
			if isValidDownloadURL(urlString, config) {
				outputPath, err = downloadFile(datasetName, urlString, config)
			} else {
				err = errors.Errorf("supplied url is invalid")
			}
		} else {
			// read the file from the request
			log.Infof("Reading byte stream from http request")

			// Create a form file reader.  It will be combined with a file writer in a subsequent step.
			var formFile multipart.File
			formFile, _, err = r.FormFile("file")
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to receive file from request"))
			}
			defer formFile.Close()

			// Figure out what type of dataset we've got
			if typ == "table" {
				log.Infof("Uploaded table dataset '%s'", datasetName)
				tmpPath := env.GetTmpPath()
				csvFilename := path.Join(tmpPath, fmt.Sprintf("%s_raw.csv", datasetName))
				outputPath = util.GetUniqueName(csvFilename)
				err = util.WriteFormFileWithDirs(outputPath, formFile, os.ModePerm)
			} else if typ == "media" {
				// Expand the data into temp storage
				log.Infof("Uploaded zipped media dataset '%s'", datasetName)
				outputPath, err = dataset.StoreZipDatasetFromFormFile(datasetName, formFile)
			} else if typ == "" {
				err = errors.Errorf("upload type parameter not specified")
			} else {
				err = errors.Errorf("unrecognized upload type")
			}
		}

		if err != nil {
			handleError(w, errors.Wrap(err, "unable to create raw dataset"))
			return
		}

		log.Infof("Uploaded new dataset %s at %s", datasetName, outputPath)
		// marshal data and sent the response back
		err = handleJSON(w, map[string]interface{}{"dataset": datasetName, "location": outputPath, "result": "uploaded"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}

func isValidDownloadURL(urlString string, config *env.Config) bool {
	u, err := url.Parse(urlString)
	if err != nil {
		return false
	}

	if u.Scheme == "" || u.Host == "" || u.Path == "" {
		return false
	}

	// check to make sure it comes from datamart
	ud, err := url.Parse(config.DatamartURINYU)
	if err != nil {
		return false
	}
	if u.Host != ud.Host {
		return false
	}

	return true
}

func downloadFile(datasetName string, urlString string, config *env.Config) (string, error) {
	// get the file
	restClient := rest.NewClient("")()
	fileData, err := restClient.Get(urlString, nil)
	if err != nil {
		return "", err
	}

	// store it to the augmented folder
	outputPath := env.ResolvePath(metadata.Augmented, datasetName)
	outputPath = util.GetUniqueName(outputPath)
	err = ioutil.WriteFile(outputPath, fileData, os.ModePerm)
	if err != nil {
		return "", errors.Wrapf(err, "unable to write downloaded dataset to the file system")
	}

	return outputPath, nil
}
