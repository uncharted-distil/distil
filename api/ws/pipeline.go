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

package ws

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/compute"
	"github.com/uncharted-distil/distil/api/dataset"
	"github.com/uncharted-distil/distil/api/env"
	apiModel "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	jutil "github.com/uncharted-distil/distil/api/util/json"
)

const (
	createSolutions = "CREATE_SOLUTIONS"
	stopSolutions   = "STOP_SOLUTIONS"
	predict         = "PREDICT"
	query           = "QUERY"
)

var (
	// shared map of running requests - accessed by message status handlers that are
	// run under separate go routines so needs to be locked
	requestMap = struct {
		sync.RWMutex
		m map[string]*api.SolutionRequest
	}{
		m: map[string]*api.SolutionRequest{},
	}
)

// SolutionHandler represents a solution websocket handler.
func SolutionHandler(client *compute.Client, metadataCtor apiModel.MetadataStorageCtor,
	dataCtor apiModel.DataStorageCtor, solutionCtor apiModel.SolutionStorageCtor,
	modelCtor apiModel.ExportedModelStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// create conn
		conn, err := NewConnection(w, r, handleSolutionMessage(client, metadataCtor, dataCtor, solutionCtor, modelCtor))
		if err != nil {
			log.Warn(err)
			return
		}
		// listen for requests and respond
		err = conn.ListenAndRespond()
		if err != nil {
			log.Info(err)
		}
		// clean up conn internals
		conn.Close()
	}
}

func handleSolutionMessage(client *compute.Client, metadataCtor apiModel.MetadataStorageCtor,
	dataCtor apiModel.DataStorageCtor, solutionCtor apiModel.SolutionStorageCtor,
	modelCtor apiModel.ExportedModelStorageCtor) func(conn *Connection, bytes []byte) {
	return func(conn *Connection, bytes []byte) {
		// parse the message
		msg, err := NewMessage(bytes)
		if err != nil {
			// send error response
			handleErr(conn, nil, errors.Wrap(err, fmt.Sprintf("unable to parse solution request message: %s", string(bytes))))
			return
		}
		// handle message
		go handleMessage(conn, client, metadataCtor, dataCtor, solutionCtor, modelCtor, msg)
	}
}

func handleMessage(conn *Connection, client *compute.Client, metadataCtor apiModel.MetadataStorageCtor,
	dataCtor apiModel.DataStorageCtor, solutionCtor apiModel.SolutionStorageCtor, modelCtor apiModel.ExportedModelStorageCtor,
	msg *Message) {
	switch msg.Type {
	case createSolutions:
		handleCreateSolutions(conn, client, metadataCtor, dataCtor, solutionCtor, msg)
		return
	case stopSolutions:
		handleStopSolutions(conn, client, msg)
		return
	case predict:
		handlePredict(conn, client, metadataCtor, dataCtor, solutionCtor, modelCtor, msg)
	case query:
		handleQuery(conn, client, metadataCtor, dataCtor, solutionCtor, msg)
	default:
		// unrecognized type
		handleErr(conn, msg, errors.New("unrecognized message type"))
		return
	}
}

