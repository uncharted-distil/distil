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

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
)

func (s *SolutionRequest) dispatchSolutionExplainPipeline(statusChan chan SolutionStatus, client *compute.Client, solutionStorage api.SolutionStorage,
	dataStorage api.DataStorage, fitSolutionID string, searchID string, searchSolutionID string, dataset string, storageName string,
	searchRequest *pipeline.SearchSolutionsRequest, datasetURI string, datasetURITrain string, datasetURITest string, variables []*model.Variable) {

	// get solution description
	desc, err := describeSolution(client, searchSolutionID)
	if err != nil {
		s.persistSolutionError(statusChan, solutionStorage, searchID, searchSolutionID, err)
		return
	}

	keywords := make([]string, 0)
	if searchRequest.Problem != nil && searchRequest.Problem.Problem != nil {
		keywords = searchRequest.Problem.Problem.TaskKeywords
	}

	_, explainOutputs := s.createExplainPipeline(desc, keywords)

	exposedOutputs := []string{}
	for _, eo := range explainOutputs {
		exposedOutputs = append(exposedOutputs, eo.output)
	}

	exposeType := []string{}
	if s.useParquet {
		exposeType = append(exposeType, compute.ParquetURIValueType)
	}
	produceSolutionRequest := createProduceSolutionRequest(datasetURI, fitSolutionID, exposedOutputs, exposeType)

	// generate predictions
	_, predictionResponses, err := client.GeneratePredictions(context.Background(), produceSolutionRequest)
	if err != nil {
		s.persistSolutionError(statusChan, solutionStorage, searchID, searchSolutionID, err)
		return
	}

	for _, response := range predictionResponses {

		if response.Progress.State != pipeline.ProgressState_COMPLETED {
			// only persist completed responses
			continue
		}

		log.Infof("processing response %v", response)
	}
}

func (s *SolutionRequest) dispatchSolutionSearchPipeline(statusChan chan SolutionStatus, client *compute.Client, solutionStorage api.SolutionStorage,
	dataStorage api.DataStorage, searchID string, solutionID string, dataset string, storageName string,
	searchRequest *pipeline.SearchSolutionsRequest, datasetURI string, datasetURITrain string, datasetURITest string, variables []*model.Variable) (string, error) {
	// get the pipeline description
	var fittedSolutionID string

	// persist the solution info
	s.persistSolutionStatus(statusChan, solutionStorage, searchID, solutionID, SolutionFittingStatus)

	// fit solution
	fitRequest := createFitSolutionRequest(datasetURITrain, solutionID)
	fitResults, err := client.GenerateSolutionFit(context.Background(), fitRequest)
	if err != nil {
		return "", err
	}

	// find the completed result and get the fitted solution ID out
	for _, result := range fitResults {
		if result.GetFittedSolutionId() != "" {
			fittedSolutionID = result.GetFittedSolutionId()
			break
		}
	}
	if fittedSolutionID == "" {
		return "", errors.Errorf("no fitted solution ID for solution `%s` ('%s')", solutionID, solutionID)
	}

	s.persistSolutionStatus(statusChan, solutionStorage, searchID, solutionID, SolutionScoringStatus)

	// score solution
	solutionScoreResponses, err := client.GenerateSolutionScores(context.Background(), solutionID, datasetURITest, s.Metrics)
	if err != nil {
		return "", err
	}

	// persist the scores
	for _, response := range solutionScoreResponses {
		// only persist scores from COMPLETED responses
		if response.Progress.State == pipeline.ProgressState_COMPLETED {
			for _, score := range response.Scores {
				metric := ""
				if score.GetMetric() == nil {
					metric = compute.ConvertMetricsFromTA3ToTA2(s.Metrics)[0].GetMetric()
				} else {
					metric = score.Metric.Metric
				}
				err := solutionStorage.PersistSolutionScore(solutionID, metric, score.Value.GetRaw().GetDouble())
				if err != nil {
					return "", err
				}
			}
		}
	}

	// persist solution running status
	s.persistSolutionStatus(statusChan, solutionStorage, searchID, solutionID, SolutionProducingStatus)

	// generate output keys, adding one extra for explanation output if we expect it to exist
	outputKeys := []string{defaultExposedOutputKey}

	// generate predictions -  for timeseries we want to use the entire source dataset, for anything else
	// we only want the test data predictions.
	produceDatasetURI := datasetURITest
	for _, task := range s.Task {
		if task == compute.ForecastingTask {
			produceDatasetURI = datasetURI
			break
		}
	}
	exposeType := []string{}
	if s.useParquet {
		exposeType = append(exposeType, compute.ParquetURIValueType)
	}
	produceSolutionRequest := createProduceSolutionRequest(produceDatasetURI, fittedSolutionID, outputKeys, exposeType)

	// generate predictions
	produceRequestID, predictionResponses, err := client.GeneratePredictions(context.Background(), produceSolutionRequest)
	if err != nil {
		return "", err
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
				return "", err
			}
			outputKeyURIs[exposedOutputKey] = outputURI
		}

		// get the result UUID. NOTE: Doing sha1 for now.
		resultID := ""
		resultURI, ok := outputKeyURIs[defaultExposedOutputKey]
		if ok {
			// reformat result to have one row per d3m index since confidences
			// can produce one row / class
			resultURI, err = reformatResult(resultURI)
			if err != nil {
				return "", err
			}

			resultID, err = util.Hash(resultURI)
			if err != nil {
				return "", err
			}
		}

		// persist results
		log.Infof("persisting results in URI '%s'", resultURI)
		s.persistSolutionResults(statusChan, client, solutionStorage, dataStorage,
			searchID, dataset, storageName, solutionID, solutionID, fittedSolutionID,
			produceRequestID, resultID, resultURI, nil, nil)
	}
	if err != nil {
		return "", err
	}

	return fittedSolutionID, nil
}
