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
	"math"
	"path"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/compute"
	"github.com/uncharted-distil/distil/api/util"
)

// Sample takes a sample of the dataset since larger datasets can lead to broken
// user experience through long lasting TA2 processing.
func Sample(originalSchemaFile string, schemaFile string, dataset string, config *IngestTaskConfig) (string, bool, int, error) {
	// load metadata from original schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile, true)
	if err != nil {
		return "", false, 0, errors.Wrap(err, "unable to load schema file")
	}
	mainDR := meta.GetMainDataResource()

	// extract a sample by simply reading the main CSV file and selecting a subset
	csvFilePath := path.Join(path.Dir(schemaFile), mainDR.ResPath)
	originalCSVFilePath := path.Join(path.Dir(originalSchemaFile), mainDR.ResPath)
	csvData, err := util.ReadCSVFile(csvFilePath, false)
	if err != nil {
		return "", false, 0, errors.Wrap(err, "unable to parse complete csv dataset")
	}

	if len(csvData) <= config.SampleRowLimit {
		return schemaFile, false, len(csvData), nil
	}

	sampledData := compute.SampleDataset(csvData, config.SampleRowLimit, true)

	// copy the full csv to keep it if needed
	err = util.CopyFile(originalCSVFilePath, path.Join(path.Dir(originalCSVFilePath), "learningData-full.csv"))
	if err != nil {
		return "", false, 0, err
	}

	// output to the expected location (learningData.csv)
	err = datasetStorage.WriteData(csvFilePath, sampledData)
	if err != nil {
		return "", false, 0, err
	}
	err = util.CopyFile(csvFilePath, originalCSVFilePath)
	if err != nil {
		return "", false, 0, err
	}

	// if csv data > max sample count, then assume max sample count is the count of
	// rows sampled in the output.
	rowCount := int(math.Min(float64(len(csvData)), float64(len(sampledData)))) - 1

	return schemaFile, rowCount < (len(csvData) - 1), rowCount, nil
}

func canSample(schemaFile string, config *IngestTaskConfig) bool {
	meta, err := metadata.LoadMetadataFromClassification(schemaFile, path.Join(path.Dir(schemaFile), config.ClassificationOutputPathRelative), true, false)
	if err != nil {
		log.Warnf("unable to load schema file for sampling: %+v", err)
		return false
	}

	for _, dr := range meta.DataResources {
		for _, v := range dr.Variables {
			if model.IsRemoteSensing(v.Type) || model.IsTimeSeries(v.Type) || model.IsGeoBounds(v.Type) {
				return false
			}
		}
	}

	return true
}
