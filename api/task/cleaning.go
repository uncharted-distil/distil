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
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"

	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
)

// Clean will clean bad data for further processing.
func Clean(schemaFile string, dataset string, config *IngestTaskConfig) (string, error) {
	outputPath := createDatasetPaths(schemaFile, dataset, compute.D3MLearningData)

	// load metadata from original schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile, true)
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

	// output the header
	output := [][]string{}
	header := make([]string, len(mainDR.Variables))
	for _, v := range mainDR.Variables {
		header[v.Index] = v.StorageName
	}
	output = append(output, header)

	// parse primitive response (raw data from the input dataset)
	// first row of the data is the header
	// first column of the data is the dataframe index
	csvData, err := util.ReadCSVFile(datasetURI, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse clean result")
	}
	output = append(output, csvData...)

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
