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

var (
	explainablePrimitivesSolution = map[string]bool{"e0ad06ce-b484-46b0-a478-c567e1ea7e02": true}
	explainablePrimitivesStep     = map[string]bool{"e0ad06ce-b484-46b0-a478-c567e1ea7e02": true}
)

func (s *SolutionRequest) createExplainPipeline(client *compute.Client, desc *pipeline.DescribeSolutionResponse) (*pipeline.PipelineDescription, error) {
	// cycle through the description to determine if any primitive can be explained
	if ok, pipExplain := s.explainablePipeline(desc); ok {
		return pipExplain, nil
	}
	return nil, nil
}

// ExplainFeatureOutput parses the explain feature output.
func ExplainFeatureOutput(resultURI string, datasetURITest string, outputURI string) (*api.SolutionFeatureWeights, error) {
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

func parseFeatureWeight(resultURI string, outputURI string, d3mIndexLookup map[int]string) (*api.SolutionFeatureWeights, error) {
	// all results on one row, with header row having feature names
	res, err := util.ReadCSVFile(outputURI, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read feature weight output")
	}

	err = setD3MIndex(0, d3mIndexLookup, res)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update d3m index")
	}
	res[0][0] = model.D3MIndexFieldName

	return &api.SolutionFeatureWeights{
		ResultURI: resultURI,
		Weights:   res,
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

func (s *SolutionRequest) explainablePipeline(solutionDesc *pipeline.DescribeSolutionResponse) (bool, *pipeline.PipelineDescription) {
	pipelineDesc := solutionDesc.Pipeline
	explainStep := -1
	explainSolution := -1
	for si, ps := range pipelineDesc.Steps {
		// get the step outputs
		primitive := ps.GetPrimitive()
		if primitive != nil {
			if isExplainablePrimitiveStep(primitive.Primitive.Id) {
				primitive.Outputs = append(primitive.Outputs, &pipeline.StepOutput{
					Id: "produce_shap_values",
				})
				explainStep = si
			}
			if isExplainablePrimitiveSolution(primitive.Primitive.Id) {
				primitive.Outputs = append(primitive.Outputs, &pipeline.StepOutput{
					Id: "produce_feature_importances",
				})
				explainSolution = si
			}
		}
	}

	if explainStep >= 0 {
		pipelineDesc.Outputs = append(pipelineDesc.Outputs, &pipeline.PipelineDescriptionOutput{
			Name: "explain_step",
			Data: fmt.Sprintf("steps.%d.produce_shap_values", explainStep),
		})
	}
	if explainSolution >= 0 {
		pipelineDesc.Outputs = append(pipelineDesc.Outputs, &pipeline.PipelineDescriptionOutput{
			Name: "explain_solution",
			Data: fmt.Sprintf("steps.%d.produce_feature_importances", explainStep),
		})
	}

	return explainSolution >= 0 || explainStep >= 0, pipelineDesc
}

func isExplainablePipeline(solutionDesc *pipeline.DescribeSolutionResponse) (bool, bool) {
	pipelineDesc := solutionDesc.Pipeline
	featureExplainable := false
	solutionExplainable := false
	for _, ps := range pipelineDesc.Steps {
		// get the step outputs
		primitive := ps.GetPrimitive()
		if primitive != nil {
			if isExplainablePrimitiveStep(primitive.Primitive.Id) {
				featureExplainable = true
			}
			if isExplainablePrimitiveSolution(primitive.Primitive.Id) {
				solutionExplainable = true
			}
		}
	}

	return featureExplainable, solutionExplainable
}

func isExplainablePrimitiveStep(primitive string) bool {
	return explainablePrimitivesStep[primitive]
}

func isExplainablePrimitiveSolution(primitive string) bool {
	return explainablePrimitivesSolution[primitive]
}

func mapRowIndex(d3mIndexCol int, data [][]string) map[int]string {
	indexMap := make(map[int]string)
	for i, row := range data {
		indexMap[i] = row[d3mIndexCol]
	}

	return indexMap
}

func setD3MIndex(indexCol int, d3mIndexLookup map[int]string, data [][]string) error {
	for i := 1; i < len(data); i++ {
		index, err := strconv.Atoi(data[i][indexCol])
		if err != nil {
			return err
		}
		data[i][indexCol] = d3mIndexLookup[index]
	}

	return nil
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
	meta, err := metadata.LoadMetadataFromOriginalSchema(uriRaw)
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
