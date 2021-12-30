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
)

// Clean will clean bad data for further processing.
func Clean(schemaFile string, dataset string, params *IngestParams, config *IngestTaskConfig) (string, error) {
	outputPath := createDatasetPaths(schemaFile, dataset, compute.D3MLearningData)
	pip, err := createCleaningPipeline(schemaFile, dataset, params, config)
	if err != nil {
		return "", err
	}

	// pipeline execution assumes datasetDoc.json as schema file
	datasetURI, err := submitPipeline([]string{outputPath.sourceFolder}, pip.steps, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to run format pipeline")
	}

	// output the header
	outputPathURI, err := processCleaningPipelineOutput(schemaFile, dataset)(datasetURI)
	if err != nil {
		return "", err
	}

	return outputPathURI, nil
}

func processCleaningPipelineOutput(schemaFile string, dataset string) func(string) (string, error) {
	return func(datasetURI string) (string, error) {
		outputPath := createDatasetPaths(schemaFile, dataset, compute.D3MLearningData)
		// load metadata from original schema
		meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile, true)
		if err != nil {
			return "", errors.Wrap(err, "unable to load original schema file")
		}
		mainDR := meta.GetMainDataResource()

		// output the header
		output := [][]string{}
		header := make([]string, len(mainDR.Variables))
		for _, v := range mainDR.Variables {
			header[v.Index] = v.HeaderName
		}
		output = append(output, header)

		// parse primitive response (raw data from the input dataset)
		// first row of the data is the header
		// first column of the data is the dataframe index
		readStorage := serialization.GetStorage(datasetURI)
		csvData, err := readStorage.ReadData(datasetURI)
		if err != nil {
			return "", errors.Wrap(err, "unable to parse clean result")
		}
		output = append(output, csvData[1:]...)

		// output the data
		datasetStorage := serialization.GetStorage(outputPath.outputData)
		err = datasetStorage.WriteData(outputPath.outputData, output)
		if err != nil {
			return "", errors.Wrap(err, "error writing clustered output")
		}
		mainDR.ResPath = outputPath.outputData

		// write the new schema to file
		err = datasetStorage.WriteMetadata(outputPath.outputSchema, meta, true, false)
		if err != nil {
			return "", errors.Wrap(err, "unable to store cluster schema")
		}

		return outputPath.outputSchema, nil
	}
}

func createCleaningPipeline(schemaFile string, dataset string, params *IngestParams, config *IngestTaskConfig) (*Pipeline, error) {
	// load metadata from original schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load original schema file")
	}
	mainDR := meta.GetMainDataResource()

	metaStorage, err := params.MetaCtor()
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize metadata storage")
	}

	vars := []*model.Variable{}
	exists, _ := metaStorage.DatasetExists(meta.ID)
	if exists {
		vars, err = metaStorage.FetchVariables(meta.ID, false, false, false)
		if err != nil {
			return nil, err
		}
	} else if params.DefinitiveTypes != nil {
		for _, v := range mainDR.Variables {
			if params.DefinitiveTypes[v.Key] != nil {
				clone := v.Clone()
				clone.Type = params.DefinitiveTypes[v.Key].Type
				vars = append(vars, clone)
			}
		}
	}

	// create & submit the solution request
	pip, err := description.CreateDataCleaningPipeline("Mary Poppins", "", vars, config.ImputeEnabled)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create format pipeline")
	}

	return &Pipeline{
		shouldCache:           true,
		steps:                 pip,
		resultParsingCallback: map[string]func(string) (string, error){"first": processCleaningPipelineOutput(schemaFile, dataset)},
	}, err
}
