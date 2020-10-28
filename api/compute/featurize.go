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

package compute

import (
	"context"
	"fmt"

	uuid "github.com/gofrs/uuid"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"

	api "github.com/uncharted-distil/distil/api/model"
)

var (
	learningPrimitives = map[string]bool{"e0ad06ce-b484-46b0-a478-c567e1ea7e02": true}
)

// FeaturizeDataset creates feature outputs that can then be used directly when
// modelling instead of needing to run the complete pipeline.
func FeaturizeDataset(dataset string, target string) ([]string, error) {
	// build the normal solution search

	// start a solution searchID
	//requestID, err := client.StartSearch(context.Background(), searchRequest)
	//if err != nil {
	//		return nil, err
	//}

	return nil, nil
}

func (s *SolutionRequest) dispatchFeaturizePipeline(client *compute.Client,
	searchID string, datasetID string, finished chan error) {

	// search for solutions, this wont return until the search finishes or it times out
	err := client.SearchSolutions(context.Background(), searchID, func(solution *pipeline.GetSearchSolutionsResultsResponse) {
		// create a new status channel for the solution
		c := newStatusChannel()
		// add the solution to the request
		s.addSolution(c)
		// dispatch it
		dispatchFeaturizeSolution(client, searchID, datasetID)
		// once done, mark as complete
		s.completeSolution()
		close(c)
	})
	if err != nil {
		finished <- err
		return
	}

	// wait until all are complete and the search has finished / timed out
	s.waitOnSolutions()
	close(s.requestChannel)

	// copy the output to the featurized dataset location and match d3m dataset structure
	//featureDatasetID := fmt.Sprintf("%s-%s", datasetID, outputID)
	//outputPath := env.ResolvePath(metadata.Augmented, featureDatasetID)

	// create the metadata file

	// end search
	finished <- nil
}

func dispatchFeaturizeSolution(client *compute.Client, initialSearchSolutionID string, datasetID string) (string, error) {
	// describe the pipeline
	desc, err := describeSolution(client, initialSearchSolutionID)
	if err != nil {
		return "", err
	}

	// modify the pipeline to get the feature output
	desc, err = createFeaturizePipeline(desc)
	if err != nil {
		return "", err
	}

	// submit the featurization pipeline
	searchID, err := client.StartSearch(context.Background(), nil)
	if err != nil {
		return "", err
	}

	err = client.SearchSolutions(context.Background(), searchID, func(solution *pipeline.GetSearchSolutionsResultsResponse) {
	})
	if err != nil {
		return "", err
	}

	// return the dataset URIs

	return "", nil
}

// createPreFeaturizedPipeline creates pipeline prepend to process a featurized dataset.
func (s *SolutionRequest) createPreFeaturizedPipeline(learningDataset string,
	sourceVariables []*model.Variable, featurizedVariables []*model.Variable,
	metaStorage api.MetadataStorage, targetIndex int) (*pipeline.PipelineDescription, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("prefeaturized-%s-%s-%s", s.Dataset, learningDataset, uuid.String())
	desc := fmt.Sprintf("Prefeaturized pipeline capturing user feature selection and type information. Dataset: `%s` ID: `%s`", s.Dataset, uuid.String())

	expandedFilters, err := api.ExpandFilterParams(s.Dataset, s.Filters, true, metaStorage)
	if err != nil {
		return nil, err
	}

	// Ensure we remove multiband image data (replaced by feature vectors) and geo coordinates polygon
	// string.  These are in the pre-featurized learning data but are not needed.
	toRemove := map[string]bool{}
	for _, variable := range sourceVariables {
		switch v := variable.Grouping.(type) {
		case *model.MultiBandImageGrouping:
			toRemove[v.BandCol] = true
			toRemove[v.ImageCol] = true
			toRemove[v.IDCol] = true
		case *model.GeoBoundsGrouping:
			toRemove[v.PolygonCol] = true
		default:
			continue
		}
	}
	selectedVariables := []string{}
	for _, v := range expandedFilters.Variables {
		if _, ok := toRemove[v]; !ok {
			selectedVariables = append(selectedVariables, v)
		}
	}
	expandedFilters.Variables = selectedVariables

	prefeaturizedPipeline, err := description.CreatePreFeaturizedDatasetPipeline(name, desc,
		&description.UserDatasetDescription{
			AllFeatures:      featurizedVariables,
			TargetFeature:    featurizedVariables[targetIndex],
			SelectedFeatures: expandedFilters.Variables,
			Filters:          s.Filters.Filters,
		}, nil)
	if err != nil {
		return nil, err
	}

	return prefeaturizedPipeline, nil
}

func createFeaturizePipeline(desc *pipeline.DescribeSolutionResponse) (*pipeline.DescribeSolutionResponse, error) {
	//TODO: can descriptions be cloned???
	// find the main step learning step of the pipeline
	primitive, index := getLearningStep(desc)
	if primitive == nil {
		return nil, nil
	}

	// use the input to the learning step as pipeline output
	desc.Pipeline.Outputs = []*pipeline.PipelineDescriptionOutput{
		{
			Name: "outputs.0",
			Data: primitive.Arguments["inputs"].GetData().GetData(),
		},
	}

	// remove extra steps
	desc.Pipeline.Steps = desc.Pipeline.Steps[:index]
	return desc, nil
}

func getLearningStep(desc *pipeline.DescribeSolutionResponse) (*pipeline.PrimitivePipelineDescriptionStep, int) {
	for si, ps := range desc.Pipeline.Steps {
		// get the step outputs
		primitive := ps.GetPrimitive()
		if primitive != nil {
			learning := learningPrimitives[primitive.Primitive.Id]
			if learning {
				return primitive, si
			}
		}
	}

	return nil, -1
}
