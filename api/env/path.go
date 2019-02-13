package env

import (
	"path"

	"github.com/pkg/errors"

	"github.com/unchartedsoftware/distil-ingest/metadata"
)

var (
	seedPath    = ""
	contribPath = ""

	seedSubPath   = ""
	augmentedPath = ""
)

// Initialize the path resolution.
func Initialize(config *Config) {
	seedPath = config.D3MInputDir
	seedSubPath = path.Join("TRAIN", "dataset_TRAIN")

	contribPath = config.DatamartImportFolder
	augmentedPath = path.Join(config.TmpDataPath, config.AugmentedSubFolder)
}

// ResolvePath returns an absolute path based on the dataset source.
func ResolvePath(datasetSource metadata.DatasetSource, relativePath string) (string, error) {
	switch datasetSource {
	case metadata.Seed:
		return resolveSeedPath(relativePath), nil
	case metadata.Contrib:
		return resolveContribPath(relativePath), nil
	case metadata.Augmented:
		return resolveAugmentedPath(relativePath), nil
	}
	return "", errors.Errorf("source %v cannot be resolved", datasetSource)
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
