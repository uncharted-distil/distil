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

	dataset, err := metadataStorage.FetchDataset(request.Dataset, false, false, false)
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

	// use target ranking if we do not have feature weights
	// ignore errors in this part because they are secondary to saving the model
	if len(ranks) == 0 {
		summaryVariables, _ := api.FetchSummaryVariables(dataset.ID, metadataStorage)
		ranks, _ = TargetRank(dataset, request.TargetFeature(), summaryVariables, dataset.Source)
	}

	varMap := make(map[string]*model.Variable)
	for _, vt := range dataset.Variables {
		varMap[vt.Key] = vt
	}
	// parse the supplied variables to ensure they're correct
	numOfCorrectVars := 0
	for _, v := range request.Features { 
		if varMap[v.FeatureName] != nil {
			numOfCorrectVars++
		}
	}
	vars := make([]string, numOfCorrectVars-1)
	varDetails := make([]*api.SolutionVariable, numOfCorrectVars-1)
	target := &api.SolutionVariable{}
	c := 0
	for _, v := range request.Features {
		// we can assume that the feature target is in the varMap
		if v.FeatureType == model.FeatureTypeTarget {
			target = api.SolutionVariableFromModelVariable(varMap[v.FeatureName], float64(-1))
		} else if varMap[v.FeatureName] != nil{
			vars[c] = v.FeatureName
			variable := varMap[v.FeatureName]
			varDetails[c] = api.SolutionVariableFromModelVariable(variable, ranks[v.FeatureName])
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
