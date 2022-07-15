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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	encjson "encoding/json"

	uuid "github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
	"github.com/uncharted-distil/distil/api/util/json"
	log "github.com/unchartedsoftware/plog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultMaxSolution = 5
	defaultMaxTime     = 5
	defaultQuality     = "quality"

	// ModelQualityFast indicates that the system should try to generate models quickly at the expense of quality
	ModelQualityFast = "speed"
	// ModelQualityHigh indicates the the system should focus on higher quality models at the expense of speed
	ModelQualityHigh = "quality"
)

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
	Confidences              *api.SolutionExplainResult
	SolutionFeatureWeightURI string
	StepFeatureWeightURI     string
}

// SolutionRequest represents a solution search request.
type SolutionRequest struct {
	Dataset              string
	DatasetMetadata      *api.Dataset
	TargetFeature        *model.Variable
	Task                 []string
	TimestampField       string
	TimestampSplitValue  float64
	MaxSolutions         int
	MaxTime              int
	Quality              string
	ProblemType          string
	Metrics              []string
	Filters              *api.FilterParams
	DatasetAugmentations []*model.DatasetOrigin
	TrainTestSplit       float64
	CancelFuncs          map[string]context.CancelFunc
	PosLabel             string
	mu                   *sync.Mutex
	wg                   *sync.WaitGroup
	requestChannel       chan SolutionStatus
	solutionChannels     []chan SolutionStatus
	listener             SolutionStatusListener
	finished             chan error
	useParquet           bool
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

	targetKey, ok := json.String(j, "target")
	if !ok {
		return nil, fmt.Errorf("no `target` in solution request")
	}
	for _, v := range variables {
		if v.Key == targetKey {
			req.TargetFeature = v
		}
	}

	req.Task, _ = json.StringArray(j, "task")
	req.MaxSolutions = json.IntDefault(j, defaultMaxSolution, "maxSolutions")
	req.MaxTime = json.IntDefault(j, defaultMaxTime, "maxTime")
	req.Quality = json.StringDefault(j, defaultQuality, "quality")
	req.ProblemType = json.StringDefault(j, "", "problemType")
	req.Metrics, _ = json.StringArray(j, "metrics")
	req.TrainTestSplit = json.FloatDefault(j, 0.9, "trainTestSplit")
	req.TimestampSplitValue = json.FloatDefault(j, 0.0, "timestampSplitValue")
	posLabel, ok := json.String(j, "positiveLabel")
	if ok {
		req.PosLabel = posLabel
	}
	filters, ok := json.Get(j, "filters")
	if ok {
		rawFilters, err := api.ParseFilterParamsFromJSON(filters)
		if err != nil {
			return nil, err
		}
		req.Filters = rawFilters
	}

	req.CancelFuncs = map[string]context.CancelFunc{}

	return req, nil
}

