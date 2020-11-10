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
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"
	log "github.com/unchartedsoftware/plog"

	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
)

const (
	// ExplainableTypeSolution represents output that explains the solution as a whole.
	ExplainableTypeSolution = "solution"
	// ExplainableTypeStep represents output that explains a specific row.
	ExplainableTypeStep = "step"
	// ExplainableTypeConfidence represents confidence output.
	ExplainableTypeConfidence = "confidence"
)

var (
	explainableOutputPrimitives = map[string][]*explainableOutput{
		"e0ad06ce-b484-46b0-a478-c567e1ea7e02": {
			{
				primitiveID:     "e0ad06ce-b484-46b0-a478-c567e1ea7e02",
				produceFunction: "produce_shap_values",
				explainableType: ExplainableTypeStep,
			},
			{
				primitiveID:     "e0ad06ce-b484-46b0-a478-c567e1ea7e02",
				produceFunction: "produce_feature_importances",
				explainableType: ExplainableTypeSolution,
			},
		},
		"fe0841b7-6e70-4bc3-a56c-0670a95ebc6a": {
			// This is currently failing for larger datasets for 2020-07 eval
			// {
			// 	primitiveID:     "fe0841b7-6e70-4bc3-a56c-0670a95ebc6a",
			// 	produceFunction: "produce_shap_values",
			// 	explainableType: explainableTypeStep,
			// },
			{
				primitiveID:     "fe0841b7-6e70-4bc3-a56c-0670a95ebc6a",
				produceFunction: "produce_feature_importances",
				explainableType: ExplainableTypeSolution,
			},
		},
		"cdbb80e4-e9de-4caa-a710-16b5d727b959": {
			{
				primitiveID:     "cdbb80e4-e9de-4caa-a710-16b5d727b959",
				produceFunction: "produce_shap_values",
				explainableType: ExplainableTypeStep,
			},
			{
				primitiveID:     "cdbb80e4-e9de-4caa-a710-16b5d727b959",
				produceFunction: "produce_feature_importances",
				explainableType: ExplainableTypeSolution,
			},
		},
		"3410d709-0a13-4187-a1cb-159dd24b584b": {
			{
				primitiveID:     "3410d709-0a13-4187-a1cb-159dd24b584b",
				produceFunction: "produce_confidence_intervals",
				explainableType: ExplainableTypeConfidence,
				parsingFunction: parseConfidencesWrapper([]int{1, 2}),
			},
		},
		"76b5a479-c209-4d94-92b5-7eba7a4d4499": {
			{
				primitiveID:     "76b5a479-c209-4d94-92b5-7eba7a4d4499",
				produceFunction: "produce_confidence_intervals",
				explainableType: ExplainableTypeConfidence,
				parsingFunction: parseConfidencesWrapper([]int{1, 2}),
			},
		},
	}

	unexplainableTask = map[string]bool{
		compute.ConvertProblemTaskToTA2(compute.RemoteSensingTask): true,
		compute.ConvertProblemTaskToTA2(compute.ImageTask):         true,
	}
)

type explainableOutput struct {
	primitiveID     string
	produceFunction string
	explainableType string
	parsingFunction func([]string) (*api.SolutionExplainValues, error)
}

type pipelineOutput struct {
	key             string
	typ             string
	output          string
	parsingFunction func([]string) (*api.SolutionExplainValues, error)
}

func parseConfidencesWrapper(params []int) func([]string) (*api.SolutionExplainValues, error) {
	lowIndex := params[0]
	highIndex := params[1]
	return func(data []string) (*api.SolutionExplainValues, error) {
		result := &api.SolutionExplainValues{}
		if data[lowIndex] != "" {
			low, err := strconv.ParseFloat(data[lowIndex], 64)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to parse low confidence")
			}
			result.LowConfidence = low
		}
		if data[highIndex] != "" {

			high, err := strconv.ParseFloat(data[highIndex], 64)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to parse high confidence")
			}
			result.HighConfidence = high
		}

		return result, nil
	}
}

func (s *SolutionRequest) createExplainPipeline(desc *pipeline.DescribeSolutionResponse,
	keywords []string) (*pipeline.PipelineDescription, map[string]*pipelineOutput) {
	// remote sensing and images are not explainable
	// TODO: we may want to look into folding this filtering functionality into
	// the function that builds the explainable pipeline (explainablePipeline).
	for _, kw := range keywords {
		if unexplainableTask[kw] {
			return nil, nil
		}
	}
	ok, pipExplain, explainOutputs := s.explainablePipeline(desc)
	if !ok {
		return nil, nil
	}

	return pipExplain, explainOutputs
}

