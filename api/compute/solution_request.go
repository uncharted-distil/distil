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
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util/json"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
)

const (
	defaultExposedOutputKey = "outputs.0"
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

	defaultMaxSolution = 5
	defaultMaxTime     = 5
	defaultQuality     = "quality"

	// ModelQualityFast indicates that the system should try to generate models quickly at the expense of quality
	ModelQualityFast = "speed"
	// ModelQualityHigh indicates the the system should focus on higher quality models at the expense of speed
	ModelQualityHigh = "quality"
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
	Confidences              *api.SolutionExplainResult
	SolutionFeatureWeightURI string
	StepFeatureWeightURI     string
}

// SolutionRequest represents a solution search request.
type SolutionRequest struct {
	Dataset              string
	DatasetInput         string
	DatasetMetadata      *api.Dataset
	TargetFeature        *model.Variable
	Task                 []string
	TimestampField       string
	MaxSolutions         int
	MaxTime              int
	Quality              string
	ProblemType          string
	Metrics              []string
	Filters              *api.FilterParams
	DatasetAugmentations []*model.DatasetOrigin
	TrainTestSplit       float64

	mu               *sync.Mutex
	wg               *sync.WaitGroup
	requestChannel   chan SolutionStatus
	solutionChannels []chan SolutionStatus
	listener         SolutionStatusListener
	finished         chan error
	useParquet       bool
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
	req.MaxSolutions = json.IntDefault(j, defaultMaxSolution, "maxSolutions")
	req.MaxTime = json.IntDefault(j, defaultMaxTime, "maxTime")
	req.Quality = json.StringDefault(j, defaultQuality, "quality")
	req.ProblemType = json.StringDefault(j, "", "problemType")
	req.Metrics, _ = json.StringArray(j, "metrics")
	req.TrainTestSplit = json.FloatDefault(j, 0.9, "trainTestSplit")

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
	return createSearchSolutionsRequest(columnIndex, preprocessing, datasetURI, userAgent, s.TargetFeature, s.Dataset,
		s.Metrics, s.Task, int64(s.MaxTime), int64(s.MaxSolutions))
}

