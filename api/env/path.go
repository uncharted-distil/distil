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
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-ingest/metadata"

	"github.com/uncharted-distil/distil/api/util"
)

const (
	seedFolderName = "seed_datasets_current"
)

var (
	seedPath    = ""
	contribPath = ""
	tmpPath     = ""

	seedSubPath   = ""
	augmentedPath = ""

	initialized = false
)

// Initialize the path resolution.
func Initialize(config *Config) error {
	if initialized {
		return errors.Errorf("path resolution already initialized")
	}

	updatedInputPath, err := determineSeedPath(config.D3MInputDir)
	if err != nil {
		return err
	}

	seedPath = updatedInputPath
	seedSubPath = path.Join("TRAIN", "dataset_TRAIN")
	tmpPath = config.TmpDataPath

	contribPath = config.DatamartImportFolder
	augmentedPath = path.Join(config.TmpDataPath, config.AugmentedSubFolder)

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
	return tmpPath
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
	return path.Join(seedPath, relativePath, seedSubPath)
}

func resolveContribPath(relativePath string) string {
	return path.Join(contribPath, relativePath)
}

func resolveAugmentedPath(relativePath string) string {
	return path.Join(augmentedPath, relativePath)
}

func resolveTmpPath(relativePath string) string {
	return path.Join(tmpPath, relativePath)
}
