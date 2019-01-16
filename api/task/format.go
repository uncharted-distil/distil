package task

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-ingest/metadata"

	"github.com/unchartedsoftware/distil/api/util"
)

// Format will format a dataset to have the required structures for D3M.
func Format(schemaFile string, config *IngestTaskConfig) (string, error) {
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile)
	if err != nil {
		return "", errors.Wrap(err, "unable to load original schema file")
	}

	// fix for d3m index requirement
	path := path.Dir(schemaFile)
	if !checkD3MIndexExists(meta) {
		path, err = addD3MIndex(schemaFile, meta, config)
		if err != nil {
			return "", errors.Wrap(err, "unable to load original schema file")
		}
	}

	return path, nil
}

func addD3MIndex(schemaFile string, meta *model.Metadata, config *IngestTaskConfig) (string, error) {
	// check to make sure only a single data resource exists
	if len(meta.DataResources) != 1 {
		return "", errors.Errorf("adding d3m index requires that the dataset have only 1 data resource (%d exist)", len(meta.DataResources))
	}

	// copy the data to a new directory
	outputPath, err := initializeDatasetCopy(schemaFile, config.FormatOutputSchemaRelative, config.FormatOutputDataRelative, config)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy source data folder")
	}

	// add the d3m index variable to the metadata
	dr := meta.DataResources[0]
	name := model.D3MIndexFieldName
	v := model.NewVariable(len(dr.Variables), name, name, name, model.IndexType, model.IndexType, []string{"index"}, model.VarRoleIndex, nil, dr.Variables, false)
	dr.Variables = append(dr.Variables, v)

	// read the raw data
	dataPath := path.Join(path.Dir(schemaFile), dr.ResPath)
	lines, err := ReadCSVFile(dataPath, config.HasHeader)
	if err != nil {
		return "", errors.Wrap(err, "error reading raw data")
	}

	// append the row count as d3m index
	// initialize csv writer
	output := &bytes.Buffer{}
	writer := csv.NewWriter(output)

	// output the header
	header := make([]string, len(dr.Variables))
	for _, v := range dr.Variables {
		header[v.Index] = v.Name
	}
	err = writer.Write(header)
	if err != nil {
		return "", errors.Wrap(err, "error storing format header")
	}

	// parse the raw output and write the line out
	for i, line := range lines {
		line = append(line, fmt.Sprintf("%d", i+1))

		err = writer.Write(line)
		if err != nil {
			return "", errors.Wrap(err, "error storing feature output")
		}
	}

	// output the data with the new feature
	writer.Flush()
	err = util.WriteFileWithDirs(config.GetTmpAbsolutePath(config.FormatOutputDataRelative), output.Bytes(), os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "error writing feature output")
	}

	relativePath := getRelativePath(path.Dir(outputPath.outputSchema), outputPath.outputData)
	dr.ResPath = relativePath

	// write the new schema to file
	err = metadata.WriteSchema(meta, config.GetTmpAbsolutePath(config.FormatOutputSchemaRelative))
	if err != nil {
		return "", errors.Wrap(err, "unable to store feature schema")
	}

	return path.Dir(config.GetTmpAbsolutePath(config.FormatOutputSchemaRelative)), nil
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
