package util

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
