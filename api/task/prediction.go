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

import "github.com/uncharted-distil/distil/api/compute"

// Predict processes input data to generate predictions.
func Predict(dataset string, fittedSolutionID string, csvData []byte, outputPath string, config *IngestTaskConfig) (string, error) {
	// create the dataset to be used for predictions
	path, err := CreateDataset(dataset, csvData, outputPath, config)
	if err != nil {
		return "", nil
	}

	// submit the new dataset for predictions
	resultURI, err := compute.GeneratePredictions(path, fittedSolutionID, client)
	if err != nil {
		return "", err
	}

	return resultURI, nil
}
