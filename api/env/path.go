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

package env

import (
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-ingest/metadata"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil/api/util"
)

const (
	seedFolderName = "seed_datasets_current"
)

var (
	seedPath    = ""
	contribPath = ""
	outputPath  = ""

	seedSubPath   = ""
	augmentedPath = ""

	initialized = false
	isTask      = false
)

// Initialize the path resolution.
func Initialize(config *Config) error {
	log.Infof("initializing path values")
	if initialized {
		return errors.Errorf("path resolution already initialized")
	}

	updatedInputPath, err := determineSeedPath(config.D3MInputDir)
	if err != nil {
		return err
	}

	seedPath = updatedInputPath
	seedSubPath = path.Join("TRAIN", "dataset_TRAIN")
	outputPath = config.D3MOutputDir

	contribPath = config.DatamartImportFolder
	augmentedPath = path.Join(config.D3MOutputDir, config.AugmentedSubFolder)

	isTask = config.IsTask1 || config.IsTask2

	log.Infof("using '%s' as seed path", seedPath)
	log.Infof("using '%s' as seed sub path", seedSubPath)
	log.Infof("using '%s' as tmp path", outputPath)
	log.Infof("using '%s' as contrib path", contribPath)
	log.Infof("using '%s' as augmented path", augmentedPath)
	log.Infof("isTask set to '%v'", isTask)

	initialized = true

	return nil
}

func determineSeedPath(inputPath string) (string, error) {
	// the input can be either:
	//   1. a dataset folder
	//   2. a folder containing dataset folders
	//   3. a parent folder of the seed dataset folder ('seed_datasets_current')
	if util.IsDatasetDir(inputPath) {
		return inputPath, nil
	}

	// get the list of folders
	dirs, err := util.GetDirectories(inputPath)
	if err != nil {
		return "", err
	}

	// empty folder
	if len(dirs) == 0 {
		return inputPath, nil
	}

	// if one folder is a dataset, then assume they all are
	if util.IsDatasetDir(dirs[0]) {
		return inputPath, nil
	}

	// find the seed dataset subfolder
	return findSeedDatasetDirectory(inputPath)
}

func findSeedDatasetDirectory(inputPath string) (string, error) {
	dirs, err := util.GetDirectories(inputPath)
	if err != nil {
		return "", err
	}

	// look for the seed dataset folder
	for _, d := range dirs {
		if path.Base(d) == seedFolderName {
			return d, nil
		}
	}

	// look in the subfolders
	for _, d := range dirs {
		dir, err := findSeedDatasetDirectory(d)
		if err != nil {
			return "", err
		}
		if dir != "" {
			return dir, nil
		}
	}

	// not found
	return "", nil
}

// GetTmpPath returns the tmp path as initialized.
func GetTmpPath() string {
	return outputPath
}

// ResolvePath returns an absolute path based on the dataset source.
func ResolvePath(datasetSource metadata.DatasetSource, relativePath string) string {
	switch datasetSource {
	case metadata.Seed:
		return resolveSeedPath(relativePath)
	case metadata.Contrib:
		return resolveContribPath(relativePath)
	case metadata.Augmented:
		return resolveAugmentedPath(relativePath)
	}
	return resolveTmpPath(relativePath)
}

func resolveSeedPath(relativePath string) string {
	dirToUse := seedPath
	if isTask {
		// when running in task mode (1 or 2), there can be overlap between paths
		// since task mode sets the seed path to a specific dataset directory
		dir, file := path.Split(seedPath)
		relativeSplits := strings.Split(relativePath, string(os.PathSeparator))
		if len(file) > 0 && relativeSplits[0] == file {
			// overlap between relative path and seed path needs to be removed
			dirToUse = dir
			log.Infof("removing '%s' path overlap when resolving", file)
		}
	}
	return path.Join(dirToUse, relativePath, seedSubPath)
}

func resolveContribPath(relativePath string) string {
	return path.Join(contribPath, relativePath)
}

func resolveAugmentedPath(relativePath string) string {
	return path.Join(augmentedPath, relativePath)
}

func resolveTmpPath(relativePath string) string {
	return path.Join(outputPath, relativePath)
}
