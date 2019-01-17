package task

import (
	"bytes"
	"encoding/csv"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-ingest/metadata"

	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/description"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/result"
	"github.com/unchartedsoftware/distil/api/util"
)

// Merge will merge data resources into a single data resource.
func Merge(schemaFile string, index string, dataset string, config *IngestTaskConfig) (string, error) {
	outputPath, err := initializeDatasetCopy(schemaFile, dataset, config.MergedOutputSchemaPathRelative, config.MergedOutputPathRelative, config)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy source data folder")
	}

	// create & submit the solution request
	pip, err := description.CreateDenormalizePipeline("3NF", "")
	if err != nil {
		return "", errors.Wrap(err, "unable to create denormalize pipeline")
	}

	// pipeline execution assumes datasetDoc.json as schema file
	datasetURI, err := submitPipeline([]string{outputPath.sourceFolder}, pip)
	if err != nil {
		return "", errors.Wrap(err, "unable to run denormalize pipeline")
	}

	// parse primitive response (raw data from the input dataset)
	rawResults, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse denormalize result")
	}

	// need to manually build the metadata and output it.
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile)
	if err != nil {
		return "", errors.Wrap(err, "unable to load original metadata")
	}
	mainDR := meta.GetMainDataResource()
	vars := mapFields(meta)
	varsDenorm := mapDenormFields(mainDR)
	for k, v := range varsDenorm {
		vars[k] = v
	}

	outputMeta := model.NewMetadata(meta.ID, meta.Name, meta.Description)
	outputMeta.DataResources = append(outputMeta.DataResources, model.NewDataResource("0", mainDR.ResType, mainDR.ResFormat))
	header := rawResults[0]
	for i, field := range header {
		// the first column is a row idnex and should be discarded.
		if i > 0 {
			fieldName, ok := field.(string)
			if !ok {
				return "", errors.Errorf("unable to cast field name")
			}

			v := vars[fieldName]
			v.Index = i - 1
			outputMeta.DataResources[0].Variables = append(outputMeta.DataResources[0].Variables, v)
		}
	}

	// initialize csv writer
	output := &bytes.Buffer{}
	writer := csv.NewWriter(output)

	// returned header doesnt match expected header so use metadata header
	headerMetadata, err := outputMeta.GenerateHeaders()
	if err != nil {
		return "", errors.Wrapf(err, "unable to generate header")
	}
	writer.Write(headerMetadata[0])

	// rewrite the output without the first column
	rawResults = rawResults[1:]
	for _, line := range rawResults {
		lineString := make([]string, len(line)-1)
		for i := 1; i < len(line); i++ {
			lineString[i-1] = line[i].(string)
		}
		writer.Write(lineString)
	}

	// output the data
	writer.Flush()
	err = util.WriteFileWithDirs(outputPath.outputData, output.Bytes(), os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "error writing merged output")
	}

	relativePath := getRelativePath(path.Dir(outputPath.outputSchema), outputPath.outputData)
	outputMeta.DataResources[0].ResPath = relativePath

	// write the new schema to file
	err = metadata.WriteSchema(outputMeta, outputPath.outputSchema)
	if err != nil {
		return "", errors.Wrap(err, "unable to store merged schema")
	}

	return outputPath.outputSchema, nil
}
