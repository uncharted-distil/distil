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
	"path/filepath"
	"strings"
	"sync"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/util"
	"github.com/uncharted-distil/distil/api/util/json"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
)

const (
	defaultExposedOutputKey  = "outputs.0"
	explainFeatureOutputkey  = "outputs.1"
	explainSolutionOutputkey = "outputs.2"
	// SolutionPendingStatus represents that the solution request has been acknoledged by not yet sent to the API
	SolutionPendingStatus = "SOLUTION_PENDING"
	// SolutionFittingStatus represents that the solution request has been sent to the API.
	SolutionFittingStatus = "SOLUTION_FITTING"
	// SolutionScoringStatus represents that the solution request has been sent to the API.
	SolutionScoringStatus = "SOLUTION_SCORING"
	// SolutionProducingStatus represents that the solution request has been sent to the API.
	SolutionProducingStatus = "SOLUTION_PRODUCING"
	// SolutionErroredStatus represents that the solution request has terminated with an error.
	SolutionErroredStatus = "SOLUTION_ERRORED"
	// SolutionCompletedStatus represents that the solution request has completed successfully.
	SolutionCompletedStatus = "SOLUTION_COMPLETED"
	// RequestPendingStatus represents that the solution request has been acknoledged by not yet sent to the API
	RequestPendingStatus = "REQUEST_PENDING"
	// RequestRunningStatus represents that the solution request has been sent to the API.
	RequestRunningStatus = "REQUEST_RUNNING"
	// RequestErroredStatus represents that the solution request has terminated with an error.
	RequestErroredStatus = "REQUEST_ERRORED"
	// RequestCompletedStatus represents that the solution request has completed successfully.
	RequestCompletedStatus = "REQUEST_COMPLETED"
)

var (
	// folder for dataset data exchanged with TA2
	datasetDir string
)

// SetDatasetDir sets the output data dir
func SetDatasetDir(dir string) {
	datasetDir = dir
}

func newStatusChannel() chan SolutionStatus {
	// NOTE: WE BUFFER THE CHANNEL TO A SIZE OF 1 HERE SO THAT THE INITIAL
	// PERSIST DOES NOT DEADLOCK
	return make(chan SolutionStatus, 1)
}

// PredictionResult contains the output from a prediction produce call.
type PredictionResult struct {
	ProduceRequestID         string
	FittedSolutionID         string
	ResultURI                string
	ConfidenceURI            string
	SolutionFeatureWeightURI string
	StepFeatureWeightURI     string
}

// SolutionRequest represents a solution search request.
type SolutionRequest struct {
	Dataset              string
	DatasetInput         string
	TargetFeature        *model.Variable
	Task                 []string
	TimestampField       string
	MaxSolutions         int
	MaxTime              int
	ProblemType          string
	Metrics              []string
	Filters              *api.FilterParams
	DatasetAugmentations []*model.DatasetOrigin

	mu               *sync.Mutex
	wg               *sync.WaitGroup
	requestChannel   chan SolutionStatus
	solutionChannels []chan SolutionStatus
	listener         SolutionStatusListener
	finished         chan error
}

// NewSolutionRequest instantiates a new SolutionRequest.
func NewSolutionRequest(variables []*model.Variable, data []byte) (*SolutionRequest, error) {
	req := &SolutionRequest{
		mu:             &sync.Mutex{},
		wg:             &sync.WaitGroup{},
		finished:       make(chan error),
		requestChannel: newStatusChannel(),
	}

	j, err := json.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	var ok bool

	req.Dataset, ok = json.String(j, "dataset")
	if !ok {
		return nil, fmt.Errorf("no `dataset` in solution request")
	}
	req.DatasetInput = req.Dataset

	targetName, ok := json.String(j, "target")
	if !ok {
		return nil, fmt.Errorf("no `target` in solution request")
	}
	for _, v := range variables {
		if v.Name == targetName {
			req.TargetFeature = v
		}
	}

	req.Task, _ = json.StringArray(j, "task")
	req.MaxSolutions = json.IntDefault(j, 5, "maxSolutions")
	req.MaxTime = json.IntDefault(j, 0, "maxTime")
	req.ProblemType = json.StringDefault(j, "", "problemType")
	req.Metrics, _ = json.StringArray(j, "metrics")

	filters, ok := json.Get(j, "filters")
	if ok {
		req.Filters, err = api.ParseFilterParamsFromJSON(filters)
		if err != nil {
			return nil, err
		}
	}

	return req, nil
}

