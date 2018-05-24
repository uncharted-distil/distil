package compute

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	protobuf "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/pipeline"

	"github.com/unchartedsoftware/distil/api/model"
	log "github.com/unchartedsoftware/plog"
)

const (
	defaultResourceID       = "0"
	defaultExposedOutputKey = "outputs.0"
	datasetDir              = "datasets"
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

// StopSolutionSearchRequest represents a request to stop any pending siolution searches.
type StopSolutionSearchRequest struct {
	RequestID string `json:"requestId"`
}

// NewStopSolutionSearchRequest instantiates a new StopSolutionSearchRequest.
func NewStopSolutionSearchRequest(data []byte) (*StopSolutionSearchRequest, error) {
	req := &StopSolutionSearchRequest{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Dispatch dispatches the stop search request.
func (s *StopSolutionSearchRequest) Dispatch(client *Client) error {
	return client.StopSearch(context.Background(), s.RequestID)
}

func newStatusChannel() chan SolutionStatus {
	// NOTE: WE BUFFER THE CHANNEL TO A SIZE OF 1 HERE SO THAT THE INITIAL
	// PERSIST DOES NOT DEADLOCK
	return make(chan SolutionStatus, 1)
}

// SolutionRequest represents a solution search request.
type SolutionRequest struct {
	Dataset          string              `json:"dataset"`
	Index            string              `json:"index"`
	TargetFeature    string              `json:"target"`
	Task             string              `json:"task"`
	MaxSolutions     int32               `json:"maxSolutions"`
	Filters          *model.FilterParams `json:"filters"`
	Metrics          []string            `json:"metrics"`
	mu               *sync.Mutex
	wg               *sync.WaitGroup
	requestChannel   chan SolutionStatus
	solutionChannels []chan SolutionStatus
	listener         SolutionStatusListener
	finished         chan error
}

// NewSolutionRequest instantiates a new SolutionRequest.
func NewSolutionRequest(data []byte) (*SolutionRequest, error) {
	req := &SolutionRequest{
		mu:             &sync.Mutex{},
		wg:             &sync.WaitGroup{},
		finished:       make(chan error),
		requestChannel: newStatusChannel(),
	}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}
	return req, nil
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wg.Add(1)
	s.solutionChannels = append(s.solutionChannels, c)
	go s.listenOnStatusChannel(c)
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

func (s *SolutionRequest) createSearchSolutionsRequest(targetIndex int, datasetURI string, userAgent string) (*pipeline.SearchSolutionsRequest, error) {
	// Grab the embedded ta3ta2 API version
	apiVersion, err := getAPIVersion()
	if err != nil {
		log.Warnf("Failed to extract API version")
		apiVersion = "unknown"
	}

	return &pipeline.SearchSolutionsRequest{
		Problem: &pipeline.ProblemDescription{
			Problem: &pipeline.Problem{
				TaskType:           convertTaskTypeFromTA3ToTA2(s.Task),
				PerformanceMetrics: convertMetricsFromTA3ToTA2(s.Metrics),
			},
			Inputs: []*pipeline.ProblemInput{
				{
					DatasetId: convertDatasetTA3ToTA2(s.Dataset),
					Targets:   convertTargetFeaturesTA3ToTA2(s.TargetFeature, targetIndex),
				},
			},
		},

		UserAgent: userAgent,
		Version:   apiVersion,

		// we accept dataset and csv uris as return types
		AllowedValueTypes: []pipeline.ValueType{
			pipeline.ValueType_DATASET_URI,
		},

		// URI of the input dataset
		Inputs: []*pipeline.Value{
			{
				Value: &pipeline.Value_DatasetUri{
					DatasetUri: datasetURI,
				},
			},
		},
	}, nil
}

func (s *SolutionRequest) createProduceSolutionRequest(datasetURI string, solutionID string) *pipeline.ProduceSolutionRequest {
	return &pipeline.ProduceSolutionRequest{
		SolutionId: solutionID,
		Inputs: []*pipeline.Value{
			{
				Value: &pipeline.Value_DatasetUri{
					DatasetUri: datasetURI,
				},
			},
		},
	}
}

func (s *SolutionRequest) persistSolutionError(statusChan chan SolutionStatus, solutionStorage model.SolutionStorage, searchID string, solutionID string, err error) {
	// persist the updated state
	// NOTE: ignoring error
	solutionStorage.PersistSolution(searchID, solutionID, SolutionErroredStatus, time.Now())
	// notify of error
	statusChan <- SolutionStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		Progress:   SolutionErroredStatus,
		Error:      err,
		Timestamp:  time.Now(),
	}
}

func (s *SolutionRequest) persistSolutionStatus(statusChan chan SolutionStatus, solutionStorage model.SolutionStorage, searchID string, solutionID string, status string) {
	// persist the updated state
	err := solutionStorage.PersistSolution(searchID, solutionID, status, time.Now())
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

func (s *SolutionRequest) persistRequestError(statusChan chan SolutionStatus, solutionStorage model.SolutionStorage, searchID string, dataset string, err error) {
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

func (s *SolutionRequest) persistRequestStatus(statusChan chan SolutionStatus, solutionStorage model.SolutionStorage, searchID string, dataset string, status string) error {
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

func (s *SolutionRequest) persistSolutionResults(statusChan chan SolutionStatus, client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, dataset string, solutionID string, resultID string, resultURI string) {
	// persist the completed state
	err := solutionStorage.PersistSolution(searchID, solutionID, SolutionCompletedStatus, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, searchID, solutionID, err)
		return
	}
	// persist result metadata
	err = solutionStorage.PersistSolutionResult(solutionID, resultID, resultURI, SolutionCompletedStatus, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, searchID, solutionID, err)
		return
	}
	// persist results
	err = dataStorage.PersistResult(dataset, resultURI)
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, solutionStorage, searchID, solutionID, err)
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
}

func (s *SolutionRequest) dispatchSolution(statusChan chan SolutionStatus, client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, solutionID string, dataset string, datasetURITrain string, datasetURITest string) {

	// score solution
	solutionScoreResponses, err := client.GenerateSolutionScores(context.Background(), solutionID)
	if err != nil {
		s.persistSolutionError(statusChan, solutionStorage, searchID, solutionID, err)
		return
	}

	// persist the scores
	for _, response := range solutionScoreResponses {
		for _, score := range response.Scores {
			err := solutionStorage.PersistSolutionScore(solutionID, score.Metric.Metric.String(), score.Value.GetDouble())
			if err != nil {
				s.persistSolutionError(statusChan, solutionStorage, searchID, solutionID, err)
				return
			}
		}
	}

	// fit solution
	_, err = client.GenerateSolutionFit(context.Background(), solutionID, datasetURITrain)
	if err != nil {
		s.persistSolutionError(statusChan, solutionStorage, searchID, solutionID, err)
		return
	}

	// persist solution running status
	s.persistSolutionStatus(statusChan, solutionStorage, searchID, solutionID, SolutionRunningStatus)

	// generate predictions
	produceSolutionRequest := s.createProduceSolutionRequest(datasetURITest, solutionID)

	// generate predictions
	predictionResponses, err := client.GeneratePredictions(context.Background(), produceSolutionRequest)
	if err != nil {
		s.persistSolutionError(statusChan, solutionStorage, searchID, solutionID, err)
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
			s.persistSolutionError(statusChan, solutionStorage, searchID, solutionID, err)
			return
		}

		datasetURI, ok := output.Value.(*pipeline.Value_DatasetUri)
		if !ok {
			err := errors.Errorf("output is not of correct format")
			s.persistSolutionError(statusChan, solutionStorage, searchID, solutionID, err)
			return
		}

		// remove the protocol portion if it exists. The returned value is either a
		// csv file or a directory.
		resultURI := datasetURI.DatasetUri
		resultURI = strings.Replace(resultURI, "file://", "", 1)
		if !strings.HasSuffix(resultURI, ".csv") {
			resultURI = path.Join(resultURI, D3MLearningData)
		}

		// get the result UUID. NOTE: Doing sha1 for now.
		hasher := sha1.New()
		hasher.Write([]byte(resultURI))
		bs := hasher.Sum(nil)
		resultID := fmt.Sprintf("%x", bs)

		// persist results
		s.persistSolutionResults(statusChan, client, solutionStorage, dataStorage, searchID, dataset, solutionID, resultID, resultURI)
	}
}

func (s *SolutionRequest) dispatchRequest(client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, dataset string, datasetURITrain string, datasetURITest string) {

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
		// persist the solution
		s.persistSolutionStatus(c, solutionStorage, searchID, solution.SolutionId, SolutionPendingStatus)
		// add the solution to the request
		s.addSolution(c)
		// dispatch it
		s.dispatchSolution(c, client, solutionStorage, dataStorage, searchID, solution.SolutionId, dataset, datasetURITrain, datasetURITest)
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

func splitTrainTest(dataset *model.QueriedDataset) (*model.QueriedDataset, *model.QueriedDataset, error) {
	trainDataset := &model.QueriedDataset{
		Metadata: dataset.Metadata,
		Filters:  dataset.Filters,
		IsTrain:  true,
		Data: &model.FilteredData{
			Name:    dataset.Data.Name,
			NumRows: dataset.Data.NumRows,
			Columns: dataset.Data.Columns,
			Types:   dataset.Data.Types,
			Values:  make([][]interface{}, 0),
		},
	}
	testDataset := &model.QueriedDataset{
		Metadata: dataset.Metadata,
		Filters:  dataset.Filters,
		IsTrain:  true,
		Data: &model.FilteredData{
			Name:    dataset.Data.Name,
			NumRows: dataset.Data.NumRows,
			Columns: dataset.Data.Columns,
			Types:   dataset.Data.Types,
			Values:  make([][]interface{}, 0),
		},
	}

	// randomly split the dataset between train and test
	for _, r := range dataset.Data.Values {
		if rand.Float64() < trainTestSplitThreshold {
			trainDataset.Data.Values = append(trainDataset.Data.Values, r)
		} else {
			testDataset.Data.Values = append(testDataset.Data.Values, r)
		}
	}

	return trainDataset, testDataset, nil
}

// PersistAndDispatch persists the solution request and dispatches it.
func (s *SolutionRequest) PersistAndDispatch(client *Client, solutionStorage model.SolutionStorage, metaStorage model.MetadataStorage, dataStorage model.DataStorage) error {

	// NOTE: D3M index field is needed in the persisted data.
	s.Filters.Variables = append(s.Filters.Variables, model.D3MIndexFieldName)

	// fetch the queried dataset
	dataset, err := model.FetchDataset(s.Dataset, s.Index, true, s.Filters, metaStorage, dataStorage)
	if err != nil {
		return err
	}

	// split the train & test data into separate datasets to be submitted to TA2
	trainDataset, testDataset, err := splitTrainTest(dataset)

	// perist the datasets and get URI
	datasetPathTrain, targetIndex, err := PersistFilteredData(datasetDir, s.TargetFeature, trainDataset)
	if err != nil {
		return err
	}
	datasetPathTest, _, err := PersistFilteredData(datasetDir, s.TargetFeature, testDataset)
	if err != nil {
		return err
	}
	// make sure the path is absolute and contains the URI prefix
	datasetPathTrain, err = filepath.Abs(datasetPathTrain)
	if err != nil {
		return err
	}
	datasetPathTrain = fmt.Sprintf("%s", filepath.Join(datasetPathTrain, D3MDataSchema))
	datasetPathTest, err = filepath.Abs(datasetPathTest)
	if err != nil {
		return err
	}
	datasetPathTest = fmt.Sprintf("%s", filepath.Join(datasetPathTest, D3MDataSchema))

	// create search solutions request
	searchRequest, err := s.createSearchSolutionsRequest(targetIndex, datasetPathTrain, client.UserAgent)
	if err != nil {
		return err
	}

	// start a solution searchID
	requestID, err := client.StartSearch(context.Background(), searchRequest)
	if err != nil {
		return err
	}

	// persist the request
	err = s.persistRequestStatus(s.requestChannel, solutionStorage, requestID, dataset.Metadata.Name, RequestPendingStatus)
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

		if v == s.TargetFeature {
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
	go s.dispatchRequest(client, solutionStorage, dataStorage, requestID, dataset.Metadata.Name, datasetPathTrain, datasetPathTest)

	return nil
}

func convertMetricsFromTA3ToTA2(metrics []string) []*pipeline.ProblemPerformanceMetric {
	var res []*pipeline.ProblemPerformanceMetric
	for _, metric := range metrics {
		res = append(res, &pipeline.ProblemPerformanceMetric{
			Metric: pipeline.PerformanceMetric(pipeline.PerformanceMetric_value[strings.ToUpper(metric)]),
		})
	}
	return res
}

func convertTaskTypeFromTA3ToTA2(taskType string) pipeline.TaskType {
	return pipeline.TaskType(pipeline.TaskType_value[strings.ToUpper(taskType)])
}

func convertTargetFeaturesTA3ToTA2(target string, targetIndex int) []*pipeline.ProblemTarget {
	return []*pipeline.ProblemTarget{
		{
			ColumnName:  target,
			ResourceId:  defaultResourceID,
			TargetIndex: int32(targetIndex),
			ColumnIndex: int32(targetIndex), // TODO: is this correct?
		},
	}
}

func convertDatasetTA3ToTA2(dataset string) string {
	return dataset
}

// getApiVersion retrieves the ta3-ta2 API version embedded in the pipeline_service.proto file
func getAPIVersion() (string, error) {
	// Get the raw file descriptor bytes
	fileDesc := proto.FileDescriptor(pipeline.E_ProtocolVersion.Filename)
	if fileDesc == nil {
		return "", fmt.Errorf("failed to find file descriptor for %v", pipeline.E_ProtocolVersion.Filename)
	}

	// Open a gzip reader and decompress
	r, err := gzip.NewReader(bytes.NewReader(fileDesc))
	if err != nil {
		return "", fmt.Errorf("failed to open gzip reader: %v", err)
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to decompress descriptor: %v", err)
	}

	// Unmarshall the bytes from the proto format
	fd := &protobuf.FileDescriptorProto{}
	if err := proto.Unmarshal(b, fd); err != nil {
		return "", fmt.Errorf("malformed FileDescriptorProto: %v", err)
	}

	// Fetch the extension from the FileDescriptorOptions message
	ex, err := proto.GetExtension(fd.GetOptions(), pipeline.E_ProtocolVersion)
	if err != nil {
		return "", fmt.Errorf("failed to fetch extension: %v", err)
	}

	return *ex.(*string), nil
}
