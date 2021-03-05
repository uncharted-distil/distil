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

package dataset

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"

	"github.com/h2non/filetype"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

var (
	simpleFileTypes = map[string]string{"txt": "txt"}
)

// ExpandedDatasetPaths stores paths info about the input dataset archive
// and the expanded archive.
type ExpandedDatasetPaths struct {
	RawFilePath       string
	ExtractedFilePath string
}

// StoreZipDataset writes the archive file to temporary storage, where data is supplied as
// a byte array.
func StoreZipDataset(dataset string, rawData []byte) (string, error) {
	fileName := generateZipPath(dataset)
	err := util.WriteFileWithDirs(fileName, rawData, os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "unable to write raw image data archive")
	}
	return fileName, nil
}

// StoreZipDatasetFromFormFile writes the archive file to temporary storage, where data is supplied as
// a form multipart file.
func StoreZipDatasetFromFormFile(dataset string, formFile multipart.File) (string, error) {
	fileName := generateZipPath(dataset)
	err := util.WriteFormFileWithDirs(fileName, formFile, os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "unable to write raw image data archive")
	}
	return fileName, nil
}

func generateZipPath(dataset string) string {
	tmpPath := env.GetTmpPath()
	zipFileName := path.Join(tmpPath, fmt.Sprintf("%s_raw.zip", dataset))
	zipFileName = util.GetUniqueName(zipFileName)
	return zipFileName
}

// ExpandZipDataset decompresses a zipped dataset for further downstream processing.
func ExpandZipDataset(datasetPath string, datasetName string) (*ExpandedDatasetPaths, error) {
	tmpPath := env.GetTmpPath()
	extractedArchivePath := util.GetUniqueFolder(path.Join(tmpPath, datasetName))
	err := util.Unzip(datasetPath, extractedArchivePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to extract raw image data archive")
	}
	return &ExpandedDatasetPaths{datasetPath, extractedArchivePath}, nil
}

// CheckFileType does a breadth first directory traversal until it finds a file, then it checks
// its type and returns it to the caller.
func CheckFileType(extractedArchivePath string) (string, error) {
	dirQueue := []string{extractedArchivePath}
	for len(dirQueue) > 0 {
		currentDir := dirQueue[0]
		dirQueue = dirQueue[1:]
		files, err := ioutil.ReadDir(currentDir)
		if err != nil {
			return "", errors.Wrap(err, "cannot check file type")
		}
		for _, f := range files {
			if !f.IsDir() {
				buf, err := ioutil.ReadFile(path.Join(currentDir, f.Name()))
				if err != nil {
					log.Error(err)
					continue
				}
				// check simple extention names since the library doesnt handle them
				ext := path.Ext(f.Name())
				if len(ext) > 1 {
					ext = ext[1:]
				}
				if simpleFileTypes[ext] != "" {
					return simpleFileTypes[ext], nil
				}

				kind, err := filetype.Match(buf)
				if err != nil {
					log.Error(err)
					continue
				}
				return kind.Extension, nil
			}
			dirQueue = append(dirQueue, path.Join(currentDir, f.Name()))
		}
	}
	return "", nil
}
