package task

import (
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
)

// Segment segments an image into separate parts.
func Segment(dataset *api.Dataset) (string, error) {
	envConfig, err := env.LoadConfig()
	if err != nil {
		return "", err
	}

	datasetInputDir := env.ResolvePath(dataset.Source, dataset.Folder)

	step, err := description.CreateRemoteSensingSegmentationPipeline("segmentation", "basic image segmentation", envConfig.RemoteSensingNumJobs)
	if err != nil {
		return "", err
	}

	datasetURI, err := submitPipeline([]string{datasetInputDir}, step, true)
	if err != nil {
		return "", err
	}

	return datasetURI, nil
}
