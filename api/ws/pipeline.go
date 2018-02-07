package ws

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/unchartedsoftware/plog"
	"golang.org/x/net/context"

	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

const (
	getSession        = "GET_SESSION"
	endSession        = "END_SESSION"
	createPipelines   = "CREATE_PIPELINES"
	streamClose       = "STREAM_CLOSE"
	datasetDir        = "datasets"
	categoricalType   = "categorical"
	numericalType     = "numerical"
	defaultResourceID = "0"
	datasetSizeLimit  = 10000
)

// PipelineHandler represents a pipeline websocket handler.
func PipelineHandler(client *pipeline.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, pipelineCtor model.PipelineStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// create conn
		conn, err := NewConnection(w, r, handlePipelineMessage(client, metadataCtor, dataCtor, pipelineCtor))
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

func handlePipelineMessage(client *pipeline.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, pipelineCtor model.PipelineStorageCtor) func(conn *Connection, bytes []byte) {
	return func(conn *Connection, bytes []byte) {
		// parse the message
		msg, err := NewMessage(bytes)
		if err != nil {
			// parsing error, send back a failure response
			err := fmt.Errorf("unable to parse pipeline request message: %s", string(bytes))
			// send error response
			handleErr(conn, nil, err)
			return
		}
		// handle message
		go handleMessage(conn, client, metadataCtor, dataCtor, pipelineCtor, msg)
	}
}

func parseMessage(bytes []byte) (*Message, error) {
	var msg *Message
	err := json.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, err
	}
	msg.Timestamp = time.Now()
	return msg, nil
}

func handleMessage(conn *Connection, client *pipeline.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, pipelineCtor model.PipelineStorageCtor, msg *Message) {
	switch msg.Type {
	case getSession:
		handleGetSession(conn, client, msg, pipelineCtor)
		return
	case endSession:
		handleEndSession(conn, client, msg)
		return
	case createPipelines:
		handleCreatePipelines(conn, client, metadataCtor, dataCtor, pipelineCtor, msg)
		return
	default:
		// unrecognized type
		handleErr(conn, msg, errors.New("unrecognized message type"))
		return
	}
}

func loadSessionRequests(msg *Message, session *pipeline.Session, pipelineStorage model.PipelineStorage) error {
	// load the stored session information.
	log.Infof("Loading requests for session %v.", msg.Session)
	reqs, err := pipelineStorage.FetchRequests(msg.Session)
	if err != nil {
		return errors.Wrap(err, "Unable to pull session request")
	}

	// parse the requests into the session object.
	for _, r := range reqs {
		// get the uuid for the request.
		requestID, err := uuid.FromString(r.RequestID)
		if err != nil {
			return errors.Wrap(err, "Unable to parse request uuid")
		}

		// add the request to the right collection
		req := &pipeline.RequestContext{
			RequestID: requestID,
		}
		if pipeline.Progress_value[r.Progress] != int32(pipeline.Progress_COMPLETED) {
			session.AddPendingRequest(req)
		} else {
			session.AddCompletedRequest(req)
		}
	}

	log.Infof("Requests for session %v loaded successfully.", msg.Session)

	return nil
}

func handleGetSession(conn *Connection, client *pipeline.Client, msg *Message, pipelineCtor model.PipelineStorageCtor) {
	// get the storage instance
	storage, err := pipelineCtor()
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// get existing session
	if msg.Session != "" {
		// try to get existing session
		session, ok := client.GetSession(msg.Session)
		if ok {
			err = loadSessionRequests(msg, session, storage)
			if err != nil {
				handleErr(conn, msg, err)
				return
			}

			handleGetSessionSuccess(conn, msg, session.ID, false, true)
			return
		}
	}

	// start a new session
	session, err := client.StartSession(context.Background())
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// store the sessions
	err = storage.PersistSession(session.ID)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	handleGetSessionSuccess(conn, msg, session.ID, true, false)
	return
}

func handleEndSession(conn *Connection, client *pipeline.Client, msg *Message) {
	// end session
	err := client.EndSession(context.Background(), msg.Session)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}
	handleEndSessionSuccess(conn, msg)
	return
}

type pipelineCreateMsg struct {
	Dataset      string          `json:"dataset"`
	Index        string          `json:"index"`
	Feature      string          `json:"feature"`
	Task         string          `json:"task"`
	MaxPipelines int32           `json:"maxPipelines"`
	Filters      json.RawMessage `json:"filters"`
	Metrics      []string        `json:"metric"`
}