// ExtractDatasetFromRawRequest extracts the dataset name from the raw message.
func ExtractDatasetFromRawRequest(data encjson.RawMessage) (string, error) {
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

// Cancel inovkes the context cancel function calls associated with this request.  This stops any
// further messaging between the ta3 and ta2 for each solution.
func (s *SolutionRequest) Cancel() {
	// Cancel all further work for each solution
	for _, cancelFunc := range s.CancelFuncs {
		cancelFunc()
	}
}

func createSearchSolutionsRequest(preprocessing *pipeline.PipelineDescription, datasetURI string,
	userAgent string, targetFeature *model.Variable, dataset string, metrics []string, task []string,
	maxTime int64, maxSolutions int64, posLabel string) (*pipeline.SearchSolutionsRequest, error) {

	return &pipeline.SearchSolutionsRequest{
		Problem: &pipeline.ProblemDescription{
			Problem: &pipeline.Problem{
				TaskKeywords:       compute.ConvertTaskKeywordsFromTA3ToTA2(task),
				PerformanceMetrics: compute.ConvertMetricsFromTA3ToTA2(metrics, posLabel),
			},
			Inputs: []*pipeline.ProblemInput{
				{
					DatasetId: compute.ConvertDatasetTA3ToTA2(dataset),
					Targets:   compute.ConvertTargetFeaturesTA3ToTA2(targetFeature.HeaderName, targetFeature.Index),
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

		// Request maximum number of solutions
		RankSolutionsLimit: int32(maxSolutions),

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
	expandedFilters, err := api.ExpandFilterParams(s.Dataset, s.Filters, true, metaStorage)
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
func GeneratePredictions(datasetURI string, solutionID string, fittedSolutionID string, client *compute.Client) (*PredictionResult, error) {
	// check if the solution can be explained
	desc, err := client.GetSolutionDescription(context.Background(), solutionID)
	if err != nil {
		return nil, err
	}

	outputs, err := getPipelineOutputs(desc)
	if err != nil {
		return nil, err
	}

	keys := []string{compute.DefaultExposedOutputKey}
	keys = append(keys, extractOutputKeys(outputs)...)

	produceRequest := createProduceSolutionRequest(datasetURI, fittedSolutionID, keys, nil)
	produceRequestID, predictionResponses, err := client.GeneratePredictions(context.Background(), produceRequest)
	if err != nil {
		return nil, err
	}

	for _, response := range predictionResponses {

		if response.Progress.State != pipeline.ProgressState_COMPLETED {
			// only persist completed responses
			continue
		}

		resultURI, err := getFileFromOutput(response, compute.DefaultExposedOutputKey)
		if err != nil {
			return nil, err
		}
		resultURI, err = reformatResult(resultURI)
		if err != nil {
			return nil, err
		}

		var explainFeatureURI string
		if outputs[ExplainableTypeStep] != nil {
			explainFeatureURI, err = getFileFromOutput(response, outputs[ExplainableTypeStep].key)
			if err != nil {
				return nil, err
			}
		}
		var explainSolutionURI string
		if outputs[ExplainableTypeSolution] != nil {
			explainSolutionURI, err = getFileFromOutput(response, outputs[ExplainableTypeSolution].key)
			if err != nil {
				return nil, err
			}
		}
		var confidenceResult *api.SolutionExplainResult
		if outputs[ExplainableTypeConfidence] != nil {
			confidenceURI, err := getFileFromOutput(response, outputs[ExplainableTypeConfidence].key)
			if err != nil {
				return nil, err
			}
			confidenceResult, err = ExplainFeatureOutput(resultURI, confidenceURI)
			if err != nil {
				return nil, err
			}
			confidenceResult.ParsingFunction = outputs[ExplainableTypeConfidence].parsingFunction
		}

		return &PredictionResult{
			ProduceRequestID:         produceRequestID,
			FittedSolutionID:         fittedSolutionID,
			ResultURI:                resultURI,
			Confidences:              confidenceResult,
			StepFeatureWeightURI:     explainFeatureURI,
			SolutionFeatureWeightURI: explainSolutionURI,
		}, nil
	}

	return nil, errors.Errorf("no results retrieved")
}

func createProduceSolutionRequest(datasetURI string, fittedSolutionID string, outputs []string, exposeValueTypes []string) *pipeline.ProduceSolutionRequest {
	evt := []string{}
	evt = append(evt, exposeValueTypes...)
	evt = append(evt, compute.CSVURIValueType)
	return &pipeline.ProduceSolutionRequest{
		FittedSolutionId: fittedSolutionID,
		Inputs: []*pipeline.Value{
			{
				Value: &pipeline.Value_DatasetUri{
					DatasetUri: compute.BuildSchemaFileURI(datasetURI),
				},
			},
		},
		ExposeOutputs:    outputs,
		ExposeValueTypes: evt,
	}
}

func createFitSolutionRequest(datasetURI string, fittedSolutionID string) *pipeline.FitSolutionRequest {
	return &pipeline.FitSolutionRequest{
		SolutionId: fittedSolutionID,
		Inputs: []*pipeline.Value{
			{
				Value: &pipeline.Value_DatasetUri{
					DatasetUri: compute.BuildSchemaFileURI(datasetURI),
				},
			},
		},
	}
}

func (s *SolutionRequest) persistSolutionError(statusChan chan SolutionStatus, solutionStorage api.SolutionStorage, searchID string, solutionID string, err error) {
	log.Errorf("solution '%s' errored: %v", solutionID, err)

	// Check to see if this is a cancellation error and use a specific code for it if so
	progress := compute.SolutionErroredStatus
	cause := errors.Cause(err)
	st, ok := status.FromError(cause)
	if ok && st.Code() == codes.Canceled {
		progress = compute.SolutionCancelledStatus
	}

	// persist the updated state
	// NOTE: ignoring error
	_ = solutionStorage.PersistSolutionState(solutionID, progress, time.Now())

	// notify of error
	statusChan <- SolutionStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		Progress:   progress,
		Error:      err,
		Timestamp:  time.Now(),
	}
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
	_ = solutionStorage.PersistRequest(searchID, dataset, compute.RequestErroredStatus, time.Now())

	// notify of error
	statusChan <- SolutionStatus{
		RequestID: searchID,
		Progress:  compute.RequestErroredStatus,
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

func (s *SolutionRequest) persistSolutionResults(statusChan chan SolutionStatus, client *compute.Client, solutionStorage api.SolutionStorage,
	dataStorage api.DataStorage, initialSearchID string, dataset string, storageName string, initialSearchSolutionID string,
	fittedSolutionID string, produceRequestID string, resultID string, resultURI string) error {
	// persist the completed state
	err := solutionStorage.PersistSolutionState(initialSearchSolutionID, compute.SolutionCompletedStatus, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return err
	}
	// persist result metadata
	err = solutionStorage.PersistSolutionResult(initialSearchSolutionID, fittedSolutionID, produceRequestID, "test", resultID, resultURI, compute.SolutionCompletedStatus, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return err
	}
	// persist results
	err = dataStorage.PersistResult(dataset, storageName, resultURI, s.TargetFeature.Key)
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return err
	}

	return nil
}

func describeSolution(client *compute.Client, initialSearchSolutionID string) (*pipeline.DescribeSolutionResponse, error) {
	// need to wait until a valid description is returned before proceeding
	var desc *pipeline.DescribeSolutionResponse
	var err error
	for wait := true; wait; {
		desc, err = client.GetSolutionDescription(context.Background(), initialSearchSolutionID)
		if err != nil {
			return nil, err
		}
		wait = desc == nil || desc.Pipeline == nil
		if wait {
			time.Sleep(10 * time.Second)
		}
	}

	return desc, nil
}

func (s *SolutionRequest) dispatchRequest(client *compute.Client, solutionStorage api.SolutionStorage,
	dataStorage api.DataStorage, searchContext pipelineSearchContext) {
	// update request status
	err := s.persistRequestStatus(s.requestChannel, solutionStorage, searchContext.searchID, searchContext.dataset, compute.RequestRunningStatus)
	if err != nil {
		s.finished <- err
		return
	}

	// search for solutions, this wont return until the search finishes or it times out
	err = client.SearchSolutions(context.Background(), searchContext.searchID, func(solution *pipeline.GetSearchSolutionsResultsResponse) {
		// create a new status channel for the solution
		c := newStatusChannel()
		// add the solution to the request
		s.addSolution(c)
		// persist the solution
		s.persistSolution(c, solutionStorage, searchContext.searchID, solution.SolutionId, "")
		s.persistSolutionStatus(c, solutionStorage, searchContext.searchID, solution.SolutionId, compute.SolutionPendingStatus)

		// once done, mark as complete and clean up the channel
		defer func() {
			s.completeSolution()
			close(c)
		}()

		// dispatch it
		searchResult, err := s.dispatchSolutionSearchPipeline(c, client, solutionStorage, dataStorage, solution.SolutionId, searchContext)
		if err != nil {
			s.persistSolutionError(c, solutionStorage, searchContext.searchID, solution.SolutionId, err)
			return
		}

		err = s.dispatchSolutionExplainPipeline(client, solutionStorage, dataStorage, solution.SolutionId, searchContext, searchResult)
		if err != nil {
			s.persistSolutionError(c, solutionStorage, searchContext.searchID, solution.SolutionId, err)
			return
		}

		// notify client of update
		c <- SolutionStatus{
			RequestID:  searchContext.searchID,
			SolutionID: solution.SolutionId,
			ResultID:   searchResult.resultID,
			Progress:   compute.SolutionCompletedStatus,
			Timestamp:  time.Now(),
		}
	})

	// wait until all are complete and the search has finished / timed out
	s.waitOnSolutions()

	// update request status
	if err != nil {
		s.persistRequestError(s.requestChannel, solutionStorage, searchContext.searchID, searchContext.dataset, err)
	} else {
		if err = s.persistRequestStatus(s.requestChannel, solutionStorage, searchContext.searchID, searchContext.dataset, compute.RequestCompletedStatus); err != nil {
			log.Errorf("failed to persist status %s for search %s", compute.RequestCompletedStatus, searchContext.searchID)
		}
	}
	close(s.requestChannel)

	// end search
	// since predictions can be requested for different datasets on the same
	// fitted solution, can't tell TA2 to end but the channel still needs
	// to be notified that the current process is complete
	//s.finished <- client.EndSearch(context.Background(), searchID)
	s.finished <- nil
}

func dispatchSegmentation(s *SolutionRequest, requestID string, solutionStorage api.SolutionStorage, metaStorage api.MetadataStorage,
	dataStorage api.DataStorage, client *compute.Client, datasetInputDir string, step *description.FullySpecifiedPipeline) {
	log.Infof("dispatching segmentation pipeline")

	// create the backing data
	err := s.persistRequestStatus(s.requestChannel, solutionStorage, requestID, s.Dataset, compute.RequestRunningStatus)
	if err != nil {
		s.finished <- err
		return
	}

	c := newStatusChannel()

	// run the pipeline
	pipelineResult, err := SubmitPipeline(client, []string{datasetInputDir}, nil, nil, step, nil, true)
	if err != nil {
		s.finished <- err
		return
	}

	// add the solution to the request
	// doing this after submission to have the solution id available!
	s.addSolution(c)
	s.persistSolution(c, solutionStorage, requestID, pipelineResult.SolutionID, "")
	s.persistSolutionStatus(c, solutionStorage, requestID, pipelineResult.SolutionID, compute.SolutionPendingStatus)
	s.persistSolutionStatus(c, solutionStorage, requestID, pipelineResult.SolutionID, compute.SolutionScoringStatus)

	// HACK: MAKE UP A SOLUTION SCORE!!!
	err = solutionStorage.PersistSolutionScore(pipelineResult.SolutionID, util.F1Micro, 0.5)
	if err != nil {
		s.finished <- err
		return
	}
	s.persistSolutionStatus(c, solutionStorage, requestID, pipelineResult.SolutionID, compute.SolutionProducingStatus)

	// update status and respond to client as needed
	uuidGen, err := uuid.NewV4()
	if err != nil {
		s.finished <- errors.Wrapf(err, "unable to generate solution id")
		return
	}
	resultID := uuidGen.String()
	c <- SolutionStatus{
		RequestID:  requestID,
		SolutionID: pipelineResult.SolutionID,
		ResultID:   resultID,
		Progress:   compute.SolutionCompletedStatus,
		Timestamp:  time.Now(),
	}
	close(c)

	// read the file and parse the output mask
	log.Infof("processing segmentation pipeline output")
	result, err := result.ParseResultCSV(pipelineResult.ResultURI)
	if err != nil {
		s.finished <- err
		return
	}

	images, err := BuildSegmentationImage(result)
	if err != nil {
		s.finished <- err
		return
	}

	// get the grouping key since it makes up part of the filename
	dataset, err := metaStorage.FetchDataset(s.Dataset, true, true, false)
	if err != nil {
		s.finished <- err
		return
	}

	var groupingKey *model.Variable
	for _, v := range dataset.Variables {
		if v.HasRole(model.VarDistilRoleGrouping) {
			groupingKey = v
			break
		}
	}
	if groupingKey == nil {
		s.finished <- errors.Errorf("no grouping found to use for output filename")
		return
	}

	// get the d3m index -> grouping key mapping
	mapping, err := api.BuildFieldMapping(dataset.ID, dataset.StorageName, model.D3MIndexFieldName, groupingKey.Key, dataStorage)
	if err != nil {
		s.finished <- err
		return
	}

	imageOutputFolder := path.Join(env.GetResourcePath(), dataset.ID, "media")
	for d3mIndex, imageBytes := range images {
		imageFilename := path.Join(imageOutputFolder, fmt.Sprintf("%s-segmentation.png", mapping[d3mIndex]))
		err = util.WriteFileWithDirs(imageFilename, imageBytes, os.ModePerm)
		if err != nil {
			s.finished <- err
			return
		}
	}

	// HACK:	INPUT FAKE RESULTS TO THE DB!!!
	//				FAKE RESULTS SHOULD JUST BE A CONSTANT!
	uuidGen, err = uuid.NewV4()
	if err != nil {
		s.finished <- errors.Wrapf(err, "unable to generate produce request id")
		return
	}
	produceRequestID := uuidGen.String()

	// HACK:	CREATE FAKE RESULTS TO PERSIST AS THE ACTUAL RESULTS SHOULD NOT BE STORED IN THE DB!!!
	resultOutput := []string{fmt.Sprintf("%s,%s,%s", model.D3MIndexFieldName, s.TargetFeature.HeaderName, "confidence")}
	for i := 1; i < len(result); i++ {
		resultOutput = append(resultOutput, fmt.Sprintf("%s,%s,%d", result[i][0].(string), "segmented", 1))
	}
	resultOutputURI := fmt.Sprintf("%s-distil-%s",
		pipelineResult.ResultURI[:len(pipelineResult.ResultURI)-4], pipelineResult.ResultURI[len(pipelineResult.ResultURI)-4:])
	log.Infof("writing distil formatted segmentation results to '%s'", resultOutputURI)
	err = util.WriteFileWithDirs(resultOutputURI, []byte(strings.Join(resultOutput, "\n")), os.ModePerm)
	if err != nil {
		s.finished <- err
		return
	}

	log.Infof("persisting results in URI '%s'", resultOutputURI)
	err = s.persistSolutionResults(c, client, solutionStorage, dataStorage, requestID, dataset.ID,
		dataset.StorageName, pipelineResult.SolutionID, pipelineResult.FittedSolutionID, produceRequestID, resultID, resultOutputURI)
	if err != nil {
		s.finished <- errors.Wrapf(err, "unable to persist solution result")
		return
	}

	log.Infof("segmentation pipeline processing complete")

	err = s.persistRequestStatus(s.requestChannel, solutionStorage, requestID, dataset.ID, compute.RequestCompletedStatus)
	if err != nil {
		s.finished <- err
		return
	}
	close(s.requestChannel)
	s.finished <- nil
}

func processSegmentation(s *SolutionRequest, client *compute.Client, solutionStorage api.SolutionStorage, metaStorage api.MetadataStorage, dataStorage api.DataStorage) error {
	// create the fully specified pipeline
	envConfig, err := env.LoadConfig()
	if err != nil {
		return err
	}

	// fetch the source dataset
	dataset, err := metaStorage.FetchDataset(s.Dataset, true, true, false)
	if err != nil {
		return nil
	}
	s.DatasetMetadata = dataset
	variablesMap := api.MapVariables(dataset.Variables, func(v *model.Variable) string { return v.Key })

	datasetInputDir := env.ResolvePath(dataset.Source, dataset.Folder)

	step, err := description.CreateRemoteSensingSegmentationPipeline("segmentation", "basic image segmentation", s.TargetFeature, envConfig.RemoteSensingNumJobs)
	if err != nil {
		return err
	}

	// need a request ID
	uuidGen, err := uuid.NewV4()
	if err != nil {
		return err
	}
	requestID := uuidGen.String()

	// persist the request
	err = s.persistRequestStatus(s.requestChannel, solutionStorage, requestID, dataset.ID, compute.RequestPendingStatus)
	if err != nil {
		return err
	}

	// store the request features - note that we are storing the original request filters, not the expanded
	// list that was generated
	// also note that augmented features should not be included
	for _, v := range s.Filters.Variables {
		var typ string
		// ignore the index field
		if v == model.D3MIndexFieldName {
			continue
		} else if variablesMap[v].HasRole(model.VarDistilRoleAugmented) {
			continue
		}

		if v == s.TargetFeature.Key {
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

	// store the original request filters
	// HACK: NO FILTERS SUPPORTED FOR SEGMENTATION!
	err = solutionStorage.PersistRequestFilters(requestID, s.Filters)
	if err != nil {
		return err
	}

	// dispatch it as if it were a model search
	go dispatchSegmentation(s, requestID, solutionStorage, metaStorage, dataStorage, client, datasetInputDir, step)

	return nil
}

// PersistAndDispatch persists the solution request and dispatches it.
func (s *SolutionRequest) PersistAndDispatch(client *compute.Client, solutionStorage api.SolutionStorage, metaStorage api.MetadataStorage, dataStorage api.DataStorage) error {

	// fetch the dataset variables
	variables, err := metaStorage.FetchVariables(s.Dataset, true, true, false)
	if err != nil {
		return err
	}
	variablesMap := api.MapVariables(variables, func(v *model.Variable) string { return v.Key })

	// NOTE: D3M index field is needed in the persisted data.
	d3mIndexIncluded := false
	for _, v := range s.Filters.Variables {
		if v == model.D3MIndexFieldName {
			d3mIndexIncluded = true
			break
		}
	}
	if !d3mIndexIncluded {
		s.Filters.Variables = append(s.Filters.Variables, model.D3MIndexFieldName)
	}

	// remove any generated / grouped features from our var list
	// TODO: imported datasets have d3m index as distil role = "index".
	//       need to figure out if that causes issues!!!
	dataVariables := []*model.Variable{}
	groupingVariableIndex := -1
	for _, variable := range variables {
		if variable.IsTA2Field() {
			dataVariables = append(dataVariables, variable)
		}
		if variable.HasRole(model.VarDistilRoleGrouping) {
			// if this is a group var, find the grouping ID col and use that
			if variable.Grouping != nil && variable.Grouping.GetIDCol() != "" {
				groupVariable, err := findVariable(variable.Grouping.GetIDCol(), variables)
				if err != nil {
					return err
				}
				groupingVariableIndex = groupVariable.Index
			}
		}
	}
	// fetch the source dataset
	dataset, err := metaStorage.FetchDataset(s.Dataset, true, true, false)
	if err != nil {
		return nil
	}
	s.DatasetMetadata = dataset

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

	// prefilter dataset if metadata fields are used in filters
	filteredDatasetPath, updatedFilters, err := filterData(client, dataset, s.Filters, dataStorage)
	if err != nil {
		return err
	}
	s.Filters = updatedFilters

	if dataset.LearningDataset != "" {
		s.useParquet = true
		groupingVariableIndex = -1
	}

	// compute the task and subtask from the target and dataset
	trainingVariables, err := findVariables(s.Filters.Variables, variables)
	if err != nil {
		return err
	}
	task, err := ResolveTask(dataStorage, dataset.StorageName, s.TargetFeature, trainingVariables)
	if err != nil {
		return err
	}
	s.Task = task.Task

	if HasTaskType(task, compute.SegmentationTask) {
		return processSegmentation(s, client, solutionStorage, metaStorage, dataStorage)
	}

	// check if TimestampSplitValue is not 0
	if s.TimestampSplitValue > 0 {
		found := false

		// update groupingVariable to the dateTime variable
		for _, variable := range variables {
			if variable.Type == model.DateTimeType {
				groupingVariableIndex = variable.Index
				found = true
				break
			}
		}
		// if not found return error, dateTime type required for split
		if !found {
			return errors.New("Timestamp value supplied but no dateTime type existing on dataset")
		}
	}

	// get the target
	meta, err := serialization.ReadMetadata(path.Join(filteredDatasetPath, compute.D3MDataSchema))
	if err != nil {
		return err
	}
	metaVars := meta.GetMainDataResource().Variables
	targetVariable, err = findVariable(targetVariable.Key, metaVars)
	if err != nil {
		return err
	}

	// when dealing with categorical data we want to stratify
	stratify := model.IsCategorical(s.TargetFeature.Type)
	// create the splitter to use for the train / test split
	splitter := createSplitter(s.Task, targetVariable.Index, groupingVariableIndex, stratify, s.Quality, s.TrainTestSplit, s.TimestampSplitValue)
	datasetPathTrain, datasetPathTest, err := SplitDataset(path.Join(filteredDatasetPath, compute.D3MDataSchema), splitter)
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

	// keep original filters to store them to the database
	originalFilters := s.Filters

	// get filters that map filters on groupings to the underlying field
	s.Filters = mapFilterKeys(s.Dataset, s.Filters, dataset.Variables)

	// generate the pre-processing pipeline to enforce feature selection and semantic type changes
	var preprocessing *pipeline.PipelineDescription
	if !client.SkipPreprocessing {
		if dataset.LearningDataset == "" {
			preprocessing, err = s.createPreprocessingPipeline(variables, metaStorage)
		} else {
			preprocessing, err = s.createPreFeaturizedPipeline(dataset.LearningDataset, variables, metaVars, metaStorage, targetVariable.Index)
		}
		if err != nil {
			return err
		}
	}

	// create search solutions request
	searchRequest, err := createSearchSolutionsRequest(preprocessing, datasetPathTrain, client.UserAgent,
		targetVariable, s.Dataset, s.Metrics, s.Task, int64(s.MaxTime), int64(s.MaxSolutions), s.PosLabel)
	if err != nil {
		return err
	}

	// start a solution searchID
	requestID, err := client.StartSearch(context.Background(), searchRequest)
	if err != nil {
		return err
	}

	// persist the request
	err = s.persistRequestStatus(s.requestChannel, solutionStorage, requestID, dataset.ID, compute.RequestPendingStatus)
	if err != nil {
		return err
	}

	// store the request features - note that we are storing the original request filters, not the expanded
	// list that was generated
	// also note that augmented features should not be included
	for _, v := range s.Filters.Variables {
		var typ string
		// ignore the index field
		if v == model.D3MIndexFieldName {
			continue
		} else if variablesMap[v].HasRole(model.VarDistilRoleAugmented) {
			continue
		}

		if v == s.TargetFeature.Key {
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

	// store the original request filters
	err = solutionStorage.PersistRequestFilters(requestID, originalFilters)
	if err != nil {
		return err
	}

	// dispatch search request
	searchContext := pipelineSearchContext{
		searchID:          requestID,
		dataset:           dataset.ID,
		storageName:       dataset.StorageName,
		sourceDatasetURI:  filteredDatasetPath,
		trainDatasetURI:   datasetPathTrain,
		testDatasetURI:    datasetPathTest,
		produceDatasetURI: datasetPathTest,
		variables:         dataVariables,
		targetCol:         s.TargetFeature.Index,
		groupingCol:       groupingVariableIndex,
		sample:            true,
	}

	// generate predictions -  for timeseries we want to use the entire source dataset, for anything else
	// we only want the test data predictions.
	for _, task := range s.Task {
		if task == compute.ForecastingTask {
			searchContext.produceDatasetURI = compute.BuildSchemaFileURI(searchContext.sourceDatasetURI)
			searchContext.sample = false
			break
		}
	}
	go s.dispatchRequest(client, solutionStorage, dataStorage, searchContext)

	return nil
}

func findVariable(key string, variables []*model.Variable) (*model.Variable, error) {
	// extract the variable instance from its name
	for _, v := range variables {
		if v.Key == key {
			return v, nil
		}
	}
	return nil, errors.Errorf("can't find target variable instance %s", key)
}

func findVariables(variableNames []string, variables []*model.Variable) ([]*model.Variable, error) {
	filterVariables := make([]*model.Variable, len(variableNames))
	for i, varName := range variableNames {
		var err error
		filterVariables[i], err = findVariable(varName, variables)
		if err != nil {
			return nil, err
		}
	}
	return filterVariables, nil
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

type confidenceValue struct {
	d3mIndex   string
	confidence float64
	row        int
}

func reformatResult(resultURI string) (string, error) {
	// read data from original file
	dataReader := serialization.GetStorage(resultURI)
	data, err := dataReader.ReadData(resultURI)
	if err != nil {
		return "", err
	}

	// only need to reformat if confidences are there (column count >= 3)
	if len(data[0]) < 3 {
		return resultURI, nil
	}
	log.Infof("reformatting '%s' to only have 1 row per d3m index", resultURI)

	// TODO: CAN WE ASSUME THESE INDICES???
	d3mIndexIndex := 0
	confidenceIndex := 2
	confidences := map[string]*confidenceValue{}
	output := [][]string{data[0]}
	for _, r := range data[1:] {
		// only keep the row with the highest confidence for each d3m index
		d3mIndex := r[d3mIndexIndex]
		confidenceParsed, err := strconv.ParseFloat(r[confidenceIndex], 64)
		if err != nil {
			return "", errors.Wrapf(err, "unable to parse confidence value '%s'", r[confidenceIndex])
		}
		confidence := confidences[d3mIndex]
		if confidence == nil || confidence.confidence < confidenceParsed {
			row := len(output)
			if confidence != nil {
				// new top confidence so overwrite existing entry in output
				row = confidence.row
				output[row] = r
			} else {
				// new d3m index so append to output
				output = append(output, r)
			}
			confidence := &confidenceValue{
				d3mIndex:   d3mIndex,
				confidence: confidenceParsed,
				row:        row,
			}
			confidences[d3mIndex] = confidence
		}
	}

	// output filtered data
	filteredURI := path.Join(path.Dir(resultURI), fmt.Sprintf("filtered-%s", path.Base(resultURI)))
	err = dataReader.WriteData(filteredURI, output)
	if err != nil {
		return "", err
	}
	log.Infof("'%s' filtered to highest confidence row per d3m index and written to '%s'", resultURI, filteredURI)

	return filteredURI, nil
}
