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
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-ingest/pkg/metadata"
	comp "github.com/uncharted-distil/distil/api/compute"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

// Predict processes input data to generate predictions.
func Predict(meta *model.Metadata, dataset string, fittedSolutionID string,
	csvData []byte, outputPath string, index string, metaStorage api.MetadataStorage, config *IngestTaskConfig) (string, error) {
	log.Infof("generating predictions for fitted solution ID %s", fittedSolutionID)
	// create the dataset to be used for predictions
	datasetPath, err := CreateDataset(dataset, csvData, outputPath, config)
	if err != nil {
		return "", err
	}
	log.Infof("created dataset for new data")

	// read the header of the new dataset to get the field names
	// if they dont match the original, then cant use the same pipeline
	rawDataPath := path.Join(datasetPath, compute.D3MDataFolder, compute.D3MLearningData)
	rawCSVData, err := util.ReadCSVFile(rawDataPath, false)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse header result")
	}
	rawHeader := rawCSVData[0]
	for i, f := range rawHeader {
		// TODO: may have to check the name rather than display name
		// TODO: col index not necessarily the same as index and thats what needs checking
		if meta.DataResources[0].Variables[i].DisplayName != f {
			return "", errors.Errorf("variables in new prediction file do not match variables in original dataset")
		}
	}
	log.Infof("dataset fields match original dataset fields")

	// update the dataset doc to reflect original types
	sourceDatasetID := meta.ID
	meta.ID = dataset
	meta.StorageName = model.NormalizeDatasetID(dataset)
	meta.DatasetFolder = path.Base(datasetPath)
	schemaPath := path.Join(datasetPath, compute.D3MDataSchema)
	err = metadata.WriteSchema(meta, schemaPath)
	if err != nil {
		return "", errors.Wrap(err, "unable to update dataset doc")
	}
	log.Infof("wrote out schema doc for new dataset")

	// ingest the dataset but without running simon, duke, etc.
	_, err = Ingest(schemaPath, schemaPath, metaStorage, index, dataset, metadata.Contrib, nil, config, false)
	if err != nil {
		return "", errors.Wrap(err, "unable to ingest ranked data")
	}
	log.Infof("finished ingesting the dataset")

	// the dataset id needs to matched the original dataset id for TA2 to be able to use the model
	meta.ID = sourceDatasetID
	err = metadata.WriteSchema(meta, schemaPath)
	if err != nil {
		return "", errors.Wrap(err, "unable to update dataset doc")
	}

	// submit the new dataset for predictions
	resultURI, err := comp.GeneratePredictions(datasetPath, fittedSolutionID, client)
	if err != nil {
		return "", err
	}
	log.Infof("generated predictions stored at %s", resultURI)

	return resultURI, nil
}