func handleCreatePipelines(conn *Connection, client *pipeline.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, pipelineCtor model.PipelineStorageCtor, msg *Message) {
	// unmarshall the request data
	clientCreateMsg := &pipelineCreateMsg{}
	err := json.Unmarshal(msg.Raw, clientCreateMsg)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// parse the features out of the create msg - done as a separate step because their structure isn't entirely
	// fixed
	filters, err := model.ParseFilterParamsJSON(clientCreateMsg.Filters)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}
	// NOTE: IF THE SIZE IS NOT SET THEN THE DEFAULT IS USED (100 rows only)!!!
	// NOTE: this could be done on the client side, but I am not sure if that
	// is more elegant or not.
	filters.Size = -1

	// initialize the storage
	dataStorage, err := dataCtor()
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// initialize metadata storage
	metadata, err := metadataCtor()
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// initialize pipeline storage
	pipelineStorage, err := pipelineCtor()
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// persist the filtered dataset if necessary
	fetchFilteredData := func(dataset string, index string, filterParams *model.FilterParams) (*model.FilteredData, error) {
		// fetch the whole data and include the target feature
		return dataStorage.FetchData(dataset, index, filterParams, false)
	}
	fetchVariable := func(dataset string, index string) ([]*model.Variable, error) {
		return metadata.FetchVariables(dataset, index, false)
	}
	datasetPath, err := pipeline.PersistFilteredData(fetchFilteredData, fetchVariable, client.DataDir, clientCreateMsg.Dataset, clientCreateMsg.Index, clientCreateMsg.Feature, filters)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// make sure the path is absolute and contains the URI prefix
	datasetPath, err = filepath.Abs(datasetPath)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}
	datasetPath = fmt.Sprintf("%s", filepath.Join(datasetPath, pipeline.D3MDataSchema))

	// Create the set of training features - we already filtered that out when we persist, but needs to be specified
	// to satisfy ta3ta2 API.
	filteredVars, err := fetchFilteredVariables(metadata, clientCreateMsg.Index, clientCreateMsg.Dataset, filters)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}
	trainFeatures := []*pipeline.Feature{}
	for _, featureName := range filteredVars {
		if featureName != clientCreateMsg.Feature {
			feature := &pipeline.Feature{
				FeatureName: featureName,
				ResourceId:  defaultResourceID,
			}
			trainFeatures = append(trainFeatures, feature)
		}
	}

	// convert received metrics into the ta3ta2 format
	metrics := []pipeline.PerformanceMetric{}
	for _, msgMetric := range clientCreateMsg.Metrics {
		metric := pipeline.PerformanceMetric(pipeline.PerformanceMetric_value[strings.ToUpper(msgMetric)])
		metrics = append(metrics, metric)
	}

	// populate the protobuf pipeline create msg
	createMsg := &pipeline.PipelineCreateRequest{
		Context: &pipeline.SessionContext{
			SessionId: msg.Session,
		},
		PredictFeatures: trainFeatures,
		Task:            pipeline.TaskType(pipeline.TaskType_value[strings.ToUpper(clientCreateMsg.Task)]),
		Metrics:         metrics,
		DatasetUri:      datasetPath,
		TargetFeatures: []*pipeline.Feature{
			{
				FeatureName: clientCreateMsg.Feature,
				ResourceId:  defaultResourceID,
			},
		},
		MaxPipelines: clientCreateMsg.MaxPipelines,
	}

	// kick off the pipeline creation, or re-attach to one that is already running
	session, ok := client.GetSession(msg.Session)
	if !ok {
		log.Warnf("Expected session %s does not exist", msg.Session)
		return
	}

	requestInfo := pipeline.GeneratePipelineCreateRequest(createMsg)
	proxy, err := session.GetOrDispatch(context.Background(), requestInfo)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// store the request using the initial progress value
	requestID := fmt.Sprintf("%s", requestInfo.RequestID)
	err = pipelineStorage.PersistRequest(session.ID, requestID, clientCreateMsg.Dataset, pipeline.Progress_name[0], time.Now())
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// store the request features
	for _, f := range trainFeatures {
		err = pipelineStorage.PersistRequestFeature(requestID, f.FeatureName, model.FeatureTypeTrain)
		if err != nil {
			handleErr(conn, msg, err)
			return
		}
	}

	for _, f := range createMsg.TargetFeatures {
		err = pipelineStorage.PersistRequestFeature(requestID, f.FeatureName, model.FeatureTypeTarget)
		if err != nil {
			handleErr(conn, msg, err)
			return
		}
	}

	// store request filters
	err = pipelineStorage.PersistRequestFilters(requestID, filters)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// handle the request
	handleCreatePipelinesSuccess(conn, msg, proxy, dataStorage, pipelineStorage, clientCreateMsg.Dataset)
}

func handleGetSessionSuccess(conn *Connection, msg *Message, session string, created bool, resumed bool) {
	// send response
	handleSuccess(conn, msg, map[string]interface{}{
		"success": true,
		"session": session,
		"created": created,
		"resumed": resumed,
	})
}

func handleEndSessionSuccess(conn *Connection, msg *Message) {
	// send response
	handleSuccess(conn, msg, map[string]interface{}{
		"success": true,
	})
}

