package task

import (
	"path"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/description"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/result"
)

// Clean will clean bad data for further processing.
func Clean(schemaFile string, index string, dataset string, config *IngestTaskConfig) (string, error) {
	// copy the data to a new directory
	outputPath, err := initializeDatasetCopy(schemaFile, path.Base(path.Dir(schemaFile)), config.CleanOutputSchemaRelative, config.CleanOutputDataRelative, config)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy source data folder")
	}

	// create & submit the solution request
	pip, err := description.CreateDataCleaningPipeline("Mary Poppins", "")
	if err != nil {
		return "", errors.Wrap(err, "unable to create format pipeline")
	}

	// pipeline execution assumes datasetDoc.json as schema file
	datasetURI, err := submitPipeline([]string{outputPath.sourceFolder}, pip)
	if err != nil {
		return "", errors.Wrap(err, "unable to run format pipeline")
	}

	// parse primitive response (raw data from the input dataset)
	_, err = result.ParseResultCSV(datasetURI)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse format result")
	}

	return path.Dir(outputPath.outputSchema), nil
}
