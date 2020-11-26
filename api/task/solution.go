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
func SaveFittedSolution(fittedSolutionID string, modelName string, modelDescription string, solutionStorage api.SolutionStorage, metadataStorage api.MetadataStorage) (*api.ExportedModel, error) {
	uri, err := client.SaveFittedSolution(context.Background(), fittedSolutionID)
	if err != nil {
		return nil, err
	}

	request, err := solutionStorage.FetchRequestByFittedSolutionID(fittedSolutionID)
	if err != nil {
		return nil, err
	}

	dataset, err := metadataStorage.FetchDataset(request.Dataset, false, false)
	if err != nil {
		return nil, err
	}

	weights, err := solutionStorage.FetchSolutionWeights(fittedSolutionID)
	if err != nil {
		return nil, err
	}

	ranks := make(map[string]float64)
	for _, fw := range weights {
		ranks[fw.FeatureName] = fw.Weight
	}

	if len(ranks) == 0 {
		summaryVariables, err := api.FetchSummaryVariables(dataset.ID, metadataStorage)
		if err != nil {
			return nil, err
		}
		ranks, err = TargetRank(dataset.Folder, request.TargetFeature(), summaryVariables, dataset.Source)
		if err != nil {
			return nil, err
		}
	}

	types := make(map[string]string)

	for _, vt := range dataset.Variables {
		types[vt.Name] = vt.Type
	}

	vars := make([]string, len(request.Features)-1)
	varDetails := make([]*api.SolutionVariable, len(request.Features)-1)
	target := ""
	c := 0
	for _, v := range request.Features {
		if v.FeatureType == model.FeatureTypeTarget {
			target = v.FeatureName
		} else {
			vars[c] = v.FeatureName
			varDetails[c] = &api.SolutionVariable{
				Name: v.FeatureName,
				Rank: ranks[v.FeatureName],
				Type: types[v.FeatureName],
			}
			c = c + 1
		}
	}

	return &api.ExportedModel{
		FilePath:         uri,
		FittedSolutionID: fittedSolutionID,
		DatasetID:        request.Dataset,
		DatasetName:      dataset.Name,
		Variables:        vars,
		VariableDetails:  varDetails,
		Target:           target,
		ModelName:        modelName,
		ModelDescription: modelDescription,
	}, nil
}

// SaveSolution saves a solution to disk via TA2TA3 API.
func SaveSolution(solutionID string) (string, error) {
	return client.SaveSolution(context.Background(), solutionID)
}

// LoadFittedSolution loads a fitted solution via TA2TA3 API.
func LoadFittedSolution(fittedSolutionURI string, solutionStorage api.SolutionStorage, metadataStorage api.MetadataStorage) (string, error) {
	fittedSolutionID, err := client.LoadFittedSolution(context.Background(), fittedSolutionURI)
	if err != nil {
		return "", err
	}
	return fittedSolutionID, nil
}

// LoadSolution loads an unfitted solution via TA2TA3 API.
func LoadSolution(solutionURI string) (string, error) {
	fittedSolutionID, err := client.LoadFittedSolution(context.Background(), solutionURI)
	if err != nil {
		return "", err
	}
	return fittedSolutionID, nil
}
