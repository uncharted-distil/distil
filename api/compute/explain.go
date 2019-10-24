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

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"

	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
)

var (
	explainablePrimitives = map[string]bool{"e0ad06ce-b484-46b0-a478-c567e1ea7e02": true}
)

func (s *SolutionRequest) explainOutput(client *compute.Client, solutionID string, resultURI string,
	searchRequest *pipeline.SearchSolutionsRequest, datasetURI string, variables []*model.Variable) (*api.SolutionFeatureWeights, error) {
	// get the pipeline description
	desc, err := client.GetSolutionDescription(context.Background(), solutionID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get solution description")
	}

	// cycle through the description to determine if any primitive can be explained
	canExplain, pipExplain := s.explainablePipeline(desc)
	if !canExplain {
		return nil, nil
	}

	// send the fully specified pipeline to TA2 (updated produce function call)
	outputURI, err := SubmitPipeline(client, []string{datasetURI}, searchRequest, pipExplain)
	if err != nil {
		return nil, errors.Wrap(err, "unable to run the fully specified pipeline")
	}

	// parse the output for the explanations
	parsed, err := s.parseSolutionFeatureWeight(resultURI, outputURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse feature weight output")
	}

	return parsed, nil
}

func (s *SolutionRequest) parseSolutionFeatureWeight(resultURI string, outputURI string) (*api.SolutionFeatureWeights, error) {
	// all results on one row, with header row having feature names
	res, err := util.ReadCSVFile(outputURI, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read feature weight output")
	}

	return &api.SolutionFeatureWeights{
		ResultURI: resultURI,
		Weights:   res,
	}, nil
}

func (s *SolutionRequest) explainablePipeline(solutionDesc *pipeline.DescribeSolutionResponse) (bool, *pipeline.PipelineDescription) {
	pipelineDesc := solutionDesc.Pipeline
	explainStep := -1
	for si, ps := range pipelineDesc.Steps {
		// get the step outputs
		primitive := ps.GetPrimitive()
		if primitive != nil {
			if s.isExplainablePrimitive(primitive.Primitive.Id) {
				primitive.Outputs[0].Id = "produce_shap_values"
				pipelineDesc.Outputs[0].Data = fmt.Sprintf("steps.%d.produce_shap_values", si)
				explainStep = si
				break
			}
		}
	}

	if explainStep < 0 {
		return false, nil
	}
	pipelineDesc.Steps = pipelineDesc.Steps[0 : explainStep+1]

	mappingStep, _ := description.NewConstructPredictionStep(
		map[string]description.DataRef{"inputs": &description.StepDataRef{StepNum: len(pipelineDesc.Steps), Output: "produce"}},
		[]string{"produce"},
		&description.PipelineDataRef{InputNum: 0},
	).BuildDescriptionStep()
	pipelineDesc.Steps = append(pipelineDesc.Steps, mappingStep)

	return true, pipelineDesc
}

func (s *SolutionRequest) isExplainablePrimitive(primitive string) bool {
	return explainablePrimitives[primitive]
}
