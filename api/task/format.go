package task

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-ingest/metadata"

	"github.com/uncharted-distil/distil/api/util"
)

// Format will format a dataset to have the required structures for D3M.
func Format(datasetSource metadata.DatasetSource, schemaFile string, config *IngestTaskConfig) (string, error) {
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile)
	if err != nil {
		return "", errors.Wrap(err, "unable to load original schema file")
	}

	// check to make sure only a single data resource exists
	if len(meta.DataResources) != 1 {
		return "", errors.Errorf("formatting requires that the dataset have only 1 data resource (%d exist)", len(meta.DataResources))
	}
	dr := meta.DataResources[0]

	// copy the data to a new directory
	outputPath, err := initializeDatasetCopy(schemaFile, path.Base(path.Dir(schemaFile)), config.FormatOutputSchemaRelative, config.FormatOutputDataRelative)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy source data folder")
	}

	// read the raw data
	dataPath := path.Join(path.Dir(schemaFile), dr.ResPath)
	lines, err := ReadCSVFile(dataPath, config.HasHeader)
	if err != nil {
		return "", errors.Wrap(err, "error reading raw data")
	}

	// fix for d3m index requirement
	if !checkD3MIndexExists(meta) {
		meta, lines, err = addD3MIndex(schemaFile, meta, lines)
		if err != nil {
			return "", errors.Wrap(err, "unable to load original schema file")
		}
	}

	// output the data
	err = outputDataset(outputPath, meta, lines)
	if err != nil {
		return "", errors.Wrap(err, "unable to store formatted dataset")
	}

	return path.Dir(outputPath.outputSchema), nil
}

func outputDataset(paths *datasetCopyPath, meta *model.Metadata, lines [][]string) error {
	dr := meta.GetMainDataResource()

	// append the row count as d3m index
	// initialize csv writer
	output := &bytes.Buffer{}
	writer := csv.NewWriter(output)

	// output the header
	header := make([]string, len(dr.Variables))
	for _, v := range dr.Variables {
		header[v.Index] = v.Name
	}
	err := writer.Write(header)
	if err != nil {
		return errors.Wrap(err, "error storing header")
	}

	// output the formatted data
	err = writer.WriteAll(lines)

	// output the data with the new feature
	writer.Flush()
	err = util.WriteFileWithDirs(paths.outputData, output.Bytes(), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "error writing output")
	}

	relativePath := getRelativePath(path.Dir(paths.outputSchema), paths.outputData)
	dr.ResPath = relativePath
	dr.ResType = model.ResTypeTable

	// write the new schema to file
	err = metadata.WriteSchema(meta, paths.outputSchema)
	if err != nil {
		return errors.Wrap(err, "unable to store schema")
	}

	return nil
}

func addD3MIndex(schemaFile string, meta *model.Metadata, data [][]string) (*model.Metadata, [][]string, error) {
	// add the d3m index variable to the metadata
	dr := meta.DataResources[0]
	name := model.D3MIndexFieldName
	v := model.NewVariable(len(dr.Variables), name, name, name, model.IntegerType, model.IntegerType, []string{"index"}, model.VarRoleIndex, nil, dr.Variables, false)
	dr.Variables = append(dr.Variables, v)

	// parse the raw output and write the line out
	for i, line := range data {
		line = append(line, fmt.Sprintf("%d", i+1))
		data[i] = line
	}

	return meta, data, nil
}

func checkD3MIndexExists(meta *model.Metadata) bool {
	// check all variables for a d3m index
	for _, dr := range meta.DataResources {
		for _, v := range dr.Variables {
			if v.Name == model.D3MIndexFieldName {
				return true
			}
		}
	}

	return false
}