// ExplainFeatureOutput parses the explain feature output.
func ExplainFeatureOutput(resultURI string, outputURI string) (*api.SolutionExplainResult, error) {
	// An unset outputURI means that there is no explanation output, which is a valid case,
	// so we return nil rather than an error so it can be handled downstream.
	if outputURI == "" {
		return nil, nil
	}

	// get the d3m index lookup
	log.Infof("explaining feature output")
	log.Infof("reading raw dataset found in '%s'", resultURI)
	rawData, err := readDatasetData(resultURI)
	if err != nil {
		return nil, err
	}
	d3mIndexField := getD3MFieldIndex(rawData[0])
	d3mIndexLookup := mapRowIndex(d3mIndexField, rawData[1:])

	// parse the output for the explanations
	log.Infof("parsing feature weight found in '%s' using results found in '%s'", outputURI, resultURI)
	parsed, err := parseFeatureWeight(resultURI, outputURI, d3mIndexLookup)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse feature weight output")
	}

	log.Infof("done explaining feature output")

	return parsed, nil
}

func (s *SolutionRequest) explainSolutionOutput(resultURI string, outputURI string,
	solutionID string, variables []*model.Variable) ([]*api.SolutionWeight, error) {

	// parse the output for the explanations
	parsed, err := s.parseSolutionWeight(solutionID, outputURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse feature weight output")
	}

	// map column name to get the feature index
	varsMapped := make(map[string]*model.Variable)
	for _, v := range variables {
		varsMapped[v.Name] = v
	}

	output := make([]*api.SolutionWeight, 0)
	for _, fw := range parsed {
		if varsMapped[fw.FeatureName] != nil {
			fw.FeatureIndex = int64(varsMapped[fw.FeatureName].Index)
			output = append(output, fw)
		}
	}

	return output, nil
}

func parseFeatureWeight(resultURI string, outputURI string, d3mIndexLookup map[int]string) (*api.SolutionExplainResult, error) {
	// all results on one row, with header row having feature names
	res, err := util.ReadCSVFile(outputURI, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read feature weight output")
	}

	res = addD3MIndex(d3mIndexLookup, res)
	return &api.SolutionExplainResult{
		ResultURI:     resultURI,
		Values:        res,
		D3MIndexIndex: len(res[0]) - 1,
	}, nil
}

func (s *SolutionRequest) parseSolutionWeight(solutionID string, outputURI string) ([]*api.SolutionWeight, error) {
	// all results on one row, with header row having feature names
	res, err := result.ParseResultCSV(outputURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read solution weight output")
	}

	weights := make([]*api.SolutionWeight, len(res[0]))
	for i := 0; i < len(res[0]); i++ {
		weight, err := strconv.ParseFloat(res[1][i].(string), 64)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse feature weight")
		}
		featureName := res[0][i].(string)
		weights[i] = &api.SolutionWeight{
			SolutionID:  solutionID,
			FeatureName: featureName,
			Weight:      weight,
		}
	}

	return weights, nil
}

func (s *SolutionRequest) explainablePipeline(solutionDesc *pipeline.DescribeSolutionResponse) (bool, *pipeline.PipelineDescription, map[string]*pipelineOutput) {
	pipelineDesc := solutionDesc.Pipeline
	explainable := false
	outputs := make(map[string]*pipelineOutput)
	for si, ps := range pipelineDesc.Steps {
		// get the step outputs
		primitive := ps.GetPrimitive()
		if primitive != nil {
			explainFunctions := explainablePrimitiveFunctions(primitive.Primitive.Id)
			for _, ef := range explainFunctions {
				outputName := fmt.Sprintf("outputs.%d", len(outputs)+1)
				primitive.Outputs = append(primitive.Outputs, &pipeline.StepOutput{
					Id: ef.produceFunction,
				})
				output := fmt.Sprintf("steps.%d.%s", si, ef.produceFunction)
				explainable = true

				// output 0 is the produce call
				outputs[ef.explainableType] = &pipelineOutput{
					typ:             ef.explainableType,
					key:             outputName,
					output:          output,
					parsingFunction: ef.parsingFunction,
				}
			}
		}
	}

	return explainable, pipelineDesc, outputs
}