func handleCreateSolutions(conn *Connection, client *compute.Client, metadataCtor apiModel.MetadataStorageCtor,
	dataCtor apiModel.DataStorageCtor, solutionCtor apiModel.SolutionStorageCtor, msg *Message) {
	dataset, err := api.ExtractDatasetFromRawRequest(msg.Body)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to pull dataset from request"))
		return
	}

	// initialize the storage
	dataStorage, err := dataCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize data storage"))
		return
	}

	// initialize metadata storage
	metaStorage, err := metadataCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize meta storage"))
		return
	}

	// initialize solution storage
	solutionStorage, err := solutionCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize solution storage"))
		return
	}

	vars, err := metaStorage.FetchVariables(dataset, false, true)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to pull variables from storage"))
		return
	}

	// unmarshal request
	request, err := api.NewSolutionRequest(vars, msg.Body)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to unmarshal create solutions request"))
		return
	}

	// load defaults
	config, _ := env.LoadConfig()
	if len(request.Task) == 0 {
		request.Task = api.DefaultTaskType(request.TargetFeature.Type, request.ProblemType)
		log.Infof("Defaulting task type to `%s`", request.Task)
	}
	if len(request.Metrics) == 0 {
		request.Metrics = api.DefaultMetrics(request.Task)
		log.Infof("Defaulting metrics to `%s`", strings.Join(request.Metrics, ","))
	}
	if request.MaxTime == 0 {
		request.MaxTime = config.SolutionSearchMaxTime
		log.Infof("Defaulting max search time to `%d`", request.MaxTime)
	}

	// set augmentation info
	requestDataset, err := metaStorage.FetchDataset(request.Dataset, true, true)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to pull joined dataset"))
		return
	}

	if requestDataset.JoinSuggestions != nil {
		request.DatasetAugmentations = make([]*model.DatasetOrigin, len(requestDataset.JoinSuggestions))
		for i, js := range requestDataset.JoinSuggestions {
			request.DatasetAugmentations[i] = js.DatasetOrigin
		}
	}

	// persist the request information and dispatch the request
	err = request.PersistAndDispatch(client, solutionStorage, metaStorage, dataStorage)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to dispatch solution request to TA2"))
		return
	}

	// listen for solution updates - handler runs under a separate go routine
	requestFinished := make(chan api.SolutionStatus, 1)
	defer close(requestFinished)
	err = request.Listen(func(status api.SolutionStatus) {

		// update the map of currently running requests - this is crappy because go's
		// read/write locks are not upgradable or re-entrant
		requestMap.RLock()
		if _, ok := requestMap.m[status.RequestID]; !ok {
			requestMap.RUnlock()
			requestMap.Lock()
			requestMap.m[status.RequestID] = request
			requestMap.Unlock()
		} else {
			requestMap.RUnlock()
		}

		// send status to client - this includes any error status we encountered
		handleSuccess(conn, msg, jutil.StructToMap(status))

		// flag request as finished if it completed normally, or an error occurred
		// note that normally can include a cancellation, as some pipelines may have completed successfully
		if status.Progress == compute.RequestCompletedStatus || status.Progress == compute.RequestErroredStatus {
			// remove completed
			requestMap.Lock()
			delete(requestMap.m, status.RequestID)
			requestMap.Unlock()

			requestFinished <- status
		}
	})
	// something went wrong internally when setting up the request handling (downstream errors should come
	// through as status messages handled by the listener)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "received internal error"))
		return
	}

	// wait on request completed / request errored status before we move on
	<-requestFinished

	// complete the request
	handleComplete(conn, msg)
}

func handleStopSolutions(conn *Connection, client *compute.Client, msg *Message) {
	// unmarshal request
	request, err := api.NewStopSolutionSearchRequest(msg.Body)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to unmarshal stop solutions request"))
		return
	}

	// dispatch request to ta2
	err = request.Dispatch(client)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "received error from TA2 system"))
		return
	}

	// cancel further requests from the client side
	requestMap.RLock()
	req, ok := requestMap.m[request.RequestID]
	if ok {
		for _, cancelFunc := range req.CancelFuncs {
			cancelFunc()
		}
	}
	requestMap.RUnlock()
}

func handleQuery(conn *Connection, client *compute.Client, metadataCtor apiModel.MetadataStorageCtor,
	dataCtor apiModel.DataStorageCtor, solutionCtor apiModel.SolutionStorageCtor, msg *Message) {
	// create the storage instances
	dataStorage, err := dataCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize data storage"))
		return
	}

	// initialize metadata storage
	metaStorage, err := metadataCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize meta storage"))
		return
	}

	// parse parameters from the message
	req, err := api.NewQueryRequest(msg.Body)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to parse query request"))
		return
	}

	params := task.QueryParams{
		DataStorage: dataStorage,
		MetaStorage: metaStorage,
		Dataset:     req.Dataset,
		TargetName:  req.Target,
		Filters:     req.Filters,
	}
	_, err = task.Query(params)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to execute query request"))
		return
	}
}

