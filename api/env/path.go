package env

import (
	"path"

	"github.com/unchartedsoftware/distil-ingest/metadata"
)

var (
	seedPath    = ""
	contribPath = ""
	tmpPath     = ""

	seedSubPath   = ""
	augmentedPath = ""
)

// Initialize the path resolution.
func Initialize(config *Config) {
	seedPath = config.D3MInputDir
	seedSubPath = path.Join("TRAIN", "dataset_TRAIN")
	tmpPath = config.TmpDataPath

	contribPath = config.DatamartImportFolder
	augmentedPath = path.Join(config.TmpDataPath, config.AugmentedSubFolder)
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
