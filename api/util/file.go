package util

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"
)

// PathConfig contains basic configuration for path resolving.
type PathConfig struct {
	InputFolder     string
	InputSubFolders string
	OutputFolder    string
}

// PathResolver resolves a path given a basic configuration.
type PathResolver struct {
	Config *PathConfig
}

// NewPathResolver creates a new path resolver.
func NewPathResolver(config *PathConfig) *PathResolver {
	return &PathResolver{
		Config: config,
	}
}

// ResolveInputAbsolute creates the input path as an absolute.
func (r *PathResolver) ResolveInputAbsolute(relativePath string) string {
	return path.Join(r.Config.InputFolder, r.Config.InputSubFolders, relativePath)
}

// ResolveInputAbsoluteFromRoot creates the input path as an absolute= from the
// root datasets dir.
func (r *PathResolver) ResolveInputAbsoluteFromRoot(relativePath string) string {
	return path.Join(r.Config.InputFolder, relativePath, r.Config.InputSubFolders)
}

// ResolveOutputAbsolute creates the output path as an absolute.
func (r *PathResolver) ResolveOutputAbsolute(relativePath string) string {
	return path.Join(r.Config.OutputFolder, relativePath)
}

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

	os.MkdirAll(destination, os.ModePerm)

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
			os.MkdirAll(path, os.ModePerm)
		} else {
			os.MkdirAll(filepath.Dir(path), os.ModePerm)
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
