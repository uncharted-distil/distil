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
	"strconv"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"

	api "github.com/uncharted-distil/distil/api/model"
)

var (
	explainablePrimitives = map[string]bool{"test": true}
)

func (s *SolutionRequest) explainOutput(client *compute.Client, solutionID string, datasetURI string, variables []*model.Variable) ([]*api.SolutionFeatureWeight, error) {
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
	outputURI, err := SubmitPipeline(client, []string{datasetURI}, pipExplain)
	if err != nil {
		return nil, errors.Wrap(err, "unable to run the fully specified pipeline")
	}

	// parse the output for the explanations
	output, err := s.parseSolutionFeatureWeight(solutionID, outputURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse feature weight output")
	}

	// map column index to get the feature name
	varsMapped := make(map[int64]*model.Variable)
	for _, v := range variables {
		varsMapped[int64(v.Index)] = v
	}

	for _, fw := range output {
		fw.FeatureName = varsMapped[fw.FeatureIndex].Name
	}

	return output, nil
}

func (s *SolutionRequest) parseSolutionFeatureWeight(solutionID string, outputURI string) ([]*api.SolutionFeatureWeight, error) {
	// each row represents one feature weight
	res, err := result.ParseResultCSV(outputURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read feature weight output")
	}

	weights := make([]*api.SolutionFeatureWeight, len(res)-1)
	for i, v := range res[1:] {
		colIndex, err := strconv.ParseInt(v[0].(string), 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse feature col index")
		}
		weight, err := strconv.ParseFloat(v[1].(string), 64)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse feature weight")
		}
		weights[i-1] = &api.SolutionFeatureWeight{
			SolutionID:   solutionID,
			FeatureIndex: colIndex,
			Weight:       weight,
		}
	}

	return weights, nil
}

func (s *SolutionRequest) explainablePipeline(solutionDesc *pipeline.DescribeSolutionResponse) (bool, *pipeline.PipelineDescription) {
	pipelineDesc := solutionDesc.Pipeline
	for si, ps := range pipelineDesc.Steps {
		// get the step outputs
		primitive := ps.GetPrimitive()
		if primitive != nil {
			if s.explainablePrimitive(primitive.Primitive.Id) {
				primitive.Outputs[0].Id = "produce_feature_importance"
				pipelineDesc.Outputs[0].Data = fmt.Sprintf("steps.%d.outputs.0", si)

				return true, pipelineDesc
			}
		}
	}

	return false, nil
}

func (s *SolutionRequest) explainablePrimitive(primitive string) bool {
	return explainablePrimitives[primitive]
}
