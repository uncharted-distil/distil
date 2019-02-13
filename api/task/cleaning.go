package task

import (
	"bytes"
	"encoding/csv"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/description"
	"github.com/unchartedsoftware/distil-ingest/metadata"

	"github.com/unchartedsoftware/distil/api/util"
)

// Clean will clean bad data for further processing.
func Clean(datasetSource metadata.DatasetSource, schemaFile string, index string, dataset string, config *IngestTaskConfig) (string, error) {
	// copy the data to a new directory
	outputPath, err := initializeDatasetCopy(datasetSource, schemaFile, dataset, config.CleanOutputSchemaRelative, config.CleanOutputDataRelative)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy source data folder")
	}

	// load metadata from original schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile)
	if err != nil {
		return "", errors.Wrap(err, "unable to load original schema file")
	}
	mainDR := meta.GetMainDataResource()

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

	// initialize csv writer
	output := &bytes.Buffer{}
	writer := csv.NewWriter(output)

	// output the header
	header := make([]string, len(mainDR.Variables))
	for _, v := range mainDR.Variables {
		header[v.Index] = v.Name
	}
	err = writer.Write(header)
	if err != nil {
		return "", errors.Wrap(err, "error storing clean header")
	}

	// parse primitive response (raw data from the input dataset)
	// first row of the data is the header
	// first column of the data is the dataframe index
	csvData, err := ReadCSVFile(datasetURI, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse clean result")
	}

	// need to remove the first column of the output (row index)
	for _, res := range csvData {
		err = writer.Write(res[1:])
		if err != nil {
			return "", errors.Wrap(err, "error storing clean data")
		}
	}

	// output the data with the new feature
	writer.Flush()

	err = util.WriteFileWithDirs(outputPath.outputData, output.Bytes(), os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "error writing clustered output")
	}

	relativePath := getRelativePath(path.Dir(outputPath.outputSchema), outputPath.outputData)
	mainDR.ResPath = relativePath

	// write the new schema to file
	err = metadata.WriteSchema(meta, outputPath.outputSchema)
	if err != nil {
		return "", errors.Wrap(err, "unable to store cluster schema")
	}

	return outputPath.outputSchema, nil
}