// ExtractDatasetFromRawRequest extracts the dataset name from the raw message.
func ExtractDatasetFromRawRequest(data []byte) (string, error) {
	j, err := json.Unmarshal(data)
	if err != nil {
		return "", err
	}

	var ok bool

	dataset, ok := json.String(j, "dataset")
	if !ok {
		return "", fmt.Errorf("no `dataset` in solution request")
	}

	return dataset, nil
}

// SolutionStatus represents a solution status.
type SolutionStatus struct {
	Progress   string    `json:"progress"`
	RequestID  string    `json:"requestId"`
	SolutionID string    `json:"solutionId"`
	ResultID   string    `json:"resultId"`
	Error      error     `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
}

// SolutionStatusListener executes on a new solution status.
type SolutionStatusListener func(status SolutionStatus)

func (s *SolutionRequest) addSolution(c chan SolutionStatus) {
	s.wg.Add(1)
	s.mu.Lock()
	s.solutionChannels = append(s.solutionChannels, c)
	if s.listener != nil {
		go s.listenOnStatusChannel(c)
	}
	s.mu.Unlock()
}

func (s *SolutionRequest) completeSolution() {
	s.wg.Done()
}

func (s *SolutionRequest) waitOnSolutions() {
	s.wg.Wait()
}

func (s *SolutionRequest) listenOnStatusChannel(statusChannel <-chan SolutionStatus) {
	for status := range statusChannel {
		s.listener(status)
	}
}

// Listen listens ont he solution requests for new solution statuses.
func (s *SolutionRequest) Listen(listener SolutionStatusListener) error {
	s.listener = listener
	s.mu.Lock()
	// listen on main request channel
	go s.listenOnStatusChannel(s.requestChannel)
	// listen on individual solution channels
	for _, c := range s.solutionChannels {
		go s.listenOnStatusChannel(c)
	}
	s.mu.Unlock()
	return <-s.finished
}

func (s *SolutionRequest) createSearchSolutionsRequest(columnIndex int, preprocessing *pipeline.PipelineDescription,
	datasetURI string, userAgent string) (*pipeline.SearchSolutionsRequest, error) {
	return createSearchSolutionsRequest(columnIndex, preprocessing, datasetURI, userAgent, s.TargetFeature, s.Dataset, s.Metrics, s.Task, int64(s.MaxTime))
}

func createSearchSolutionsRequest(columnIndex int, preprocessing *pipeline.PipelineDescription,
	datasetURI string, userAgent string, targetFeature *model.Variable, dataset string, metrics []string, task []string, maxTime int64) (*pipeline.SearchSolutionsRequest, error) {

	return &pipeline.SearchSolutionsRequest{
		Problem: &pipeline.ProblemDescription{
			Problem: &pipeline.Problem{
				TaskKeywords:       compute.ConvertTaskKeywordsFromTA3ToTA2(task),
				PerformanceMetrics: compute.ConvertMetricsFromTA3ToTA2(metrics),
			},
			Inputs: []*pipeline.ProblemInput{
				{
					DatasetId: compute.ConvertDatasetTA3ToTA2(dataset),
					Targets:   compute.ConvertTargetFeaturesTA3ToTA2(targetFeature.Name, columnIndex),
				},
			},
		},

		// Our agent/version info
		UserAgent: userAgent,
		Version:   compute.GetAPIVersion(),

		// Requested max time for solution search - not guaranteed to be honoured
		TimeBoundSearch: float64(maxTime),

		// Requested max time for pipeline run - not guaranteed to be honoured
		TimeBoundRun: float64(maxTime),

		// we accept dataset and csv uris as return types
		AllowedValueTypes: []string{
			compute.CSVURIValueType,
			compute.DatasetURIValueType,
			compute.RawValueType,
		},

		// URI of the input dataset
		Inputs: []*pipeline.Value{
			{
				Value: &pipeline.Value_DatasetUri{
					DatasetUri: datasetURI,
				},
			},
		},

		Template: preprocessing,
	}, nil
}

// createPreprocessingPipeline creates pipeline to enfore user feature selection and typing
func (s *SolutionRequest) createPreprocessingPipeline(featureVariables []*model.Variable, metaStorage api.MetadataStorage) (*pipeline.PipelineDescription, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("preprocessing-%s-%s", s.Dataset, uuid.String())
	desc := fmt.Sprintf("Preprocessing pipeline capturing user feature selection and type information. Dataset: `%s` ID: `%s`", s.Dataset, uuid.String())

	var augments []*description.UserDatasetAugmentation
	if s.DatasetAugmentations != nil {
		augments = make([]*description.UserDatasetAugmentation, len(s.DatasetAugmentations))
		for i, da := range s.DatasetAugmentations {
			augments[i] = &description.UserDatasetAugmentation{
				SearchResult:  da.SearchResult,
				SystemID:      da.Provenance,
				BaseDatasetID: s.Dataset,
			}
		}
	}

	// replace any grouped variables in filter params with the group's
	expandedFilters, err := api.ExpandFilterParams(s.Dataset, s.Filters, metaStorage)
	if err != nil {
		return nil, err
	}

	preprocessingPipeline, err := description.CreateUserDatasetPipeline(name, desc,
		&description.UserDatasetDescription{
			AllFeatures:      featureVariables,
			TargetFeature:    s.TargetFeature,
			SelectedFeatures: expandedFilters.Variables,
			Filters:          s.Filters.Filters,
		}, augments)
	if err != nil {
		return nil, err
	}

	return preprocessingPipeline, nil
}

// GeneratePredictions produces predictions using the specified.
func GeneratePredictions(datasetURI string, explainedSolutionID string,
	fittedSolutionID string, client *compute.Client) (*PredictionResult, error) {
	// check if the solution can be explained
	desc, err := client.GetSolutionDescription(context.Background(), explainedSolutionID)
	if err != nil {
		return nil, err
	}

	outputs := getPipelineOutputs(desc)
	keys := []string{defaultExposedOutputKey}
	keys = append(keys, extractOutputKeys(outputs)...)

	produceRequest := createProduceSolutionRequest(datasetURI, fittedSolutionID, keys)
	produceRequestID, predictionResponses, err := client.GeneratePredictions(context.Background(), produceRequest)
	if err != nil {
		return nil, err
	}

	for _, response := range predictionResponses {

		if response.Progress.State != pipeline.ProgressState_COMPLETED {
			// only persist completed responses
			continue
		}

		resultURI, err := getFileFromOutput(response, defaultExposedOutputKey)
		if err != nil {
			return nil, err
		}
		var explainFeatureURI string
		if outputs[explainableTypeStep] != nil {
			explainFeatureURI, err = getFileFromOutput(response, outputs[explainableTypeStep].key)
			if err != nil {
				return nil, err
			}
		}
		var explainSolutionURI string
		if outputs[explainableTypeSolution] != nil {
			explainSolutionURI, err = getFileFromOutput(response, outputs[explainableTypeSolution].key)
			if err != nil {
				return nil, err
			}
		}
		var confidenceURI string
		if outputs[explainableTypeConfidence] != nil {
			confidenceURI, err = getFileFromOutput(response, outputs[explainableTypeConfidence].key)
			if err != nil {
				return nil, err
			}
		}

		return &PredictionResult{
			ProduceRequestID:         produceRequestID,
			FittedSolutionID:         fittedSolutionID,
			ResultURI:                resultURI,
			ConfidenceURI:            confidenceURI,
			StepFeatureWeightURI:     explainFeatureURI,
			SolutionFeatureWeightURI: explainSolutionURI,
		}, nil
	}

	return nil, errors.Errorf("no results retrieved")
}

func createProduceSolutionRequest(datasetURI string, fittedSolutionID string, outputs []string) *pipeline.ProduceSolutionRequest {
	return &pipeline.ProduceSolutionRequest{
		FittedSolutionId: fittedSolutionID,
		Inputs: []*pipeline.Value{
			{
				Value: &pipeline.Value_DatasetUri{
					DatasetUri: compute.BuildSchemaFileURI(datasetURI),
				},
			},
		},
		ExposeOutputs: outputs,
		ExposeValueTypes: []string{
			compute.CSVURIValueType,
		},
	}
}

func (s *SolutionRequest) persistSolutionError(statusChan chan SolutionStatus, solutionStorage api.SolutionStorage, searchID string, solutionID string, err error) {
	// persist the updated state
	// NOTE: ignoring error
	solutionStorage.PersistSolutionState(solutionID, SolutionErroredStatus, time.Now())

	// notify of error
	statusChan <- SolutionStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		Progress:   SolutionErroredStatus,
		Error:      err,
		Timestamp:  time.Now(),
	}

	log.Errorf("solution '%s' errored: %v", solutionID, err)
}

func (s *SolutionRequest) persistSolution(statusChan chan SolutionStatus, solutionStorage api.SolutionStorage, searchID string, solutionID string, explainedSolutionID string) {
	err := solutionStorage.PersistSolution(searchID, solutionID, explainedSolutionID, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, searchID, solutionID, err)
		return
	}
}

func (s *SolutionRequest) persistSolutionStatus(statusChan chan SolutionStatus, solutionStorage api.SolutionStorage, searchID string, solutionID string, status string) {
	// persist the updated state
	err := solutionStorage.PersistSolutionState(solutionID, status, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, searchID, solutionID, err)
		return
	}

	// notify of update
	statusChan <- SolutionStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		Progress:   status,
		Timestamp:  time.Now(),
	}
}

func (s *SolutionRequest) persistRequestError(statusChan chan SolutionStatus, solutionStorage api.SolutionStorage, searchID string, dataset string, err error) {
	// persist the updated state
	// NOTE: ignoring error
	solutionStorage.PersistRequest(searchID, dataset, RequestErroredStatus, time.Now())

	// notify of error
	statusChan <- SolutionStatus{
		RequestID: searchID,
		Progress:  RequestErroredStatus,
		Error:     err,
		Timestamp: time.Now(),
	}
}

func (s *SolutionRequest) persistRequestStatus(statusChan chan SolutionStatus, solutionStorage api.SolutionStorage, searchID string, dataset string, status string) error {
	// persist the updated state
	err := solutionStorage.PersistRequest(searchID, dataset, status, time.Now())
	if err != nil {
		// notify of error
		s.persistRequestError(statusChan, solutionStorage, searchID, dataset, err)
		return err
	}

	// notify of update
	statusChan <- SolutionStatus{
		RequestID: searchID,
		Progress:  status,
		Timestamp: time.Now(),
	}
	return nil
}

func (s *SolutionRequest) persistSolutionResults(statusChan chan SolutionStatus, client *compute.Client,
	solutionStorage api.SolutionStorage, dataStorage api.DataStorage, searchID string, initialSearchID string, dataset string,
	explainedSolutionID string, initialSearchSolutionID string, fittedSolutionID string, produceRequestID string, resultID string,
	resultURI string, confidenceValues *api.SolutionExplainResult) {
	// persist the completed state
	err := solutionStorage.PersistSolutionState(initialSearchSolutionID, SolutionCompletedStatus, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}
	// persist result metadata
	err = solutionStorage.PersistSolutionResult(initialSearchSolutionID, fittedSolutionID, produceRequestID, "test", resultID, resultURI, SolutionCompletedStatus, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}
	// persist results
	err = dataStorage.PersistResult(dataset, model.NormalizeDatasetID(dataset), resultURI, s.TargetFeature.Name, confidenceValues)
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}

	// notify client of update
	statusChan <- SolutionStatus{
		RequestID:  initialSearchID,
		SolutionID: initialSearchSolutionID,
		ResultID:   resultID,
		Progress:   SolutionCompletedStatus,
		Timestamp:  time.Now(),
	}
}

func (s *SolutionRequest) dispatchSolution(statusChan chan SolutionStatus, client *compute.Client, solutionStorage api.SolutionStorage,
	dataStorage api.DataStorage, initialSearchID string, initialSearchSolutionID string, dataset string, searchRequest *pipeline.SearchSolutionsRequest,
	datasetURI string, datasetURITrain string, datasetURITest string, variables []*model.Variable) {

	// need to wait until a valid description is returned before proceeding
	var desc *pipeline.DescribeSolutionResponse
	var err error
	for wait := true; wait; {
		desc, err = client.GetSolutionDescription(context.Background(), initialSearchSolutionID)
		if err != nil {
			s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
			return
		}
		wait = desc == nil || desc.Pipeline == nil
		if wait {
			time.Sleep(10 * time.Second)
		}
	}

	// Need to create a new solution that has the explain output. This is the solution
	// that will be used throughout distil except for the export (which will use the original solution).
	// The client API will also reference things by the initial IDs.

	// get the pipeline description
	keywords := make([]string, 0)
	if searchRequest.Problem != nil && searchRequest.Problem.Problem != nil {
		keywords = searchRequest.Problem.Problem.TaskKeywords
	}

	explainDesc, outputKeysExplain, err := s.createExplainPipeline(client, desc, keywords)
	if err != nil {
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}

	// Use the updated explain pipeline if it exists, otherwise use the baseline pipeline
	if explainDesc != nil {
		searchRequest.Template = explainDesc
	} else {
		searchRequest.Template = desc.GetPipeline()
	}

	searchID, err := client.StartSearch(context.Background(), searchRequest)
	if err != nil {
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}
	wg := &sync.WaitGroup{}

	err = client.SearchSolutions(context.Background(), searchID, func(solution *pipeline.GetSearchSolutionsResultsResponse) {
		wg.Add(1)
		defer wg.Done() // make sure wg is flagged on any return

		solutionID := solution.SolutionId

		// persist the solution info
		s.persistSolutionStatus(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, SolutionFittingStatus)

		err = solutionStorage.UpdateSolution(initialSearchSolutionID, solutionID)
		if err != nil {
			s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
			return
		}

		// fit solution
		fitResults, err := client.GenerateSolutionFit(context.Background(), solutionID, []string{datasetURITrain})
		if err != nil {
			s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
			return
		}

		// find the completed result and get the fitted solution ID out
		var fittedSolutionID string
		for _, result := range fitResults {
			if result.GetFittedSolutionId() != "" {
				fittedSolutionID = result.GetFittedSolutionId()
				break
			}
		}
		if fittedSolutionID == "" {
			s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID,
				errors.Errorf("no fitted solution ID for solution `%s` ('%s')", solutionID, initialSearchSolutionID))
		}

		s.persistSolutionStatus(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, SolutionScoringStatus)

		// score solution
		solutionScoreResponses, err := client.GenerateSolutionScores(context.Background(), solutionID, datasetURITest, s.Metrics)
		if err != nil {
			s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
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
					err := solutionStorage.PersistSolutionScore(initialSearchSolutionID, metric, score.Value.GetRaw().GetDouble())
					if err != nil {
						s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
						return
					}
				}
			}
		}

		// persist solution running status
		s.persistSolutionStatus(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, SolutionProducingStatus)

		// generate output keys, adding one extra for explanation output if we expect it to exist
		outputKeys := []string{defaultExposedOutputKey}
		outputKeys = append(outputKeys, extractOutputKeys(outputKeysExplain)...)

		// generate predictions -  for timeseries we want to use the entire source dataset, for anything else
		// we only want the test data predictions.
		produceDatasetURI := datasetURITest
		for _, task := range s.Task {
			if task == compute.ForecastingTask {
				produceDatasetURI = datasetURI
				break
			}
		}
		produceSolutionRequest := createProduceSolutionRequest(produceDatasetURI, fittedSolutionID, outputKeys)

		// generate predictions
		produceRequestID, predictionResponses, err := client.GeneratePredictions(context.Background(), produceSolutionRequest)
		if err != nil {
			s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
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
				outputURI, err := getFileFromOutput(response, exposedOutputKey)
				if err != nil {
					s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
					return
				}
				outputKeyURIs[exposedOutputKey] = outputURI
			}

			// get the result UUID. NOTE: Doing sha1 for now.
			resultID := ""
			resultURI, ok := outputKeyURIs[defaultExposedOutputKey]
			if ok {
				resultID, err = util.Hash(resultURI)
				if err != nil {
					s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
					return
				}
			}

			// explain features per-record if the explanation is available
			explainedResults := make(map[string]*api.SolutionExplainResult)
			for _, explain := range outputKeysExplain {
				if explain.typ == explainableTypeStep || explain.typ == explainableTypeConfidence {
					explainURI := outputKeyURIs[explain.key]
					log.Infof("explaining feature output from URI '%s'", explainURI)
					parsedExplainResult, err := ExplainFeatureOutput(resultURI, datasetURITest, explainURI)
					if err != nil {
						log.Warnf("failed to fetch output explanation - %v", err)
					}
					parsedExplainResult.ParsingParams = explain.parsingParams
					explainedResults[explain.typ] = parsedExplainResult
				}
			}

			featureWeights := explainedResults[explainableTypeStep]
			if featureWeights != nil {
				log.Infof("persisting feature weights")
				err = dataStorage.PersistSolutionFeatureWeight(dataset, model.NormalizeDatasetID(dataset), featureWeights.ResultURI, featureWeights.Values)
				if err != nil {
					s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
					return
				}
			}

			// explain the features at the model level if the explanation is available
			explainSolutionOutput := outputKeysExplain[explainableTypeSolution]
			if explainSolutionOutput != nil {
				explainSolutionURI := outputKeyURIs[explainSolutionOutput.key]
				log.Infof("explaining solution output from URI '%s'", explainSolutionURI)
				solutionWeights, err := s.explainSolutionOutput(resultURI, explainSolutionURI, initialSearchSolutionID, variables)
				if err != nil {
					log.Warnf("failed to fetch output explanantion - %v", err)
				}
				for _, fw := range solutionWeights {
					err = solutionStorage.PersistSolutionWeight(fw.SolutionID, fw.FeatureName, fw.FeatureIndex, fw.Weight)
					if err != nil {
						s.persistSolutionError(statusChan, solutionStorage, searchID, initialSearchSolutionID, err)
						return
					}
				}
			}

			// persist results
			log.Infof("persisting results in URI '%s'", resultURI)
			s.persistSolutionResults(statusChan, client, solutionStorage, dataStorage, searchID,
				initialSearchID, dataset, solutionID, initialSearchSolutionID, fittedSolutionID,
				produceRequestID, resultID, resultURI, explainedResults[explainableTypeConfidence])
		}
	})
	if err != nil {
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}

	wg.Wait()
}

func (s *SolutionRequest) dispatchRequest(client *compute.Client, solutionStorage api.SolutionStorage, dataStorage api.DataStorage,
	searchID string, dataset string, searchRequest *pipeline.SearchSolutionsRequest,
	datasetURI string, datasetURITrain string, datasetURITest string, variables []*model.Variable) {

	// update request status
	err := s.persistRequestStatus(s.requestChannel, solutionStorage, searchID, dataset, RequestRunningStatus)
	if err != nil {
		s.finished <- err
		return
	}

	// search for solutions, this wont return until the search finishes or it times out
	err = client.SearchSolutions(context.Background(), searchID, func(solution *pipeline.GetSearchSolutionsResultsResponse) {
		// create a new status channel for the solution
		c := newStatusChannel()
		// add the solution to the request
		s.addSolution(c)
		// persist the solution
		s.persistSolution(c, solutionStorage, searchID, solution.SolutionId, "")
		s.persistSolutionStatus(c, solutionStorage, searchID, solution.SolutionId, SolutionPendingStatus)
		// dispatch it
		s.dispatchSolution(c, client, solutionStorage, dataStorage, searchID, solution.SolutionId, dataset, searchRequest, datasetURI, datasetURITrain, datasetURITest, variables)
		// once done, mark as complete
		s.completeSolution()
		close(c)
	})

	// wait until all are complete and the search has finished / timed out
	s.waitOnSolutions()

	// update request status
	if err != nil {
		s.persistRequestError(s.requestChannel, solutionStorage, searchID, dataset, err)
	} else {
		s.persistRequestStatus(s.requestChannel, solutionStorage, searchID, dataset, RequestCompletedStatus)
	}
	close(s.requestChannel)

	// end search
	// since predictions can be requested for different datasets on the same
	// fitted solution, can't tell TA2 to end but the channel still needs
	// to be notified that the current process is complete
	//s.finished <- client.EndSearch(context.Background(), searchID)
	s.finished <- nil
}

// PersistAndDispatch persists the solution request and dispatches it.
func (s *SolutionRequest) PersistAndDispatch(client *compute.Client, solutionStorage api.SolutionStorage, metaStorage api.MetadataStorage, dataStorage api.DataStorage) error {

	// NOTE: D3M index field is needed in the persisted data.
	s.Filters.AddVariable(model.D3MIndexFieldName)

	// fetch the dataset variables
	variables, err := metaStorage.FetchVariables(s.Dataset, true, true)
	if err != nil {
		return err
	}

	// remove any generated / grouped features from our var list
	// TODO: imported datasets have d3m index as distil role = "index".
	//       need to figure out if that causes issues!!!
	dataVariables := []*model.Variable{}
	groupingVariableIndex := -1
	for _, variable := range variables {
		if model.IsTA2Field(variable.DistilRole, variable.SelectedRole) {
			dataVariables = append(dataVariables, variable)
		}
		if variable.DistilRole == model.VarDistilRoleGrouping {
			// if this is a group var, find the grouping ID col and use that
			if variable.Grouping != nil {
				groupVariable, err := findVariable(variable.Grouping.GetIDCol(), variables)
				if err != nil {
					return err
				}
				groupingVariableIndex = groupVariable.Index
			} else {
				groupingVariableIndex = variable.Index
			}
		}
	}

	// fetch the source dataset
	dataset, err := metaStorage.FetchDataset(s.Dataset, true, true)
	if err != nil {
		return nil
	}

	// fetch the input dataset (should only differ on augmented)
	datasetInput, err := metaStorage.FetchDataset(s.DatasetInput, true, true)
	if err != nil {
		return err
	}

	// timeseries specific handling - the target needs to be set to the timeseries Y field, and we need to
	// save timestamp variable index for data splitting
	targetVariable := s.TargetFeature
	if model.IsTimeSeries(targetVariable.Type) {
		tsg := targetVariable.Grouping.(*model.TimeseriesGrouping)
		// find the index of the timestamp variable of the timeseries
		timestampVariable, err := findVariable(tsg.XCol, dataVariables)
		if err != nil {
			return err
		}
		groupingVariableIndex = timestampVariable.Index

		// update the target variable to be the Y col of the timeseries group
		targetVariable, err = findVariable(tsg.YCol, dataVariables)
		if err != nil {
			return err
		}
	}

	// add dataset name to path
	datasetInputDir := env.ResolvePath(datasetInput.Source, datasetInput.Folder)

	// compute the task and subtask from the target and dataset
	task, err := ResolveTask(dataStorage, dataset.StorageName, s.TargetFeature, variables)
	if err != nil {
		return err
	}
	s.Task = task.Task

	// when dealing with categorical data we want to stratify
	stratify := model.IsCategorical(s.TargetFeature.Type)

	// perist the datasets and get URI
	params := &persistedDataParams{
		DatasetName:        s.DatasetInput,
		SchemaFile:         compute.D3MDataSchema,
		SourceDataFolder:   datasetInputDir,
		TmpDataFolder:      datasetDir,
		TaskType:           s.Task,
		GroupingFieldIndex: groupingVariableIndex,
		TargetFieldIndex:   targetVariable.Index,
		Stratify:           stratify,
	}
	datasetPathTrain, datasetPathTest, err := persistOriginalData(params)
	if err != nil {
		return err
	}

	// make sure the path is absolute and contains the URI prefix
	datasetPathTrain, err = filepath.Abs(datasetPathTrain)
	if err != nil {
		return err
	}
	datasetPathTrain = fmt.Sprintf("file://%s", datasetPathTrain)
	datasetPathTest, err = filepath.Abs(datasetPathTest)
	if err != nil {
		return err
	}
	datasetPathTest = fmt.Sprintf("file://%s", datasetPathTest)

	// generate the pre-processing pipeline to enforce feature selection and semantic type changes
	var preprocessing *pipeline.PipelineDescription
	if !client.SkipPreprocessing {
		preprocessing, err = s.createPreprocessingPipeline(dataVariables, metaStorage)
		if err != nil {
			return err
		}
	}

	// create search solutions request
	searchRequest, err := createSearchSolutionsRequest(targetVariable.Index, preprocessing, datasetPathTrain, client.UserAgent, targetVariable, s.DatasetInput, s.Metrics, s.Task, int64(s.MaxTime))
	if err != nil {
		return err
	}

	// start a solution searchID
	requestID, err := client.StartSearch(context.Background(), searchRequest)
	if err != nil {
		return err
	}

	// persist the request
	err = s.persistRequestStatus(s.requestChannel, solutionStorage, requestID, dataset.ID, RequestPendingStatus)
	if err != nil {
		return err
	}

	// store the request features - note that we are storing the original request filters, not the expanded
	// list that was generated
	for _, v := range s.Filters.Variables {
		var typ string
		// ignore the index field
		if v == model.D3MIndexFieldName {
			continue
		}

		if v == s.TargetFeature.Name {
			// store target feature
			typ = model.FeatureTypeTarget
		} else {
			// store training feature
			typ = model.FeatureTypeTrain
		}
		err = solutionStorage.PersistRequestFeature(requestID, v, typ)
		if err != nil {
			return err
		}
	}

	// store request filters
	err = solutionStorage.PersistRequestFilters(requestID, s.Filters)
	if err != nil {
		return err
	}

	// dispatch search request
	go s.dispatchRequest(client, solutionStorage, dataStorage, requestID, dataset.ID, searchRequest, datasetInputDir, datasetPathTrain, datasetPathTest, dataVariables)

	return nil
}

func findVariable(variableName string, variables []*model.Variable) (*model.Variable, error) {
	// extract the variable instance from its name
	var variable *model.Variable
	for _, v := range variables {
		if v.Name == variableName {
			variable = v
		}
	}
	if variable == nil {
		return nil, errors.Errorf("can't find target variable instance %s", variableName)
	}
	return variable, nil
}

func getFileFromOutput(response *pipeline.GetProduceSolutionResultsResponse, outputKey string) (string, error) {
	output, ok := response.ExposedOutputs[outputKey]
	if !ok {
		return "", errors.Errorf("output is missing from response")
	}

	csvURI, ok := output.Value.(*pipeline.Value_CsvUri)
	if !ok {
		return "", errors.Errorf("output is not of correct format")
	}

	return strings.Replace(csvURI.CsvUri, "file://", "", 1), nil
}
