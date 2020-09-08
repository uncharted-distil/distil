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

package task

import (
	"fmt"
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"

	"github.com/uncharted-distil/distil/api/util"
)

// Format will format a dataset to have the required structures for D3M.
func Format(schemaFile string, dataset string, config *IngestTaskConfig) (string, error) {
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to load original schema file")
	}
	dr := getMainDataResource(meta)

	// copy the data to a new directory
	outputPath, err := initializeDatasetCopy(schemaFile, dataset, config.FormatOutputSchemaRelative, config.FormatOutputDataRelative)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy source data folder")
	}

	// read the raw data
	dataPath := path.Join(path.Dir(schemaFile), dr.ResPath)
	lines, err := util.ReadCSVFile(dataPath, config.HasHeader)
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
	dr := getMainDataResource(meta)

	// output the header
	header := make([]string, len(dr.Variables))
	for _, v := range dr.Variables {
		header[v.Index] = v.Name
	}
	output := [][]string{header}
	output = append(output, lines...)

	// output the data with the new feature
	err := datasetStorage.WriteData(paths.outputData, output)
	if err != nil {
		return errors.Wrap(err, "error writing output")
	}

	relativePath := getRelativePath(path.Dir(paths.outputSchema), paths.outputData)
	dr.ResPath = relativePath
	dr.ResType = model.ResTypeTable

	// write the new schema to file
	err = datasetStorage.WriteMetadata(paths.outputSchema, meta, true)
	if err != nil {
		return errors.Wrap(err, "unable to store schema")
	}

	return nil
}

func addD3MIndex(schemaFile string, meta *model.Metadata, data [][]string) (*model.Metadata, [][]string, error) {
	// add the d3m index variable to the metadata
	dr := getMainDataResource(meta)
	name := model.D3MIndexFieldName
	v := model.NewVariable(len(dr.Variables), name, name, name, model.IntegerType, model.IntegerType,
		"required index field", []string{model.RoleIndex}, model.VarDistilRoleIndex, nil, dr.Variables, false)
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

func getMainDataResource(meta *model.Metadata) *model.DataResource {
	dr := meta.GetMainDataResource()
	if dr == nil {
		dr = meta.DataResources[0]
	}

	return dr
}
