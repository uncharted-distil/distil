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
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-ingest/metadata"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/util"
)

// Merge will merge data resources into a single data resource.
func Merge(datasetSource metadata.DatasetSource, schemaFile string, index string, dataset string, config *IngestTaskConfig) (string, error) {
	outputPath, err := initializeDatasetCopy(schemaFile, dataset, config.MergedOutputSchemaPathRelative, config.MergedOutputPathRelative)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy source data folder")
	}

	// need to manually build the metadata and output it.
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile)
	if err != nil {
		return "", errors.Wrap(err, "unable to load original metadata")
	}

	// create & submit the solution request
	var pip *pipeline.PipelineDescription
	timeseries, linkResID, _ := isTimeseriesDataset(meta)
	if timeseries {
		pip, err = description.CreateTimeseriesFormatterPipeline("Time Cop", "", linkResID)
		if err != nil {
			return "", errors.Wrap(err, "unable to create denormalize pipeline")
		}
	} else {
		pip, err = description.CreateDenormalizePipeline("3NF", "")
		if err != nil {
			return "", errors.Wrap(err, "unable to create denormalize pipeline")
		}
	}

	// pipeline execution assumes datasetDoc.json as schema file
	datasetURI, err := submitPipeline([]string{outputPath.sourceFolder}, pip)
	if err != nil {
		return "", errors.Wrap(err, "unable to run denormalize pipeline")
	}

	// parse primitive response (raw data from the input dataset)
	// first row of the data is the header
	csvData, err := util.ReadCSVFile(datasetURI, false)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse denormalize result")
	}

	mainDR := meta.GetMainDataResource()
	vars := mapFields(meta)
	varsDenorm := mapDenormFields(mainDR)
	for k, v := range varsDenorm {
		vars[k] = v
	}

	outputMeta := model.NewMetadata(meta.ID, meta.Name, meta.Description, meta.StorageName)
	outputMeta.DataResources = append(outputMeta.DataResources, model.NewDataResource(mainDR.ResID, mainDR.ResType, mainDR.ResFormat))
	header := csvData[0]
	for i, field := range header {
		v := vars[field]
		if v == nil {
			// create new variables (ex: series_id)
			v = model.NewVariable(i, field, field, field, model.StringType, model.StringType, []string{"attribute"}, model.VarRoleData, nil, outputMeta.DataResources[0].Variables, false)
		}
		v.Index = i
		outputMeta.DataResources[0].Variables = append(outputMeta.DataResources[0].Variables, v)
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

	// rewrite the output
	csvData = csvData[1:]
	for _, line := range csvData {
		writer.Write(line)
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

func isTimeseriesDataset(meta *model.Metadata) (bool, string, int) {
	mainDR := meta.GetMainDataResource()

	// check references to see if any point to a time series
	for _, v := range mainDR.Variables {
		if v.RefersTo != nil {
			resID := v.RefersTo["resID"].(string)
			res := getResource(meta, resID)
			if res != nil && res.ResType == "timeseries" {
				return true, resID, v.Index
			}
		}
	}

	return false, "", -1
}

func getResource(meta *model.Metadata, resID string) *model.DataResource {
	for _, dr := range meta.DataResources {
		if dr.ResID == resID {
			return dr
		}
	}

	return nil
}
