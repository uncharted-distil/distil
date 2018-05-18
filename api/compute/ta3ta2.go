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
	// PendingStatus represents that the solution request has been acknoledged by not yet sent to the API
	PendingStatus = "PENDING"
	// RunningStatus represents that the solution request has been sent to the API.
	RunningStatus = "RUNNING"
	// ErroredStatus represents that the solution request has terminated with an error.
	ErroredStatus = "ERRORED"
	// CompletedStatus represents that the solution request has completed successfully.
	CompletedStatus = "COMPLETED"
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

// SolutionRequest represents a solution search request.
type SolutionRequest struct {
	Dataset        string              `json:"dataset"`
	Index          string              `json:"index"`
	TargetFeature  string              `json:"target"`
	Task           string              `json:"task"`
	MaxSolutions   int32               `json:"maxSolutions"`
	Filters        *model.FilterParams `json:"filters"`
	Metrics        []string            `json:"metrics"`
	mu             *sync.Mutex
	wg             *sync.WaitGroup
	statusChannels []chan SolutionStatus
	listener       SolutionStatusListener
}

// NewSolutionRequest instantiates a new SolutionRequest.
func NewSolutionRequest(data []byte) (*SolutionRequest, error) {
	req := &SolutionRequest{
		mu: &sync.Mutex{},
		wg: &sync.WaitGroup{},
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
	s.statusChannels = append(s.statusChannels, c)
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
func (s *SolutionRequest) Listen(listener SolutionStatusListener) {
	s.listener = listener
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.statusChannels {
		go s.listenOnStatusChannel(c)
	}
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

func (s *SolutionRequest) persistSolutionError(statusChan chan SolutionStatus, client *Client, solutionStorage model.SolutionStorage, searchID string, solutionID string, err error) {
	// persist the updated state
	// NOTE: ignoring error
	solutionStorage.PersistSolution(searchID, solutionID, ErroredStatus, time.Now())
	// notify of error
	statusChan <- SolutionStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		Progress:   ErroredStatus,
		Error:      err,
		Timestamp:  time.Now(),
	}
}

func (s *SolutionRequest) persistSolutionStatus(statusChan chan SolutionStatus, client *Client, solutionStorage model.SolutionStorage, searchID string, solutionID string, status string) {
	// persist the updated state
	err := solutionStorage.PersistSolution(searchID, solutionID, status, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// notify client of update
	statusChan <- SolutionStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		Progress:   status,
		Timestamp:  time.Now(),
	}
}

func (s *SolutionRequest) persistSolutionResults(statusChan chan SolutionStatus, client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, dataset string, solutionID string, resultID string, resultURI string) {
	// persist the completed state
	err := solutionStorage.PersistSolution(searchID, solutionID, CompletedStatus, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// persist result metadata
	err = solutionStorage.PersistSolutionResult(solutionID, resultID, resultURI, CompletedStatus, time.Now())
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// persist results
	err = dataStorage.PersistResult(dataset, resultURI)
	if err != nil {
		// notify of error
		s.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// notify client of update
	statusChan <- SolutionStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		ResultID:   resultID,
		Progress:   CompletedStatus,
		Timestamp:  time.Now(),
	}
}

func (s *SolutionRequest) dispatchSolution(statusChan chan SolutionStatus, client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, solutionID string, dataset string, datasetURITrain string, datasetURITest string) {

	// score solution
	solutionScoreResponses, err := client.GenerateSolutionScores(context.Background(), solutionID)
	if err != nil {
		s.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}

	// persist the scores
	for _, response := range solutionScoreResponses {
		for _, score := range response.Scores {
			err := solutionStorage.PersistSolutionScore(solutionID, score.Metric.Metric.String(), score.Value.GetDouble())
			if err != nil {
				s.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
				return
			}
		}
	}

	// fit solution
	_, err = client.GenerateSolutionFit(context.Background(), solutionID, datasetURITrain)
	if err != nil {
		s.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}

	// persist solution running status
	s.persistSolutionStatus(statusChan, client, solutionStorage, searchID, solutionID, RunningStatus)

	// generate predictions
	produceSolutionRequest := s.createProduceSolutionRequest(datasetURITest, solutionID)

	// generate predictions
	predictionResponses, err := client.GeneratePredictions(context.Background(), produceSolutionRequest)
	if err != nil {
		s.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}

	for _, response := range predictionResponses {

		if response.Progress.State != pipeline.ProgressState_COMPLETED {
			// only persist completed responses
			continue
		}

		output, ok := response.ExposedOutputs[defaultExposedOutputKey]
		if !ok {
			s.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, errors.Errorf("output is missing from response"))
			return
		}

		datasetURI, ok := output.Value.(*pipeline.Value_DatasetUri)
		if !ok {
			s.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, errors.Errorf("output is not of correct format"))
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

func (s *SolutionRequest) createStatusChannel(client *Client, solution *pipeline.GetSearchSolutionsResultsResponse, solutionStorage model.SolutionStorage, searchID string) chan SolutionStatus {
	// NOTE: WE BUFFER THE CHANNEL TO A SIZE OF 1 HERE SO WE CAN PERSIST BELOW
	// WITHOUT DEADLOCKING.
	statusChannel := make(chan SolutionStatus, 1)

	// persist all solutions as pending

	// NOTE: we persist the solution here so that it exists in the DB when the
	// method returns.
	// NOTE: THE CHANNELS MUST BE BUFFERED TO A SIZE OF 1 OR ELSE THIS WILL DEADLOCK.
	s.persistSolutionStatus(statusChannel, client, solutionStorage, searchID, solution.SolutionId, PendingStatus)

	return statusChannel
}

func (s *SolutionRequest) dispatchRequest(client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, dataset string, datasetURITrain string, datasetURITest string) error {

	fmt.Println("SEARCHING SOLUTIONS")

	// search for solutions, this wont return until the search finishes or it times out
	err := client.SearchSolutions(context.Background(), searchID, func(solution *pipeline.GetSearchSolutionsResultsResponse) {

		// create a new status channel for the solution
		c := s.createStatusChannel(client, solution, solutionStorage, searchID)
		// add the solution to the request
		s.addSolution(c)

		fmt.Println("CREATED", solution.SolutionId)

		// dispatch it
		s.dispatchSolution(c, client, solutionStorage, dataStorage, searchID, solution.SolutionId, dataset, datasetURITrain, datasetURITest)

		fmt.Println("DISPATCHED", solution.SolutionId)

		// once done, mark as complete
		s.completeSolution()

		fmt.Println("COMPLETED", solution.SolutionId)
	})
	if err != nil {
		return err
	}

	fmt.Println("SEARCH FINISHED")

	// wait until all are complete and the search has finished / timed out
	s.waitOnSolutions()

	fmt.Println("FINISHED WAITING ON SOLUTIONS")

	// end search
	return client.EndSearch(context.Background(), searchID)
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
	err = solutionStorage.PersistRequest(requestID, dataset.Metadata.Name, PendingStatus, time.Now())
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