func handleCreatePipelinesSuccess(conn *Connection, msg *Message, proxy *pipeline.ResultProxy, dataStorage model.DataStorage, pipelineStorage model.PipelineStorage, dataset string) {
	// process the result proxy, which is replicated for completed, pending requests
	for {
		select {
		case result := <-proxy.Results:
			res := (*result).(*pipeline.PipelineCreateResult)

			if res.ResponseInfo.Status.Code != pipeline.StatusCode_OK {
				status := res.ResponseInfo.Status.Code
				statusDesc := res.ResponseInfo.Status.Details
				handleErr(conn, msg, errors.Errorf("pipeline create failed - %s: %s", status, statusDesc))
			}

			// extract the baseline pipeline status
			progress := pipeline.Progress_name[int32(res.ProgressInfo)]

			// update the request progress
			currentTime := time.Now()
			err := pipelineStorage.UpdateRequest(fmt.Sprintf("%s", proxy.RequestID), progress, currentTime)
			if err != nil {
				handleErr(conn, msg, errors.Wrap(err, "unable to store request update"))
			}
			response := map[string]interface{}{
				"requestId":  proxy.RequestID,
				"pipelineId": res.PipelineId,
				"progress":   progress,
			}

			// on complete, persist scores
			if res.ProgressInfo == pipeline.Progress_COMPLETED {
				for _, score := range res.PipelineInfo.Scores {

					scoreMetric := pipeline.PerformanceMetric_name[int32(score.Metric)]
					scoreValue := float64(score.Value)

					// store the result score
					pipelineStorage.PersistResultScore(res.PipelineId, scoreMetric, scoreValue)
				}
			}

			resultURI := ""
			resultID := ""
			outputType := ""

			// on update / complete, persist the resultURI
			if res.ProgressInfo == pipeline.Progress_COMPLETED ||
				res.ProgressInfo == pipeline.Progress_UPDATED {
				// Get the result URI, removing the protocol portion if it exists. The returned value
				// is either a csv or a directory.  If we get a directory back, it should match the standard structure.
				// Look for the trainTargets.csv
				resultURI = res.PipelineInfo.PredictResultUri
				resultURI = strings.Replace(resultURI, "file://", "", 1)
				if !strings.HasSuffix(resultURI, ".csv") {
					resultURI = path.Join(resultURI, pipeline.D3MLearningData)
				}

				// get the result UUID. NOTE: Doing sha1 for now.
				hasher := sha1.New()
				hasher.Write([]byte(resultURI))
				bs := hasher.Sum(nil)
				resultID = fmt.Sprintf("%x", bs)

				response["pipeline"] = map[string]interface{}{
					"resultId": resultID,
				}

				outputTypeName := int32(res.PipelineInfo.Output)
				outputType = pipeline.OutputType_name[outputTypeName]
			}

			// store the result metadata
			err = pipelineStorage.PersistResultMetadata(
				proxy.RequestID.String(),
				res.PipelineId,
				resultID,
				resultURI,
				progress,
				outputType,
				currentTime)
			if err != nil {
				handleErr(conn, msg, errors.Wrap(err, "unable to store result metadata"))
			}

			// persist results, if they are available
			if res.ProgressInfo == pipeline.Progress_COMPLETED ||
				res.ProgressInfo == pipeline.Progress_UPDATED {

				err = dataStorage.PersistResult(dataset, resultURI)
				if err != nil {
					handleErr(conn, msg, errors.Wrap(err, "unable to store pipeline results"))
				}
			}
			handleSuccess(conn, msg, response)

		case err := <-proxy.Errors:
			progress := pipeline.Progress_name[int32(pipeline.Progress_ERRORED)]
			currentTime := time.Now()

			otherErr := pipelineStorage.UpdateRequest(fmt.Sprintf("%s", proxy.RequestID), progress, currentTime)
			if otherErr != nil {
				handleErr(conn, msg, errors.Wrap(otherErr, "unable to store request update"))
			}

			otherErr = pipelineStorage.PersistResultMetadata(
				proxy.RequestID.String(),
				err.PipelineID,
				"",
				"",
				progress,
				"",
				currentTime)
			if otherErr != nil {
				handleErr(conn, msg, errors.Wrap(otherErr, "unable to store result metadata"))
			}

			handleErr(conn, msg, err.Error)

		case <-proxy.Done:
			// notify the downstream client that the stream is closed
			response := map[string]interface{}{
				streamClose: true,
			}
			handleSuccess(conn, msg, response)
			return
		}
	}
}

// TODO: We don't store this anywhere, so we end up running an ES query to get the var list.  This should
// be cached by Redis, but still worth looking into storing some of the dataset info.
func fetchFilteredVariables(metadata model.MetadataStorage, index string, dataset string, filters *model.FilterParams) ([]string, error) {
	// fetch the variable set from es
	variables, err := metadata.FetchVariables(dataset, index, false)
	if err != nil {
		return nil, err
	}

	variablesToUse := model.GetFilterVariables(filters, variables, false)

	// create a list minus those that are in the filtered list
	filteredVars := []string{}
	for _, variable := range variablesToUse {
		filteredVars = append(filteredVars, variable.Name)
	}
	return filteredVars, nil
}
