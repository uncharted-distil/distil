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

package compute

import (
	"context"
	"path"
	"strings"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
)

type searchResult struct {
	fittedSolutionID string
	resultID         string
	resultURI        string
}

type pipelineSearchContext struct {
	searchID          string
	dataset           string
	storageName       string
	sourceDatasetURI  string
	trainDatasetURI   string
	testDatasetURI    string
	produceDatasetURI string
	variables         []*model.Variable
	targetCol         int
	groupingCol       int
	sample            bool
}

func (s *SolutionRequest) dispatchSolutionExplainPipeline(client *compute.Client, solutionStorage api.SolutionStorage,
	dataStorage api.DataStorage, searchSolutionID string, searchContext pipelineSearchContext, searchResult *searchResult) error {

	// get solution description
	desc, err := describeSolution(client, searchSolutionID)
	if err != nil {
		return err
	}

	_, explainOutputs := s.createExplainPipeline(desc)

	// if nothing to explain, then exit
	if len(explainOutputs) == 0 {
		return nil
	}

	// create a subset of the test dataset for the explain call if sampling
	explainDatasetURI := searchContext.produceDatasetURI
	if searchContext.sample {
		outputFolder := path.Dir(path.Dir(strings.TrimPrefix(searchContext.produceDatasetURI, "file://")))
		maxRows := getExplainDatasetMaxRows(searchContext.variables)
		explainDatasetURI, err = SampleDataset(searchContext.produceDatasetURI, outputFolder, maxRows, true, searchContext.targetCol, searchContext.groupingCol)
		if err != nil {
			return err
		}
	}

	exposedOutputs := []string{}
	for _, eo := range explainOutputs {
		exposedOutputs = append(exposedOutputs, eo.output)
	}

	exposeType := []string{}
	if s.useParquet {
		exposeType = append(exposeType, compute.ParquetURIValueType)
	}
	produceSolutionRequest := createProduceSolutionRequest(explainDatasetURI, searchResult.fittedSolutionID, exposedOutputs, exposeType)

	// generate predictions
	_, predictionResponses, err := client.GeneratePredictions(context.Background(), produceSolutionRequest)
	if err != nil {
		return err
	}

	for _, response := range predictionResponses {

		if response.Progress.State != pipeline.ProgressState_COMPLETED {
			// only persist completed responses
			continue
		}
		// Generate a path for each output key that has been exposed
		outputKeyURIs := map[string]string{}
		for _, exposedOutput := range explainOutputs {
			outputURI, err := getFileFromOutput(response, exposedOutput.output)
			if err != nil {
				return err
			}
			outputKeyURIs[exposedOutput.key] = outputURI
		}

		// explain features per-record if the explanation is available
		produceOutputs := map[string]*api.SolutionExplainResult{}
		explainedResults := make(map[string]*api.SolutionExplainResult)
		for _, explain := range explainOutputs {
			if explain.typ == ExplainableTypeStep || explain.typ == ExplainableTypeConfidence {
				explainURI := outputKeyURIs[explain.key]
				log.Infof("explaining feature output from URI '%s'", explainURI)
				explainDatasetURI = compute.BuildSchemaFileURI(explainDatasetURI)
				parsedExplainResult, err := ExplainFeatureOutput(explainDatasetURI, explainURI)
				parsedExplainResult.ResultURI = searchResult.resultURI
				if err != nil {
					log.Warnf("failed to fetch output explanation - %v", err)
					continue
				}
				parsedExplainResult.ParsingFunction = explain.parsingFunction
				explainedResults[explain.typ] = parsedExplainResult
			}
			produceOutputs[explain.typ] = &api.SolutionExplainResult{
				ResultURI: outputKeyURIs[explain.key],
			}
		}

		featureWeights := explainedResults[ExplainableTypeStep]
		if featureWeights != nil {
			log.Infof("persisting feature weights")
			err = dataStorage.PersistSolutionFeatureWeight(searchContext.dataset, searchContext.storageName, featureWeights.ResultURI, featureWeights.Values)
			if err != nil {
				return err
			}
		}

		// explain the features at the model level if the explanation is available
		explainSolutionOutput := explainOutputs[ExplainableTypeSolution]
		if explainSolutionOutput != nil {
			explainSolutionURI := outputKeyURIs[explainSolutionOutput.key]
			log.Infof("explaining solution output from URI '%s'", explainSolutionURI)
			solutionWeights, err := s.explainSolutionOutput(explainSolutionURI, searchSolutionID, searchContext.variables)
			if err != nil {
				log.Warnf("failed to fetch output explanantion - %v", err)
			}
			for _, fw := range solutionWeights {
				err = solutionStorage.PersistSolutionWeight(fw.SolutionID, fw.FeatureName, fw.FeatureIndex, fw.Weight)
				if err != nil {
					return err
				}
			}
		}

		// store the explain URIs
		err = solutionStorage.PersistSolutionExplainedOutput(searchResult.resultID, produceOutputs)
		if err != nil {
			return err
		}

		// update results to store additional confidence / explain information
		if produceOutputs[ExplainableTypeConfidence] != nil {
			err = dataStorage.PersistExplainedResult(searchContext.dataset, searchContext.storageName, searchResult.resultURI, explainedResults[ExplainableTypeConfidence])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SolutionRequest) dispatchSolutionSearchPipeline(statusChan chan SolutionStatus, client *compute.Client,
	solutionStorage api.SolutionStorage, dataStorage api.DataStorage, searchSolutionID string, searchContext pipelineSearchContext) (*searchResult, error) {
	var fittedSolutionID string
	var resultURI string
	var resultID string

	// persist the solution info
	s.persistSolutionStatus(statusChan, solutionStorage, searchContext.searchID, searchSolutionID, compute.SolutionFittingStatus)

	// fit solution
	fitRequest := createFitSolutionRequest(searchContext.trainDatasetURI, searchSolutionID)

	// create a context that will let us cancel the streaming request from the client side.  We have to do this at the transport
	// level rather than the ta3ta2 API level since the API only allows for stopping the search portion of the process
	cancelContext, cancelFunc := context.WithCancel(context.Background())
	s.CancelFuncs[searchSolutionID] = cancelFunc

	fitResults, err := client.GenerateSolutionFit(cancelContext, fitRequest)
	if err != nil {
		return nil, err
	}

	// find the completed result and get the fitted solution ID out
	for _, result := range fitResults {
		if result.GetFittedSolutionId() != "" {
			fittedSolutionID = result.GetFittedSolutionId()
			break
		}
	}
	if fittedSolutionID == "" {
		return nil, errors.Errorf("no fitted solution ID for solution `%s`", searchSolutionID)
	}

	s.persistSolutionStatus(statusChan, solutionStorage, searchContext.searchID, searchSolutionID, compute.SolutionScoringStatus)

	// score solution
	solutionScoreResponses, err := client.GenerateSolutionScores(cancelContext, searchSolutionID, searchContext.testDatasetURI, s.Metrics, s.PosLabel)
	if err != nil {
		return nil, err
	}

	// persist the scores
	for _, response := range solutionScoreResponses {
		// only persist scores from COMPLETED responses
		if response.Progress.State == pipeline.ProgressState_COMPLETED {
			for _, score := range response.Scores {
				metric := ""
				if score.GetMetric() == nil {
					metric = compute.ConvertMetricsFromTA3ToTA2(s.Metrics, s.PosLabel)[0].GetMetric()
				} else {
					metric = score.Metric.Metric
				}
				err := solutionStorage.PersistSolutionScore(searchSolutionID, metric, score.Value.GetRaw().GetDouble())
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// persist solution running status
	s.persistSolutionStatus(statusChan, solutionStorage, searchContext.searchID, searchSolutionID, compute.SolutionProducingStatus)

	// generate output keys, adding one extra for explanation output if we expect it to exist
	outputKeys := []string{compute.DefaultExposedOutputKey}
	exposeType := []string{}
	if s.useParquet {
		exposeType = append(exposeType, compute.ParquetURIValueType)
	}
	produceSolutionRequest := createProduceSolutionRequest(searchContext.produceDatasetURI, fittedSolutionID, outputKeys, exposeType)

	// generate predictions
	produceRequestID, predictionResponses, err := client.GeneratePredictions(cancelContext, produceSolutionRequest)
	if err != nil {
		return nil, err
	}

	for _, response := range predictionResponses {

		if response.Progress.State != pipeline.ProgressState_COMPLETED {
			// only persist completed responses
			continue
		}

		// Generate a path for each output key that has been exposed
		outputKeyURIs := map[string]string{}
		for _, exposedOutputKey := range outputKeys {
			outputURI, err := getFileFromOutput(response, exposedOutputKey)
			if err != nil {
				return nil, err
			}
			outputKeyURIs[exposedOutputKey] = outputURI
		}

		// get the result UUID. NOTE: Doing sha1 for now.

		var ok bool
		resultURI, ok = outputKeyURIs[compute.DefaultExposedOutputKey]
		if ok {
			// reformat result to have one row per d3m index since confidences
			// can produce one row / class
			resultURI, err = reformatResult(resultURI)
			if err != nil {
				return nil, err
			}

			resultID, err = util.Hash(resultURI)
			if err != nil {
				return nil, err
			}
		}

		// persist results
		log.Infof("persisting results in URI '%s'", resultURI)
		err = s.persistSolutionResults(statusChan, client, solutionStorage, dataStorage, searchContext.searchID,
			searchContext.dataset, searchContext.storageName, searchSolutionID, fittedSolutionID, produceRequestID, resultID, resultURI)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	return &searchResult{
		resultID:         resultID,
		resultURI:        resultURI,
		fittedSolutionID: fittedSolutionID,
	}, nil
}
