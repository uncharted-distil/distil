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
	"sync"

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

	// get the explainable produce calls which we want to expose
	explainOutputs := s.explainableOutputs(desc)
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
	dataStorage api.DataStorage, initialSearchID string, initialSearchSolutionID string, dataset string, storageName string,
	searchRequest *pipeline.SearchSolutionsRequest, datasetURI string, datasetURITrain string, datasetURITest string, variables []*model.Variable) (string, error) {
	// get solution description
	desc, err := describeSolution(client, initialSearchSolutionID)
	if err != nil {
		return "", err
	}

	// Need to create a new solution that has the explain output. This is the solution
	// that will be used throughout distil except for the export (which will use the original solution).
	// The client API will also reference things by the initial IDs.

	// get the pipeline description
	keywords := make([]string, 0)
	if searchRequest.Problem != nil && searchRequest.Problem.Problem != nil {
		keywords = searchRequest.Problem.Problem.TaskKeywords
	}

	explainDesc, _, err := s.createExplainPipeline(desc, keywords)
	if err != nil {
		return "", err
	}

	// Use the updated explain pipeline if it exists, otherwise use the baseline pipeline
	if explainDesc != nil {
		searchRequest.Template = explainDesc
	} else {
		searchRequest.Template = desc.GetPipeline()
	}

	searchID, err := client.StartSearch(context.Background(), searchRequest)
	if err != nil {
		return "", err
	}
	wg := &sync.WaitGroup{}
	var fittedSolutionID string

	errSearch := client.SearchSolutions(context.Background(), searchID, func(solution *pipeline.GetSearchSolutionsResultsResponse) {
		wg.Add(1)
		defer wg.Done() // make sure wg is flagged on any return

		solutionID := solution.SolutionId

		// persist the solution info
		s.persistSolutionStatus(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, SolutionFittingStatus)

		errNested := solutionStorage.UpdateSolution(initialSearchSolutionID, solutionID)
		if errNested != nil {
			err = errNested
			return
		}

		// fit solution
		fitRequest := createFitSolutionRequest(datasetURITrain, solutionID)
		fitResults, errNested := client.GenerateSolutionFit(context.Background(), fitRequest)
		if errNested != nil {
			err = errNested
			return
		}

		// find the completed result and get the fitted solution ID out
		for _, result := range fitResults {
			if result.GetFittedSolutionId() != "" {
				fittedSolutionID = result.GetFittedSolutionId()
				break
			}
		}
		if fittedSolutionID == "" {
			err = errors.Errorf("no fitted solution ID for solution `%s` ('%s')", solutionID, initialSearchSolutionID)
			return
		}

		s.persistSolutionStatus(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, SolutionScoringStatus)

		// score solution
		solutionScoreResponses, errNested := client.GenerateSolutionScores(context.Background(), solutionID, datasetURITest, s.Metrics)
		if errNested != nil {
			err = errNested
			return
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
					errNested := solutionStorage.PersistSolutionScore(initialSearchSolutionID, metric, score.Value.GetRaw().GetDouble())
					if errNested != nil {
						err = errNested
						return
					}
				}
			}
		}

		// persist solution running status
		s.persistSolutionStatus(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, SolutionProducingStatus)

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
		produceRequestID, predictionResponses, errNested := client.GeneratePredictions(context.Background(), produceSolutionRequest)
		if errNested != nil {
			err = errNested
			return
		}

		for _, response := range predictionResponses {

			if response.Progress.State != pipeline.ProgressState_COMPLETED {
				// only persist completed responses
				continue
			}

			// Generate a path for each output key that has been exposed
			outputKeyURIs := map[string]string{}
			for _, exposedOutputKey := range outputKeys {
				outputURI, errNested := getFileFromOutput(response, exposedOutputKey)
				if errNested != nil {
					err = errNested
					return
				}
				outputKeyURIs[exposedOutputKey] = outputURI
			}

			// get the result UUID. NOTE: Doing sha1 for now.
			resultID := ""
			resultURI, ok := outputKeyURIs[defaultExposedOutputKey]
			if ok {
				// reformat result to have one row per d3m index since confidences
				// can produce one row / class
				resultURI, errNested = reformatResult(resultURI)
				if errNested != nil {
					err = errNested
					return
				}

				resultID, errNested = util.Hash(resultURI)
				if errNested != nil {
					err = errNested
					return
				}
			}

			// persist results
			log.Infof("persisting results in URI '%s'", resultURI)
			s.persistSolutionResults(statusChan, client, solutionStorage, dataStorage,
				initialSearchID, dataset, storageName, solutionID, initialSearchSolutionID, fittedSolutionID,
				produceRequestID, resultID, resultURI, nil, nil)
		}
	})
	if errSearch != nil {
		return "", errSearch
	}
	if err != nil {
		return "", err
	}

	wg.Wait()

	return fittedSolutionID, nil
}