func isExplainablePipeline(solutionDesc *pipeline.DescribeSolutionResponse) bool {
	pipelineDesc := solutionDesc.Pipeline
	for _, ps := range pipelineDesc.Steps {
		// get the step outputs
		primitive := ps.GetPrimitive()
		if primitive != nil {
			if len(explainablePrimitiveFunctions(primitive.Primitive.Id)) > 0 {
				return true
			}
		}
	}

	return false
}

func explainablePrimitiveFunctions(primitiveID string) []*explainableOutput {
	return explainableOutputPrimitives[primitiveID]
}

func explainablePipelineFunctions(solutionDesc *pipeline.DescribeSolutionResponse) []*explainableOutput {
	explainableCalls := make([]*explainableOutput, 0)
	pipelineDesc := solutionDesc.Pipeline
	for _, ps := range pipelineDesc.Steps {
		// get the step outputs
		primitive := ps.GetPrimitive()
		if primitive != nil {
			ep := explainablePrimitiveFunctions(primitive.Primitive.Id)
			if len(ep) > 0 {
				explainableCalls = append(explainableCalls, ep...)
			}
		}
	}

	return explainableCalls
}

func getPipelineOutputs(solutionDesc *pipeline.DescribeSolutionResponse) (map[string]*pipelineOutput, error) {
	outputs := make(map[string]*pipelineOutput)
	for _, o := range solutionDesc.Pipeline.Outputs {
		output, stepIndex, err := createPipelineOutputFromDescription(o)
		if err != nil {
			return nil, err
		}
		if output != nil {
			outputPrimitive := solutionDesc.Pipeline.Steps[stepIndex].GetPrimitive()
			explainOutputs := explainableOutputPrimitives[outputPrimitive.Primitive.Id]
			for _, eo := range explainOutputs {
				if eo.explainableType == output.typ {
					output.parsingFunction = eo.parsingFunction
				}
			}
			outputs[output.typ] = output
		}
	}
	return outputs, nil
}

func createPipelineOutputFromDescription(output *pipeline.PipelineDescriptionOutput) (*pipelineOutput, int, error) {
	// use the produce function name to determine what kind of data is being output
	var explainedOutput *pipelineOutput
	produceFunction := output.Data
	if strings.Contains(produceFunction, "confidence") {
		explainedOutput = &pipelineOutput{
			typ: ExplainableTypeConfidence,
			key: output.Name,
		}
	} else if strings.Contains(produceFunction, "feature") {
		explainedOutput = &pipelineOutput{
			typ: ExplainableTypeSolution,
			key: output.Name,
		}
	} else if strings.Contains(produceFunction, "shap") {
		explainedOutput = &pipelineOutput{
			typ: ExplainableTypeStep,
			key: output.Name,
		}
	}

	step := 0
	if explainedOutput != nil {
		stepRaw, err := strconv.Atoi(strings.Split(produceFunction, ".")[1])
		if err != nil {
			return nil, -1, errors.Wrapf(err, "unable to parse output step for explained output")
		}
		step = stepRaw
	}

	return explainedOutput, step, nil
}

func mapRowIndex(d3mIndexCol int, data [][]string) map[int]string {
	indexMap := make(map[int]string)
	for i, row := range data {
		indexMap[i] = row[d3mIndexCol]
	}

	return indexMap
}

func addD3MIndex(d3mIndexLookup map[int]string, data [][]string) [][]string {
	// assume row order matches for the lookup
	clone := make([][]string, len(data))
	clone[0] = append(data[0], model.D3MIndexFieldName)
	for i := 1; i < len(data); i++ {
		clone[i] = append(data[i], d3mIndexLookup[i-1])
	}

	return clone
}

func getD3MFieldIndex(header []string) int {
	for i, f := range header {
		if f == model.D3MIndexFieldName {
			return i
		}
	}

	return -1
}

func readDatasetData(uri string) ([][]string, error) {
	uriRaw := strings.TrimPrefix(uri, "file://")
	dataPath := uriRaw
	if path.Ext(dataPath) == ".json" {
		meta, err := metadata.LoadMetadataFromOriginalSchema(uriRaw, true)
		if err != nil {
			return nil, errors.Wrap(err, "unable to load original schema file")
		}
		mainDR := meta.GetMainDataResource()

		dataPath = path.Join(path.Dir(uriRaw), mainDR.ResPath)
	}

	res, err := util.ReadCSVFile(dataPath, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read raw input data")
	}

	return res, nil
}

func extractOutputKeys(outputs map[string]*pipelineOutput) []string {
	keys := make([]string, 0)
	for _, po := range outputs {
		keys = append(keys, po.key)
	}

	return keys
}