func createSearchSolutionsRequest(columnIndex int, preprocessing *pipeline.PipelineDescription,
	datasetURI string, userAgent string, targetFeature *model.Variable, dataset string, metrics []string, task []string,
	maxTime int64, maxSolutions int64) (*pipeline.SearchSolutionsRequest, error) {

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
func GeneratePredictions(datasetURI string, explainedSolutionID string,
	fittedSolutionID string, client *compute.Client) (*PredictionResult, error) {
	// check if the solution can be explained
	desc, err := client.GetSolutionDescription(context.Background(), explainedSolutionID)
	if err != nil {
		return nil, err
	}

	outputs, err := getPipelineOutputs(desc)
	if err != nil {
		return nil, err
	}

	keys := []string{defaultExposedOutputKey}
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

		resultURI, err := getFileFromOutput(response, defaultExposedOutputKey)
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

func (s *SolutionRequest) persistSolutionResults(statusChan chan SolutionStatus, client *compute.Client, solutionStorage api.SolutionStorage,
	dataStorage api.DataStorage, initialSearchID string, dataset string, storageName string, initialSearchSolutionID string,
	fittedSolutionID string, produceRequestID string, resultID string, resultURI string) {
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
	err = dataStorage.PersistResult(dataset, storageName, resultURI, s.TargetFeature.Name)
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}
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
	err := s.persistRequestStatus(s.requestChannel, solutionStorage, searchContext.searchID, searchContext.dataset, RequestRunningStatus)
	if err != nil {
		s.finished <- err
		return
	}

	// search for solutions, this wont return until the search finishes or it times out
	err = client.SearchSolutions(context.Background(), searchContext.searchID, func(solution *pipeline.GetSearchSolutionsResultsResponse) {
		searchContext.searchSolutionID = solution.SolutionId
		// create a new status channel for the solution
		c := newStatusChannel()
		// add the solution to the request
		s.addSolution(c)
		// persist the solution
		s.persistSolution(c, solutionStorage, searchContext.searchID, solution.SolutionId, "")
		s.persistSolutionStatus(c, solutionStorage, searchContext.searchID, solution.SolutionId, SolutionPendingStatus)
		// dispatch it
		searchContext.searchResult, err = s.dispatchSolutionSearchPipeline(c, client, solutionStorage, dataStorage, searchContext)
		if err != nil {
			s.persistSolutionError(c, solutionStorage, searchContext.searchID, solution.SolutionId, err)
			return
		}

		err = s.dispatchSolutionExplainPipeline(client, solutionStorage, dataStorage, searchContext)
		if err != nil {
			s.persistSolutionError(c, solutionStorage, searchContext.searchID, solution.SolutionId, err)
			return
		}

		// notify client of update
		c <- SolutionStatus{
			RequestID:  searchContext.searchID,
			SolutionID: solution.SolutionId,
			ResultID:   searchContext.searchResult.resultID,
			Progress:   SolutionCompletedStatus,
			Timestamp:  time.Now(),
		}
		// once done, mark as complete
		s.completeSolution()
		close(c)
	})

	// wait until all are complete and the search has finished / timed out
	s.waitOnSolutions()

	// update request status
	if err != nil {
		s.persistRequestError(s.requestChannel, solutionStorage, searchContext.searchID, searchContext.dataset, err)
	} else {
		s.persistRequestStatus(s.requestChannel, solutionStorage, searchContext.searchID, searchContext.dataset, RequestCompletedStatus)
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
	dataset, err := metaStorage.FetchDataset(s.Dataset, true, true)
	if err != nil {
		return nil
	}

	// fetch the input dataset (should only differ on augmented)
	datasetInput, err := metaStorage.FetchDataset(s.DatasetInput, true, true)
	if err != nil {
		return err
	}
	s.DatasetMetadata = datasetInput

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
	var featurizedVariables []*model.Variable
	var datasetInputDir string
	targetIndex := -1
	if datasetInput.LearningDataset == "" {
		datasetInputDir = datasetInput.Folder
		datasetInputDir = env.ResolvePath(datasetInput.Source, datasetInputDir)
		targetIndex = targetVariable.Index
	} else {
		s.useParquet = true
		datasetInputDir = datasetInput.LearningDataset
		groupingVariableIndex = -1

		// need to lookup the target variable and the variables in the featurized dataset
		meta, err := metadata.LoadMetadataFromOriginalSchema(path.Join(datasetInputDir, compute.D3MDataSchema), false)
		if err != nil {
			return err
		}
		featurizedVariables = meta.GetMainDataResource().Variables

		for _, v := range featurizedVariables {
			if v.Name == targetVariable.Name {
				targetIndex = v.Index
				break
			}
		}
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

	// when dealing with categorical data we want to stratify
	stratify := model.IsCategorical(s.TargetFeature.Type)
	// create the splitter to use for the train / test split
	splitter := createSplitter(s.Task, targetIndex, groupingVariableIndex, stratify, s.Quality, s.TrainTestSplit)
	datasetPathTrain, datasetPathTest, err := SplitDataset(path.Join(datasetInputDir, compute.D3MDataSchema), splitter)
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
		if datasetInput.LearningDataset == "" {
			preprocessing, err = s.createPreprocessingPipeline(variables, metaStorage)
		} else {
			preprocessing, err = s.createPreFeaturizedPipeline(datasetInput.LearningDataset, variables, featurizedVariables, metaStorage, targetIndex)
		}
		if err != nil {
			return err
		}
	}

	// create search solutions request
	searchRequest, err := createSearchSolutionsRequest(targetIndex, preprocessing, datasetPathTrain, client.UserAgent,
		targetVariable, s.DatasetInput, s.Metrics, s.Task, int64(s.MaxTime), int64(s.MaxSolutions))
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
	searchContext := pipelineSearchContext{
		searchID:          requestID,
		dataset:           dataset.ID,
		storageName:       datasetInput.StorageName,
		sourceDatasetURI:  datasetInputDir,
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

func findVariable(variableName string, variables []*model.Variable) (*model.Variable, error) {
	// extract the variable instance from its name
	for _, v := range variables {
		if v.Name == variableName {
			return v, nil
		}
	}
	return nil, errors.Errorf("can't find target variable instance %s", variableName)
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
