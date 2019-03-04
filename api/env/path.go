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

	seedPath = config.D3MInputDir
	seedSubPath = path.Join("TRAIN", "dataset_TRAIN")
	tmpPath = config.TmpDataPath

	contribPath = config.DatamartImportFolder
	augmentedPath = path.Join(config.TmpDataPath, config.AugmentedSubFolder)

	initialized = true

	return nil
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
