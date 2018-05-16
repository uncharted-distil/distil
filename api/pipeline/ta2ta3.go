package pipeline

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha1"
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

// CreateMessage represents a create model message.
type CreateMessage struct {
	Dataset       string              `json:"dataset"`
	Index         string              `json:"index"`
	TargetFeature string              `json:"target"`
	Task          string              `json:"task"`
	MaxSolutions  int32               `json:"maxSolutions"`
	Filters       *model.FilterParams `json:"filters"`
	Metrics       []string            `json:"metrics"`
}

// CreateStatus represents a solution status.
type CreateStatus struct {
	Progress   string    `json:"progress"`
	RequestID  string    `json:"requestId"`
	SolutionID string    `json:"solutionId"`
	ResultID   string    `json:"resultId"`
	Error      error     `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
}

func (m *CreateMessage) createSearchSolutionsRequest(targetIndex int, datasetURI string, userAgent string) (*SearchSolutionsRequest, error) {
	// Grab the embedded ta3ta2 API version
	apiVersion, err := getAPIVersion()
	if err != nil {
		log.Warnf("Failed to extract API version")
		apiVersion = "unknown"
	}

	return &SearchSolutionsRequest{
		Problem: &ProblemDescription{
			Problem: &Problem{
				TaskType:           convertTaskTypeFromTA3ToTA2(m.Task),
				PerformanceMetrics: convertMetricsFromTA3ToTA2(m.Metrics),
			},
			Inputs: []*ProblemInput{
				{
					DatasetId: convertDatasetTA3ToTA2(m.Dataset),
					Targets:   convertTargetFeaturesTA3ToTA2(m.TargetFeature, targetIndex),
				},
			},
		},

		UserAgent: userAgent,
		Version:   apiVersion,

		// we accept dataset and csv uris as return types
		AllowedValueTypes: []ValueType{
			ValueType_DATASET_URI,
		},

		// URI of the input dataset
		Inputs: []*Value{
			{
				Value: &Value_DatasetUri{
					DatasetUri: m.Dataset,
				},
			},
		},
	}, nil
}

func (m *CreateMessage) createProduceSolutionRequest(datasetURI string, solutionID string) *ProduceSolutionRequest {
	return &ProduceSolutionRequest{
		SolutionId: solutionID,
		Inputs: []*Value{
			{
				Value: &Value_DatasetUri{
					DatasetUri: datasetURI,
				},
			},
		},
	}
}

func (m *CreateMessage) persistSolutionError(statusChan chan CreateStatus, client *Client, solutionStorage model.SolutionStorage, searchID string, solutionID string, err error) {
	// persist the updated state
	// NOTE: ignoring error
	solutionStorage.PersistSolution(searchID, solutionID, ErroredStatus, time.Now())
	// notify of error
	statusChan <- CreateStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		Progress:   ErroredStatus,
		Error:      err,
		Timestamp:  time.Now(),
	}
}

func (m *CreateMessage) persistSolutionStatus(statusChan chan CreateStatus, client *Client, solutionStorage model.SolutionStorage, searchID string, solutionID string, status string) {
	// persist the updated state
	err := solutionStorage.PersistSolution(searchID, solutionID, status, time.Now())
	if err != nil {
		// notify of error
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// notify client of update
	statusChan <- CreateStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		Progress:   status,
		Timestamp:  time.Now(),
	}
}

func (m *CreateMessage) persistSolutionResults(statusChan chan CreateStatus, client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, dataset string, solutionID string, resultID string, resultURI string) {
	// persist the completed state
	err := solutionStorage.PersistSolution(searchID, solutionID, CompletedStatus, time.Now())
	if err != nil {
		// notify of error
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// persist result metadata
	err = solutionStorage.PersistSolutionResult(solutionID, resultID, resultURI, CompletedStatus, time.Now())
	if err != nil {
		// notify of error
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// persist results
	err = dataStorage.PersistResult(dataset, resultURI)
	if err != nil {
		// notify of error
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}
	// notify client of update
	statusChan <- CreateStatus{
		RequestID:  searchID,
		SolutionID: solutionID,
		ResultID:   resultID,
		Progress:   CompletedStatus,
		Timestamp:  time.Now(),
	}
}

func (m *CreateMessage) dispatchSolution(statusChan chan CreateStatus, client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, solutionID string, dataset string, datasetURITrain string, datasetURITest string) {

	// score solution
	solutionScoreResponses, err := client.GenerateSolutionScores(context.Background(), solutionID)
	if err != nil {
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}

	// persist the scores
	for _, response := range solutionScoreResponses {
		for _, score := range response.Scores {
			err := solutionStorage.PersistSolutionScore(solutionID, score.Metric.Metric.String(), score.Value.GetDouble())
			if err != nil {
				m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
				return
			}
		}
	}

	// fit solution
	_, err = client.GenerateSolutionFit(context.Background(), solutionID, datasetURITrain)
	if err != nil {
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}

	// persist solution running status
	m.persistSolutionStatus(statusChan, client, solutionStorage, searchID, solutionID, RunningStatus)

	// generate predictions
	produceSolutionRequest := m.createProduceSolutionRequest(datasetURITest, solutionID)

	// generate predictions
	predictionResponses, err := client.GeneratePredictions(context.Background(), produceSolutionRequest)
	if err != nil {
		m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, err)
		return
	}

	for _, response := range predictionResponses {

		if response.Progress.State != ProgressState_COMPLETED {
			// only persist completed responses
			continue
		}

		output, ok := response.ExposedOutputs[defaultExposedOutputKey]
		if !ok {
			m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, errors.Errorf("output is missing from response"))
			return
		}

		datasetURI, ok := output.Value.(*Value_DatasetUri)
		if !ok {
			m.persistSolutionError(statusChan, client, solutionStorage, searchID, solutionID, errors.Errorf("output is not of correct format"))
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
		m.persistSolutionResults(statusChan, client, solutionStorage, dataStorage, searchID, dataset, solutionID, resultID, resultURI)
	}
}

func (m *CreateMessage) createStatusChannels(client *Client, solutions []*GetSearchSolutionsResultsResponse, solutionStorage model.SolutionStorage, searchID string) []chan CreateStatus {

	// create channels

	// NOTE: WE BUFFER THE CHANNELS TO A SIZE OF 1 HERE SO WE CAN PERSIST BELOW
	// WITHOUT DEADLOCKING.
	var statusChannels []chan CreateStatus
	for range solutions {
		statusChannels = append(statusChannels, make(chan CreateStatus, 1))
	}

	// persist all solutions as pending

	// NOTE: we persist the solutions here so that they exist in the DB when the
	// method returns.
	// NOTE: THE CHANNELS MUST BE BUFFERED TO A SIZE OF 1 OR ELSE THIS WILL DEADLOCK.
	for i, solution := range solutions {
		m.persistSolutionStatus(statusChannels[i], client, solutionStorage, searchID, solution.SolutionId, PendingStatus)
	}

	return statusChannels
}

// DispatchSolutions dispatches all solution requests
func (m *CreateMessage) DispatchSolutions(client *Client, solutionStorage model.SolutionStorage, dataStorage model.DataStorage, searchID string, dataset string, datasetURITrain string, datasetURITest string) ([]chan CreateStatus, error) {

	solutions, err := client.SearchSolutions(context.Background(), searchID)
	if err != nil {
		return nil, err
	}

	// create status channels and persist solutions
	statusChannels := m.createStatusChannels(client, solutions, solutionStorage, searchID)

	// dispatch all solutions
	go func() {

		wg := &sync.WaitGroup{}

		// dispatch individual solutions
		for i, solution := range solutions {

			// increment waitgroup
			wg.Add(1)

			go func(statusChan chan CreateStatus, solutionID string) {
				m.dispatchSolution(statusChan, client, solutionStorage, dataStorage, searchID, solutionID, dataset, datasetURITrain, datasetURITest)
				wg.Done()
			}(statusChannels[i], solution.SolutionId)

		}

		// wait until all are complete
		wg.Wait()

		// end search
		client.EndSearch(context.Background(), searchID)
	}()

	return statusChannels, nil
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
func (m *CreateMessage) PersistAndDispatch(client *Client, solutionStorage model.SolutionStorage, metaStorage model.MetadataStorage, dataStorage model.DataStorage) ([]chan CreateStatus, error) {

	// NOTE: D3M index field is needed in the persisted data.
	m.Filters.Variables = append(m.Filters.Variables, model.D3MIndexFieldName)

	// fetch the queried dataset
	dataset, err := model.FetchDataset(m.Dataset, m.Index, true, m.Filters, metaStorage, dataStorage)
	if err != nil {
		return nil, err
	}

	// split the train & test data into separate datasets to be submitted to TA2
	trainDataset, testDataset, err := splitTrainTest(dataset)

	// perist the datasets and get URI
	datasetPathTrain, targetIndex, err := PersistFilteredData(datasetDir, m.TargetFeature, trainDataset)
	if err != nil {
		return nil, err
	}
	datasetPathTest, _, err := PersistFilteredData(datasetDir, m.TargetFeature, testDataset)
	if err != nil {
		return nil, err
	}
	// make sure the path is absolute and contains the URI prefix
	datasetPathTrain, err = filepath.Abs(datasetPathTrain)
	if err != nil {
		return nil, err
	}
	datasetPathTrain = fmt.Sprintf("%s", filepath.Join(datasetPathTrain, D3MDataSchema))
	datasetPathTest, err = filepath.Abs(datasetPathTest)
	if err != nil {
		return nil, err
	}
	datasetPathTest = fmt.Sprintf("%s", filepath.Join(datasetPathTest, D3MDataSchema))

	// create search solutions request
	searchRequest, err := m.createSearchSolutionsRequest(targetIndex, datasetPathTrain, client.UserAgent)
	if err != nil {
		return nil, err
	}

	// start a solution searchID
	requestID, err := client.StartSearch(context.Background(), searchRequest)
	if err != nil {
		return nil, err
	}

	// persist the request
	err = solutionStorage.PersistRequest(requestID, dataset.Metadata.Name, PendingStatus, time.Now())
	if err != nil {
		return nil, err
	}

	// store the request features
	for _, v := range m.Filters.Variables {
		var typ string
		// ignore the index field
		if v == model.D3MIndexFieldName {
			continue
		}

		if v == m.TargetFeature {
			// store target feature
			typ = model.FeatureTypeTarget
		} else {
			// store training feature
			typ = model.FeatureTypeTrain
		}
		err = solutionStorage.PersistRequestFeature(requestID, v, typ)
		if err != nil {
			return nil, err
		}
	}

	// store request filters
	err = solutionStorage.PersistRequestFilters(requestID, m.Filters)
	if err != nil {
		return nil, err
	}

	// dispatch solutions
	statusChannels, err := m.DispatchSolutions(client, solutionStorage, dataStorage, requestID, dataset.Metadata.Name, datasetPathTrain, datasetPathTest)
	if err != nil {
		return nil, err
	}

	return statusChannels, nil
}

func convertMetricsFromTA3ToTA2(metrics []string) []*ProblemPerformanceMetric {
	var res []*ProblemPerformanceMetric
	for _, metric := range metrics {
		res = append(res, &ProblemPerformanceMetric{
			Metric: PerformanceMetric(PerformanceMetric_value[strings.ToUpper(metric)]),
		})
	}
	return res
}

func convertTaskTypeFromTA3ToTA2(taskType string) TaskType {
	return TaskType(TaskType_value[strings.ToUpper(taskType)])
}

func convertTargetFeaturesTA3ToTA2(target string, targetIndex int) []*ProblemTarget {
	return []*ProblemTarget{
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
	fileDesc := proto.FileDescriptor(E_ProtocolVersion.Filename)
	if fileDesc == nil {
		return "", fmt.Errorf("failed to find file descriptor for %v", E_ProtocolVersion.Filename)
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
	ex, err := proto.GetExtension(fd.GetOptions(), E_ProtocolVersion)
	if err != nil {
		return "", fmt.Errorf("failed to fetch extension: %v", err)
	}

	return *ex.(*string), nil
}
