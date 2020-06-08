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
	explainableTypeSolution   = "solution"
	explainableTypeStep       = "step"
	explainableTypeConfidence = "confidence"
)

var (
	explainableOutputPrimitives = map[string][]*explainableOutput{
		"e0ad06ce-b484-46b0-a478-c567e1ea7e02": {
			{
				primitiveID:     "e0ad06ce-b484-46b0-a478-c567e1ea7e02",
				produceFunction: "produce_shap_values",
				explainableType: explainableTypeStep,
			},
			{
				primitiveID:     "e0ad06ce-b484-46b0-a478-c567e1ea7e02",
				produceFunction: "produce_feature_importances",
				explainableType: explainableTypeSolution,
			},
		},
		"STUBBED OUT": {
			{
				primitiveID:     "76b5a479-c209-4d94-92b5-7eba7a4d4499",
				produceFunction: "produce_confidence_intervals",
				explainableType: explainableTypeConfidence,
				parsingParams:   []interface{}{0: 3, 1: 4},
			},
		},
	}
)

type explainableOutput struct {
	primitiveID     string
	produceFunction string
	explainableType string
	parsingParams   []interface{}
}

type pipelineOutput struct {
	key           string
	typ           string
	parsingParams []interface{}
}

func (s *SolutionRequest) createExplainPipeline(client *compute.Client,
	desc *pipeline.DescribeSolutionResponse, keywords []pipeline.TaskKeyword) (*pipeline.PipelineDescription, map[string]*pipelineOutput, error) {
	// remote sensing is not explainable
	for _, kw := range keywords {
		if kw == pipeline.TaskKeyword_REMOTE_SENSING {
			return nil, nil, nil
		}
	}

	if ok, pipExplain, outputs := s.explainablePipeline(desc); ok {
		return pipExplain, outputs, nil
	}
	return nil, nil, nil
}

// ExplainFeatureOutput parses the explain feature output.
func ExplainFeatureOutput(resultURI string, datasetURITest string, outputURI string) (*api.SolutionExplainResult, error) {
	// An unset outputURI means that there is no explanation output, which is a valid case,
	// so we return nil rather than an error so it can be handled downstream.
	if outputURI == "" {
		return nil, nil
	}

	// get the d3m index lookup
	log.Infof("explaining feature output")
	log.Infof("reading raw dataset found in '%s'", datasetURITest)
	rawData, err := readDatasetData(datasetURITest)
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
				pipelineDesc.Outputs = append(pipelineDesc.Outputs, &pipeline.PipelineDescriptionOutput{
					Name: outputName,
					Data: fmt.Sprintf("steps.%d.%s", si, ef.produceFunction),
				})
				explainable = true

				// output 0 is the produce call
				outputs[ef.explainableType] = &pipelineOutput{
					typ:           ef.explainableType,
					key:           outputName,
					parsingParams: ef.parsingParams,
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

func getPipelineOutputs(solutionDesc *pipeline.DescribeSolutionResponse) map[string]*pipelineOutput {
	outputs := make(map[string]*pipelineOutput)
	for _, o := range solutionDesc.Pipeline.Outputs {
		output := createPipelineOutputFromDescription(o)
		if output != nil {
			outputs[output.typ] = output
		}
	}
	return outputs
}

func createPipelineOutputFromDescription(output *pipeline.PipelineDescriptionOutput) *pipelineOutput {
	// use the produce function name to determine what kind of data is being output
	produceFunction := output.Data
	if strings.Contains(produceFunction, "confidence") {
		return &pipelineOutput{
			typ: explainableTypeConfidence,
			key: output.Name,
		}
	} else if strings.Contains(produceFunction, "feature") {
		return &pipelineOutput{
			typ: explainableTypeSolution,
			key: output.Name,
		}
	} else if strings.Contains(produceFunction, "shap") {
		return &pipelineOutput{
			typ: explainableTypeStep,
			key: output.Name,
		}
	}

	return nil
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
	meta, err := metadata.LoadMetadataFromOriginalSchema(uriRaw, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load original schema file")
	}
	mainDR := meta.GetMainDataResource()

	dataPath := path.Join(path.Dir(uriRaw), mainDR.ResPath)
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
