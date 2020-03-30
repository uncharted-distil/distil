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
	"context"

	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// SaveFittedSolution saves a fitted solution to disk via TA2TA3 API.
func SaveFittedSolution(fittedSolutionID string, solutionStorage api.SolutionStorage, metadataStorage api.MetadataStorage) (*api.ExportedModel, error) {
	uri, err := client.SaveFittedSolution(context.Background(), fittedSolutionID)
	if err != nil {
		return nil, err
	}

	request, err := solutionStorage.FetchRequestByFittedSolutionID(fittedSolutionID)
	if err != nil {
		return nil, err
	}

	metadata, err := metadataStorage.FetchDataset(request.Dataset, false, false)
	if err != nil {
		return nil, err
	}

	vars := make([]string, len(request.Features)-1)
	target := ""
	c := 0
	for _, v := range request.Features {
		if v.FeatureType == model.FeatureTypeTarget {
			target = v.FeatureName
		} else {
			vars[c] = v.FeatureName
			c = c + 1
		}
	}

	return &api.ExportedModel{
		FilePath:    uri,
		DatasetID:   request.Dataset,
		DatasetName: metadata.Name,
		Variables:   vars,
		Target:      target,
	}, nil
}

// SaveSolution saves a solution to disk via TA2TA3 API.
func SaveSolution(solutionID string) (string, error) {
	return client.SaveSolution(context.Background(), solutionID)
}
