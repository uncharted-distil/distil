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
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"

	"github.com/uncharted-distil/distil/api/util"
)

var (
	pythonDictRE = regexp.MustCompile("'([^'\"]*)'")
)

// Featurize will featurize the dataset fields using a primitive.
func Featurize(datasetSource metadata.DatasetSource, schemaFile string, index string, dataset string, config *IngestTaskConfig) (string, error) {
	outputPath, err := initializeDatasetCopy(schemaFile, dataset, config.FeaturizationOutputSchemaRelative, config.FeaturizationOutputDataRelative)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy source data folder")
	}

	// load metadata from original schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile)
	if err != nil {
		return "", errors.Wrap(err, "unable to load original schema file")
	}
	mainDR := meta.GetMainDataResource()

	// add feature variables
	features, err := getFeatureVariables(meta, model.FeatureVarPrefix)
	if err != nil {
		return "", errors.Wrap(err, "unable to get feature variables")
	}

	d3mIndexField := getD3MIndexField(mainDR)

	// open the input file
	dataPath := path.Join(outputPath.sourceFolder, mainDR.ResPath)
	lines, err := util.ReadCSVFile(dataPath, config.HasHeader)
	if err != nil {
		return "", errors.Wrap(err, "error reading raw data")
	}

	// add the cluster data to the raw data
	for _, f := range features {
		mainDR.Variables = append(mainDR.Variables, f.Variable)

		// header already removed, lines does not have a header
		lines, err = appendFeature(outputPath.outputFolder, d3mIndexField, false, f, lines)
		if err != nil {
			return "", errors.Wrap(err, "error appending feature data")
		}
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
		return "", errors.Wrap(err, "error storing feature header")
	}

	// parse the raw output and write the line out
	for _, line := range lines {
		if len(features) > 0 {
			fieldIndex := len(line) - 1
			p, err := parseFeatureOutput(line[fieldIndex])
			if err != nil {
				return "", errors.Wrap(err, "unable to parse raw feature output")
			}
			line[fieldIndex] = p
		}

		err = writer.Write(line)
		if err != nil {
			return "", errors.Wrap(err, "error storing feature output")
		}
	}

	// output the data with the new feature
	writer.Flush()
	err = util.WriteFileWithDirs(outputPath.outputData, output.Bytes(), os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "error writing feature output")
	}

	relativePath := getRelativePath(path.Dir(outputPath.outputSchema), outputPath.outputData)
	mainDR.ResPath = relativePath

	// write the new schema to file
	err = metadata.WriteSchema(meta, outputPath.outputSchema, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to store feature schema")
	}

	return outputPath.outputSchema, nil
}

func parseFeatureOutput(field string) (string, error) {
	fieldAugmented := pythonDictRE.ReplaceAllString(field, "\"$1\"")
	parsed := make(map[string]interface{})
	err := json.Unmarshal([]byte(fieldAugmented), &parsed)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse raw output field")
	}

	joined := ""
	for _, v := range parsed {
		joined = fmt.Sprintf("%s,%s", joined, v.(string))
	}

	return joined[1:], nil
}
