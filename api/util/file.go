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

package util

import (
	"archive/zip"
	"encoding/csv"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"
)

// WriteFileWithDirs writes the file and creates any missing directories along
// the way.
func WriteFileWithDirs(filename string, data []byte, perm os.FileMode) error {

	dir, _ := filepath.Split(filename)

	// make all dirs up to the destination
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	// write the file
	return ioutil.WriteFile(filename, data, perm)
}

// Unzip extracts an archive to the given location.
func Unzip(zipFile string, destination string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return errors.Wrap(err, "unable to open archive")
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	err = os.MkdirAll(destination, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to make containing directory")
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return errors.Wrap(err, "unable to open archived file")
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(destination, f.Name)

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "unable to make archive directory")
			}
		} else {
			err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "unable to make file directory")
			}
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "unable to write archived file")
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return errors.Wrap(err, "unable to copy archived file")
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return errors.Wrap(err, "unable to extract files")
		}
	}

	return nil
}

// Copy copies a source folder to a destination folder.
func Copy(sourceFolder string, destinationFolder string) error {
	// copy the source folder to have all the linked files for merging
	err := copy.Copy(sourceFolder, destinationFolder)
	if err != nil {
		return errors.Wrap(err, "unable to copy source data")
	}

	return nil
}

// RemoveContents removes the files and directories from the supplied parent.
func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// IsDatasetDir indicate whether or not a directory contains a single d3m dataset.
func IsDatasetDir(dir string) bool {
	datasetPath := path.Join(dir, "TRAIN", "dataset_TRAIN")
	_, err := os.Stat(datasetPath)
	return !os.IsNotExist(err)
}

// GetDirectories returns a list of directories found using the supplied path.
func GetDirectories(inputPath string) ([]string, error) {
	files, err := ioutil.ReadDir(inputPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list directory content")
	}

	dirs := make([]string, 0)
	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, path.Join(inputPath, f.Name()))
		}
	}

	return dirs, nil
}

// ReadCSVFile reads a csv file and returns the string slice representation of the data.
func ReadCSVFile(filename string, hasHeader bool) ([][]string, error) {
	// open the file
	csvFile, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open data file")
	}
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)

	lines := make([][]string, 0)

	// skip the header as needed
	if hasHeader {
		_, err = reader.Read()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read header from file")
		}
	}

	// read the raw data
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Wrap(err, "failed to read line from file")
		}

		lines = append(lines, line)
	}

	return lines, nil
}
