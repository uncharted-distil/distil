//
//   Copyright © 2021 Uncharted Software Inc.
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
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
)

// Merge will merge data resources into a single data resource.
func Merge(schemaFile string, dataset string, config *IngestTaskConfig) (string, error) {
	outputPath := createDatasetPaths(schemaFile, dataset, compute.D3MLearningData)

	// need to manually build the metadata and output it.
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to load original metadata")
	}

	// create & submit the solution request
	var pip *description.FullySpecifiedPipeline
	timeseries, _, _ := isTimeseriesDataset(meta)
	if timeseries {
		pip, err = description.CreateTimeseriesFormatterPipeline("Time Cop", "", compute.DefaultResourceID)
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
	datasetURI, err := submitPipeline([]string{outputPath.sourceFolder}, pip, true)
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
	vars := mapHeaderFields(meta)
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
			v = model.NewVariable(i, field, field, field, field, model.StringType, model.StringType, "", []string{"attribute"}, model.VarDistilRoleData, nil, outputMeta.DataResources[0].Variables, false)
		}
		v.Index = i
		outputMeta.DataResources[0].Variables = append(outputMeta.DataResources[0].Variables, v)
	}

	// returned header doesnt match expected header so use metadata header
	headerMetadata, err := outputMeta.GenerateHeaders()
	if err != nil {
		return "", errors.Wrapf(err, "unable to generate header")
	}
	output := [][]string{headerMetadata[0]}
	output = append(output, csvData[1:]...)

	// output the data
	datasetStorage := serialization.GetStorage(outputPath.outputData)
	err = datasetStorage.WriteData(outputPath.outputData, output)
	if err != nil {
		return "", errors.Wrap(err, "error writing merged output")
	}
	outputMeta.DataResources[0].ResPath = outputPath.outputData

	// add every source data resource that isnt the main data resource to not lose them
	for _, dr := range meta.DataResources {
		if dr != mainDR {
			outputMeta.DataResources = append(outputMeta.DataResources, dr)
		}
	}

	// write the new schema to file
	err = datasetStorage.WriteMetadata(outputPath.outputSchema, outputMeta, true, false)
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
