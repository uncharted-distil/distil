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
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
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

// WriteFormFileWithDirs writes the form multipart file and creates any missing directories along
// the way.
func WriteFormFileWithDirs(filename string, formFile multipart.File, perm os.FileMode) error {

	dir, _ := filepath.Split(filename)

	// make all dirs up to the destination
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, perm)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer file.Close()
	_, err = io.Copy(file, formFile)
	if err != nil {
		return errors.Wrap(err, "failed to write form file")
	}
	return nil
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

// CopyFile the source file to destination. Any existing file will be overwritten and will not
// copy file attributes.
func CopyFile(sourceFile string, destinationFile string) error {
	in, err := os.Open(sourceFile)
	if err != nil {
		return errors.Wrap(err, "unable to open source file")
	}
	defer in.Close()

	// check if the target directory exists and create it if not
	destinationDir := path.Dir(destinationFile)
	if !FileExists(destinationDir) {
		err = os.MkdirAll(destinationDir, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "unable to make destination folder")
		}
	}

	out, err := os.Create(destinationFile)
	if err != nil {
		return errors.Wrap(err, "unable to create destination file")
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return errors.Wrap(err, "unable to copy file")
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

// IsDatasetDir indicates whether or not a directory contains a single d3m dataset.
func IsDatasetDir(dir string) bool {
	datasetPath := path.Join(dir, "TRAIN", "dataset_TRAIN")
	if IsDirectory(datasetPath) {
		dir = datasetPath
	}

	// check for dataset doc
	schemaPath := path.Join(dir, "datasetDoc.json")
	return FileExists(schemaPath)
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

// ReadCSVHeader reads the first line of a CSV file.
func ReadCSVHeader(filename string) ([]string, error) {
	// open the file
	csvFile, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open data file")
	}
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = 0

	header, err := reader.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read header from file")
	}

	return header, nil
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
	reader.FieldsPerRecord = 0

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
			log.Warnf("failed to read line - %v", err)
			continue
		}

		lines = append(lines, line)
	}

	return lines, nil
}

// GenerateTimeFileNameStr generates an ISO 8601 representation of current time, with
// any colons and dashes removed to that it can be used as part of a filename.
func GenerateTimeFileNameStr() string {
	timeStr := time.Now().Format(time.RFC3339)
	timeStr = strings.ReplaceAll(timeStr, ":", "")
	timeStr = strings.ReplaceAll(timeStr, "-", "")
	return timeStr
}

// FileExists checks if a file already exists on disk.
func FileExists(filename string) bool {
	// check that the directory of the file exists as a directory
	// if the "directory" of the file is actually a file, then the simple
	// stat check fails with an unrelated error
	if !IsDirectory(path.Dir(filename)) {
		return false
	}

	_, err := os.Stat(filename)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}

// IsDirectory checks if a path is a directory.
func IsDirectory(filePath string) bool {
	fi, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// GetFolderFileType returns the extension of the first file in the media folder.
func GetFolderFileType(folder string) (string, error) {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read folder")
	}
	if len(files) == 0 {
		return "", nil
	}

	extension := path.Ext(files[0].Name())
	if len(extension) > 0 {
		extension = extension[1:]
	}

	return extension, nil
}

// IsArchiveFile returns true if the specified path is an archive.
func IsArchiveFile(filePath string) bool {
	return path.Ext(filePath) == ".zip" || path.Ext(filePath) == ".tar.gz"
}

// GetUniqueName creates a unique filename using a base filename.
func GetUniqueName(filename string) string {
	extension := path.Ext(filename)
	baseFilename := strings.TrimSuffix(filename, extension)
	currentFilename := filename
	for i := 1; FileExists(currentFilename); i++ {
		currentFilename = fmt.Sprintf("%s_%d%s", baseFilename, i, extension)
	}

	return currentFilename
}

// GetUniqueFolder creates a unique folder name using a base folder name.
func GetUniqueFolder(folder string) string {
	currentFilename := folder
	for i := 1; FileExists(currentFilename); i++ {
		currentFilename = fmt.Sprintf("%s_%d", folder, i)
	}

	return currentFilename
}