func handlePredict(conn *Connection, client *compute.Client, metadataCtor apiModel.MetadataStorageCtor,
	dataCtor apiModel.DataStorageCtor, solutionCtor apiModel.SolutionStorageCtor,
	modelCtor apiModel.ExportedModelStorageCtor, msg *Message) {

	// initialize the storage
	dataStorage, err := dataCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize data storage"))
		return
	}

	// initialize metadata storage
	metaStorage, err := metadataCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize meta storage"))
		return
	}

	// initialize solution storage
	solutionStorage, err := solutionCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize solution storage"))
		return
	}

	// initialize save model storage
	modelStorage, err := modelCtor()
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to initialize solution storage"))
		return
	}

	// unmarshal request
	request, err := api.NewPredictRequest(msg.Body)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to unmarshal create solutions request"))
		return
	}

	// get the solution id from the fitted solution ID
	solutionResults, err := solutionStorage.FetchSolutionResultsByFittedSolutionID(request.FittedSolutionID)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to fetch solution results fitted solution id"))
		return
	}
	if len(solutionResults) == 0 {
		handleErr(conn, msg, errors.Errorf("unable to map fitted solution id to dataset or solution id"))
		return
	}
	sr := solutionResults[0]

	// read the metadata of the original dataset
	datasetES, err := metaStorage.FetchDataset(sr.Dataset, false, false)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to fetch dataset from es"))
		return
	}

	// get the source dataset from the fitted solution ID
	req, err := solutionStorage.FetchRequestByFittedSolutionID(sr.FittedSolutionID)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	schemaPath := path.Join(env.ResolvePath(datasetES.Source, datasetES.Folder), compute.D3MDataSchema)
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaPath, true)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to load metadata from source dataset schema doc"))
		return
	}

	target := getTarget(req)

	// In the case of grouped variables, the target will not be variable itself, but one of its property
	// values.  We need to fetch using the original dataset, since it will have grouped variable info,
	// and then resolve the actual target.
	targetVar, err := metaStorage.FetchVariable(meta.ID, target)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	variables, err := metaStorage.FetchVariablesByName(req.Dataset, req.Filters.Variables, false, false)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// resolve the task so we know what type of data we should be expecting
	requestTask, err := api.ResolveTask(dataStorage, meta.StorageName, targetVar, variables)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// config objects required for ingest
	config, _ := env.LoadConfig()
	ingestConfig := task.NewConfig(config)

	predictParams := &task.PredictParams{
		Meta:             meta,
		Dataset:          request.DatasetID,
		SolutionID:       sr.SolutionID,
		FittedSolutionID: request.FittedSolutionID,
		OutputPath:       path.Join(config.D3MOutputDir, config.AugmentedSubFolder),
		Target:           targetVar,
		MetaStorage:      metaStorage,
		DataStorage:      dataStorage,
		SolutionStorage:  solutionStorage,
		ModelStorage:     modelStorage,
		DatasetIngested:  false,
		DatasetImported:  false,
		Config:           &config,
		IngestConfig:     ingestConfig,
	}

	ds, err := createPredictionDataset(requestTask, request, predictParams)
	if err != nil {
		handleErr(conn, msg, errors.Wrap(err, "unable to create raw dataset"))
		return
	}
	predictParams.DatasetConstructor = ds

	// run predictions - synchronous call for now
	result, err := task.Predict(predictParams)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// send the status update to the client
	// TODO: we are only sending when complete - should send in progress similar to solution create
	handleSuccess(conn, msg, jutil.StructToMap(result))

	// notify the client that we're done
	handleComplete(conn, msg)
}

func getTarget(request *apiModel.Request) string {
	for _, f := range request.Features {
		if f.FeatureType == "target" {
			return f.FeatureName
		}
	}

	return ""
}

func createPredictionDataset(requestTask *api.Task, request *api.PredictRequest,
	predictParams *task.PredictParams) (task.DatasetConstructor, error) {
	datasetID := request.DatasetID
	datasetPath := request.DatasetPath
	var ds task.DatasetConstructor
	var err error
	if api.HasTaskType(requestTask, compute.RemoteSensingTask) {
		ds, err = dataset.NewSatelliteDataset(datasetID, "tif", datasetPath)
	} else if api.HasTaskType(requestTask, compute.ImageTask) {
		ds, err = dataset.NewMediaDataset(datasetID, "png", "jpeg", datasetPath)
	} else if api.HasTaskType(requestTask, compute.TimeSeriesTask) && api.HasTaskType(requestTask, compute.ForecastingTask) {
		ds, err = task.NewPredictionTimeseriesDataset(predictParams, request.IntervalLength, request.IntervalCount)
	} else {
		var data []byte
		data, err = ioutil.ReadFile(datasetPath)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to read raw tabular data")
		}
		ds, err = dataset.NewTableDataset(datasetID, data, false)
	}
	if err != nil {
		return nil, err
	}

	return ds, nil
}
