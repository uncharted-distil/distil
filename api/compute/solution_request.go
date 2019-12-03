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
	"crypto/sha1"
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
	"github.com/uncharted-distil/distil/api/util/json"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
)

const (
	defaultResourceID       = "learningData"
	defaultExposedOutputKey = "outputs.0"
	trainTestSplitThreshold = 0.9
	// SolutionPendingStatus represents that the solution request has been acknoledged by not yet sent to the API
	SolutionPendingStatus = "SOLUTION_PENDING"
	// SolutionRunningStatus represents that the solution request has been sent to the API.
	SolutionRunningStatus = "SOLUTION_RUNNING"
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

	ta2RoleMap = map[string]bool{
		"data":  true,
		"index": true,
	}
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

// SolutionRequest represents a solution search request.
type SolutionRequest struct {
	Dataset              string
	DatasetInput         string
	TargetFeature        *model.Variable
	Task                 string
	SubTask              string
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

// SolutionRequestDiscovery represents a discovered problem solution request.
type SolutionRequestDiscovery struct {
	Dataset              string
	DatasetInput         string
	TargetFeature        *model.Variable
	AllFeatures          []*model.Variable
	SelectedFeatures     []string
	SourceURI            string
	UserAgent            string
	DatasetAugmentations []*model.DatasetOrigin
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

	req.Task = json.StringDefault(j, "", "task")
	req.SubTask = json.StringDefault(j, "", "subTask")
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

func (s *SolutionRequest) listenOnStatusChannel(statusChannel chan SolutionStatus) {
	for {
		// read status from, channel
		status := <-statusChannel
		// execute callback
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
	return createSearchSolutionsRequest(columnIndex, preprocessing, datasetURI, userAgent, s.TargetFeature, s.Dataset, s.Metrics, s.Task, s.SubTask, int64(s.MaxTime))
}

func createSearchSolutionsRequest(columnIndex int, preprocessing *pipeline.PipelineDescription,
	datasetURI string, userAgent string, targetFeature *model.Variable, dataset string, metrics []string, task string, subTask string, maxTime int64) (*pipeline.SearchSolutionsRequest, error) {

	return &pipeline.SearchSolutionsRequest{
		Problem: &pipeline.ProblemDescription{
			Problem: &pipeline.Problem{
				TaskType:           compute.ConvertTaskTypeFromTA3ToTA2(task),
				TaskSubtype:        compute.ConvertTaskSubTypeFromTA3ToTA2(subTask),
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
		AllowedValueTypes: []pipeline.ValueType{
			pipeline.ValueType_DATASET_URI,
			pipeline.ValueType_CSV_URI,
			pipeline.ValueType_RAW,
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
func (s *SolutionRequest) createPreprocessingPipeline(featureVariables []*model.Variable) (*pipeline.PipelineDescription, error) {
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

	preprocessingPipeline, err := description.CreateUserDatasetPipeline(name, desc,
		&description.UserDatasetDescription{
			AllFeatures:      featureVariables,
			TargetFeature:    s.TargetFeature,
			SelectedFeatures: s.Filters.Variables,
			Filters:          s.Filters.Filters,
		}, augments)
	if err != nil {
		return nil, err
	}

	return preprocessingPipeline, nil
}

func (s *SolutionRequest) createProduceSolutionRequest(datasetURI string, fittedSolutionID string, outputs []string) *pipeline.ProduceSolutionRequest {
	return &pipeline.ProduceSolutionRequest{
		FittedSolutionId: fittedSolutionID,
		Inputs: []*pipeline.Value{
			{
				Value: &pipeline.Value_DatasetUri{
					DatasetUri: datasetURI,
				},
			},
		},
		ExposeOutputs: outputs,
		ExposeValueTypes: []pipeline.ValueType{
			pipeline.ValueType_CSV_URI,
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

func (s *SolutionRequest) persistSolution(statusChan chan SolutionStatus, solutionStorage api.SolutionStorage, searchID string, solutionID string, initialSearchSolutionID string) {
	err := solutionStorage.PersistSolution(searchID, solutionID, initialSearchSolutionID, time.Now())
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
	solutionID string, initialSearchSolutionID string, fittedSolutionID string, resultID string, resultURI string) {
	// persist the completed state
	err := solutionStorage.PersistSolutionState(initialSearchSolutionID, SolutionCompletedStatus, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}

	// persist the completed state
	err = solutionStorage.PersistSolutionState(solutionID, SolutionCompletedStatus, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}
	// persist result metadata
	err = solutionStorage.PersistSolutionResult(initialSearchSolutionID, fittedSolutionID, resultID, resultURI, SolutionCompletedStatus, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}
	// persist results
	err = dataStorage.PersistResult(dataset, model.NormalizeDatasetID(dataset), resultURI, s.TargetFeature.Name)
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}

	// notify client of update
	statusChan <- SolutionStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		ResultID:   resultID,
		Progress:   SolutionCompletedStatus,
		Timestamp:  time.Now(),
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
	datasetURITrain string, datasetURITest string, variables []*model.Variable) {

	// Need to create a new solution that has the explain output. This is the solution
	// that will be used throughout distil except for the export (which will use the original solution).
	// start a solution searchID
	explainDesc, err := s.createExplainPipeline(client, initialSearchSolutionID)
	if err != nil {
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}

	searchRequest.Template = explainDesc

	searchID, err := client.StartSearch(context.Background(), searchRequest)
	if err != nil {
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}
	wg := &sync.WaitGroup{}

	err = client.SearchSolutions(context.Background(), searchID, func(solution *pipeline.GetSearchSolutionsResultsResponse) {
		wg.Add(1)
		solutionID := solution.SolutionId

		// persist the solution info
		s.persistSolution(statusChan, solutionStorage, searchID, solutionID, initialSearchSolutionID)
		s.persistSolutionStatus(statusChan, solutionStorage, searchID, solutionID, SolutionPendingStatus)

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
						metric = compute.ConvertMetricsFromTA3ToTA2(s.Metrics)[0].GetMetric().String()
					} else {
						metric = score.Metric.Metric.String()
					}
					err := solutionStorage.PersistSolutionScore(solutionID, metric, score.Value.GetRaw().GetDouble())
					if err != nil {
						s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
						return
					}
				}
			}
		}

		// persist solution running status
		s.persistSolutionStatus(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, SolutionRunningStatus)
		s.persistSolutionStatus(statusChan, solutionStorage, searchID, solutionID, SolutionRunningStatus)

		// generate predictions
		produceSolutionRequest := s.createProduceSolutionRequest(datasetURITest, fittedSolutionID, []string{defaultExposedOutputKey, "outputs.1"})

		// generate predictions
		predictionResponses, err := client.GeneratePredictions(context.Background(), produceSolutionRequest)
		if err != nil {
			s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
			return
		}

		for _, response := range predictionResponses {

			if response.Progress.State != pipeline.ProgressState_COMPLETED {
				// only persist completed responses
				continue
			}

			output, ok := response.ExposedOutputs[defaultExposedOutputKey]
			if !ok {
				err := errors.Errorf("output is missing from response")
				s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
				return
			}

			csvURI, ok := output.Value.(*pipeline.Value_CsvUri)
			if !ok {
				err := errors.Errorf("output is not of correct format")
				s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
				return
			}

			// remove the protocol portion if it exists. The returned value is either a
			// csv file or a directory.
			resultURI := csvURI.CsvUri
			resultURI = strings.Replace(resultURI, "file://", "", 1)

			// get the result UUID. NOTE: Doing sha1 for now.
			hasher := sha1.New()
			_, err = hasher.Write([]byte(resultURI))
			if err != nil {
				s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
			}
			bs := hasher.Sum(nil)
			resultID := fmt.Sprintf("%x", bs)

			// explain the pipeline
			featureWeights, err := s.explainOutput(client, solutionID, resultURI, searchRequest, datasetURITest, resultURI, variables)
			if err != nil {
				log.Warnf("failed to fetch output explanantion - %s", err)
			}
			if featureWeights != nil {
				err = dataStorage.PersistSolutionFeatureWeight(dataset, model.NormalizeDatasetID(dataset), featureWeights.ResultURI, featureWeights.Weights)
				if err != nil {
					s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
					return
				}
			}

			// persist results
			s.persistSolutionResults(statusChan, client, solutionStorage, dataStorage, searchID,
				initialSearchID, dataset, solutionID, initialSearchSolutionID, fittedSolutionID, resultID, resultURI)
		}
		wg.Done()
	})
	if err != nil {
		s.persistSolutionError(statusChan, solutionStorage, initialSearchID, initialSearchSolutionID, err)
		return
	}

	wg.Wait()
}

func (s *SolutionRequest) dispatchRequest(client *compute.Client, solutionStorage api.SolutionStorage, dataStorage api.DataStorage,
	searchID string, dataset string, searchRequest *pipeline.SearchSolutionsRequest, datasetURITrain string, datasetURITest string, variables []*model.Variable) {

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
		s.dispatchSolution(c, client, solutionStorage, dataStorage, searchID, solution.SolutionId, dataset, searchRequest, datasetURITrain, datasetURITest, variables)
		// once done, mark as complete
		s.completeSolution()
	})

	// update request status
	if err != nil {
		s.persistRequestError(s.requestChannel, solutionStorage, searchID, dataset, err)
	} else {
		s.persistRequestStatus(s.requestChannel, solutionStorage, searchID, dataset, RequestCompletedStatus)
	}

	// wait until all are complete and the search has finished / timed out
	s.waitOnSolutions()

	// end search
	s.finished <- client.EndSearch(context.Background(), searchID)
}

// PersistAndDispatch persists the solution request and dispatches it.
func (s *SolutionRequest) PersistAndDispatch(client *compute.Client, solutionStorage api.SolutionStorage, metaStorage api.MetadataStorage, dataStorage api.DataStorage) error {

	// NOTE: D3M index field is needed in the persisted data.
	s.Filters.AddVariable(model.D3MIndexFieldName)

	// fetch the full set of variables associated with the dataset
	variables, err := metaStorage.FetchVariables(s.Dataset, true, true, true)
	if err != nil {
		return err
	}

	// remove generated features from our var list
	// TODO: imported datasets have d3m index as distil role = "index".
	//       need to figure out if that causes issues!!!
	dataVariables := []*model.Variable{}
	for _, variable := range variables {
		if isTA2Field(variable.DistilRole) {
			dataVariables = append(dataVariables, variable)
		}
	}
	targetVariable := s.TargetFeature

	// Timeseries are grouped entries and we want to use the Y Col from the group as the target
	// rather than the group itself, and the X col as the timestamp variable
	var groupingTargetVariable = targetVariable
	if targetVariable.Grouping != nil && model.IsTimeSeries(targetVariable.Grouping.Type) {
		// filter list needs to include all the individual grouping components
		for _, subID := range targetVariable.Grouping.SubIDs {
			s.Filters.AddVariable(subID)
		}
		s.Filters.AddVariable(targetVariable.Grouping.Properties.XCol)
		s.Filters.AddVariable(targetVariable.Grouping.Properties.YCol)

		// extract the time series value column
		targetVarName := targetVariable.Grouping.Properties.YCol
		targetVariable, err = metaStorage.FetchVariable(s.Dataset, targetVarName)
		if err != nil {
			return err
		}

		dataVariables = append(dataVariables, targetVariable)

	}

	// make sure that we include all non-generated variables in our persisted
	// dataset - the column removal preprocessing step will mark them for
	// removal by ta2
	allVarFilters := s.Filters.Clone()
	allVarFilters.Variables = []string{}
	var timeseriesField *model.Variable
	for _, variable := range dataVariables {
		// exclude cluster/feature generated columns
		allVarFilters.AddVariable(variable.Name)
		if variable.Name == s.TimestampField {
			timeseriesField = variable
		}
	}

	// fetch the queried dataset
	dataset, err := api.FetchDataset(s.Dataset, true, true, allVarFilters, metaStorage, dataStorage)
	if err != nil {
		return err
	}

	// fetch the input dataset (should only differ on augmented)
	datasetInput, err := metaStorage.FetchDataset(s.DatasetInput, true, true)
	if err != nil {
		return err
	}

	columnIndex := getColumnIndex(targetVariable, dataset.Filters.Variables)
	timeseriesColumnIndex := -1
	if timeseriesField != nil {
		// extract timeseries timestamp column from solution request
		timeseriesColumnIndex = getColumnIndex(timeseriesField, dataset.Filters.Variables)
	} else if groupingTargetVariable.Grouping != nil {
		// extract the timeseries timestamp column from a grouping
		timeseriesVarName := groupingTargetVariable.Grouping.Properties.XCol
		timeseriesVariable, err := metaStorage.FetchVariable(s.Dataset, timeseriesVarName)
		if err != nil {
			return err
		}
		timeseriesColumnIndex = getColumnIndex(timeseriesVariable, dataset.Filters.Variables)
	}

	// add dataset name to path
	datasetInputDir := env.ResolvePath(datasetInput.Source, datasetInput.Folder)

	// compute the task and subtask from the target and dataset
	task, err := ResolveTask(dataStorage, dataset.Metadata.StorageName, groupingTargetVariable)
	if err != nil {
		return err
	}
	s.Task = task.Task
	s.SubTask = task.SubTask

	// when dealing with categorical data we want to stratify
	stratify := false
	if targetVariable.Type == model.CategoricalType {
		stratify = true
	}

	// perist the datasets and get URI
	params := &persistedDataParams{
		DatasetName:          s.DatasetInput,
		SchemaFile:           compute.D3MDataSchema,
		SourceDataFolder:     datasetInputDir,
		TmpDataFolder:        datasetDir,
		TaskType:             s.Task,
		TimeseriesFieldIndex: timeseriesColumnIndex,
		TargetFieldIndex:     columnIndex,
		Stratify:             stratify,
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
		preprocessing, err = s.createPreprocessingPipeline(dataVariables)
		if err != nil {
			return err
		}
	}

	// create search solutions request
	searchRequest, err := createSearchSolutionsRequest(columnIndex, preprocessing, datasetPathTrain, client.UserAgent, targetVariable, s.DatasetInput, s.Metrics, s.Task, s.SubTask, int64(s.MaxTime))

	if err != nil {
		return err
	}

	// start a solution searchID
	requestID, err := client.StartSearch(context.Background(), searchRequest)
	if err != nil {
		return err
	}

	// persist the request
	err = s.persistRequestStatus(s.requestChannel, solutionStorage, requestID, dataset.Metadata.ID, RequestPendingStatus)
	if err != nil {
		return err
	}

	// store the request features
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
	go s.dispatchRequest(client, solutionStorage, dataStorage, requestID, dataset.Metadata.ID, searchRequest, datasetPathTrain, datasetPathTest, dataVariables)

	return nil
}

// CreateSearchSolutionRequest creates a search solution request, including
// the pipeline steps required to process the data.
func CreateSearchSolutionRequest(request *SolutionRequestDiscovery, skipPreprocessing bool) (*pipeline.SearchSolutionsRequest, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create uuid")
	}

	name := fmt.Sprintf("preprocessing-%s-%s", request.Dataset, uuid.String())
	desc := fmt.Sprintf("Preprocessing pipeline capturing user feature selection and type information. Dataset: `%s` ID: `%s`", request.Dataset, uuid.String())

	var preprocessingPipeline *pipeline.PipelineDescription
	if !skipPreprocessing {
		var augments []*description.UserDatasetAugmentation
		if request.DatasetAugmentations != nil {
			augments = make([]*description.UserDatasetAugmentation, len(request.DatasetAugmentations))
			for i, da := range request.DatasetAugmentations {
				augments[i] = &description.UserDatasetAugmentation{
					SearchResult:  da.SearchResult,
					SystemID:      da.Provenance,
					BaseDatasetID: request.Dataset,
				}
			}
		}

		preprocessingPipeline, err = description.CreateUserDatasetPipeline(name, desc,
			&description.UserDatasetDescription{
				AllFeatures:      request.AllFeatures,
				TargetFeature:    request.TargetFeature,
				SelectedFeatures: request.SelectedFeatures,
				Filters:          nil,
			}, augments)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create preprocessing pipeline")
		}
	}

	targetVariable := request.TargetFeature
	columnIndex := getColumnIndex(targetVariable, request.SelectedFeatures)
	task := DefaultTaskType(targetVariable.Type, "")
	taskSubType := DefaultTaskSubType(task)
	metrics := DefaultMetrics(task)

	// create search solutions request
	searchRequest, err := createSearchSolutionsRequest(columnIndex, preprocessingPipeline, request.SourceURI,
		request.UserAgent, request.TargetFeature, request.DatasetInput, metrics, task, taskSubType, 600)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create search solution request")
	}

	return searchRequest, nil
}

func getColumnIndex(variable *model.Variable, selectedVariables []string) int {
	colIndex := 0
	for i := 0; i < len(selectedVariables); i++ {
		if selectedVariables[i] == variable.Name {
			break
		}
		colIndex = colIndex + 1
	}

	return colIndex
}

func isTA2Field(distilRole string) bool {
	return ta2RoleMap[distilRole]
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
