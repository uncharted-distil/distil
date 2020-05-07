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

package dataset

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/h2non/filetype"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

type ExpandedDataset struct {
	RawFilePath       string
	ExtractedFilePath string
}

// ExpandZipDataset decompresses a zipped dataset for further downstream processing.
func ExpandZipDataset(dataset string, rawData []byte) (*ExpandedDataset, error) {
	// store and expand raw data
	tmpPath := env.GetTmpPath()
	zipFilename := path.Join(tmpPath, fmt.Sprintf("%s_raw.zip", dataset))
	zipFilename = getUniqueName(zipFilename)
	err := util.WriteFileWithDirs(zipFilename, rawData, os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "unable to write raw image data archive")
	}

	extractedArchivePath := getUniqueFolder(path.Join(tmpPath, dataset))
	err = util.Unzip(zipFilename, extractedArchivePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to extract raw image data archive")
	}
	return &ExpandedDataset{zipFilename, extractedArchivePath}, nil
}

// CheckFileType does a bread first directory traversal until it finds a file, then it checks
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
				kind, err := filetype.Match(buf)
				if err != nil {
					log.Error(err)
					continue
				}
				return kind.Extension, nil
			} else {
				dirQueue = append(dirQueue, path.Join(currentDir, f.Name()))
			}
		}
	}
	return "", nil
}

func getUniqueName(filename string) string {
	extension := path.Ext(filename)
	baseFilename := strings.TrimSuffix(filename, extension)
	currentFilename := filename
	for i := 1; util.FileExists(currentFilename); i++ {
		currentFilename = fmt.Sprintf("%s_%d.%s", baseFilename, i, extension)
	}

	return currentFilename
}

func getUniqueFolder(folder string) string {
	currentFilename := folder
	for i := 1; util.FileExists(currentFilename); i++ {
		currentFilename = fmt.Sprintf("%s_%d", folder, i)
	}

	return currentFilename
}
