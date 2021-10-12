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

package env

import (
	"path"

	"github.com/uncharted-distil/distil-compute/metadata"
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
	batchPath     = ""
	publicPath    = ""
	resourcePath  = ""

	initialized = false
	isTask      = false
)

// Initialize the path resolution.
func Initialize(config *Config) error {
	log.Infof("initializing path values")
	if initialized {
		log.Warn("path resolution already initialized - ignoring")
		return nil
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
	batchPath = path.Join(config.D3MOutputDir, config.BatchSubFolder)
	publicPath = path.Join(config.D3MOutputDir, config.PublicSubFolder)
	resourcePath = path.Join(config.D3MOutputDir, config.ResourceSubFolder)

	log.Infof("using '%s' as seed path", seedPath)
	log.Infof("using '%s' as seed sub path", seedSubPath)
	log.Infof("using '%s' as tmp path", outputPath)
	log.Infof("using '%s' as contrib path", contribPath)
	log.Infof("using '%s' as augmented path", augmentedPath)
	log.Infof("using '%s' as batch path", batchPath)
	log.Infof("using '%s' as public path", publicPath)
	log.Infof("using '%s' as resource path", resourcePath)

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

// GetAugmentedPath returns the augmented path as initialized.
func GetAugmentedPath() string {
	return augmentedPath
}

// GetBatchPath returns the batch path as initialized.
func GetBatchPath() string {
	return batchPath
}

// GetPublicPath returns the public path as initialized.
func GetPublicPath() string {
	return publicPath
}

// GetResourcePath returns the resource path as initialized.
func GetResourcePath() string {
	return resourcePath
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
	case metadata.Batch:
		return resolveBatchPath(relativePath)
	case metadata.Public:
		return resolvePublicPath(relativePath)
	}
	return resolveTmpPath(relativePath)
}

func resolveSeedPath(relativePath string) string {
	if isTask {
		// if in task then the relative path is not relevant
		return path.Join(seedPath, seedSubPath)
	}
	return path.Join(seedPath, relativePath, seedSubPath)
}

func resolveContribPath(relativePath string) string {
	return path.Join(contribPath, relativePath)
}

func resolveAugmentedPath(relativePath string) string {
	return path.Join(augmentedPath, relativePath)
}

func resolveBatchPath(relativePath string) string {
	return path.Join(batchPath, relativePath)
}

func resolvePublicPath(relativePath string) string {
	return path.Join(publicPath, relativePath)
}

func resolveTmpPath(relativePath string) string {
	return path.Join(outputPath, relativePath)
}
