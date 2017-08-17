package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/unchartedsoftware/distil/api/model"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	es "github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/pipeline"
	"github.com/unchartedsoftware/plog"
	"golang.org/x/net/context"
)

const (
	getSession      = "GET_SESSION"
	endSession      = "END_SESSION"
	createPipelines = "CREATE_PIPELINES"
	streamClose     = "STREAM_CLOSE"
	datasetDir      = "datasets"
)

// PipelineHandler represents a pipeline websocket handler.
func PipelineHandler(client *pipeline.Client, esClientCtor es.ClientCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// create conn
		conn, err := NewConnection(w, r, handlePipelineMessage(client, esClientCtor))
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

func handlePipelineMessage(client *pipeline.Client, esClientCtor es.ClientCtor) func(conn *Connection, bytes []byte) {
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
		go handleMessage(conn, client, esClientCtor, msg)
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

func handleMessage(conn *Connection, client *pipeline.Client, esClientCtor es.ClientCtor, msg *Message) {
	switch msg.Type {
	case getSession:
		handleGetSession(conn, client, msg)
		return
	case endSession:
		handleEndSession(conn, client, msg)
		return
	case createPipelines:
		handleCreatePipelines(conn, client, esClientCtor, msg)
		return
	default:
		// unrecognized type
		handleErr(conn, msg, errors.New("unrecognized message type"))
		return
	}
}

func handleGetSession(conn *Connection, client *pipeline.Client, msg *Message) {
	// get existing session
	if msg.Session != "" {
		// try to get existing session
		session, ok := client.GetSession(msg.Session)
		if ok {
			handleGetSessionSuccess(conn, msg, session.ID, false, true, session.GetExistingUUIDs())
			return
		}
	}
	// start a new session
	session, err := client.StartSession(context.Background())
	if err != nil {
		handleErr(conn, msg, err)
		return
	}
	handleGetSessionSuccess(conn, msg, session.ID, true, false, session.GetExistingUUIDs())
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
	Feature      string          `json:"feature"`
	Task         string          `json:"task"`
	Metric       string          `json:"metric"`
	Output       string          `json:"output"`
	MaxPipelines int32           `json:"maxPipelines"`
	Filters      json.RawMessage `json:"filters"`
}

func handleCreatePipelines(conn *Connection, client *pipeline.Client, esClientCtor es.ClientCtor, msg *Message) {
	// unmarshall the request data
	clientCreateMsg := &pipelineCreateMsg{}
	err := json.Unmarshal(msg.Raw, clientCreateMsg)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// persist the filtered dataset if necessary
	fetchFilteredData := func(dataset string, filters *model.FilterParams) (*model.FilteredData, error) {
		esClient, err := esClientCtor()
		if err != nil {
			return nil, err
		}
		return model.FetchFilteredData(esClient, dataset, filters)
	}
	datasetPath, err := pipeline.PersistFilteredData(fetchFilteredData, datasetDir, clientCreateMsg.Dataset, clientCreateMsg.Filters)
	if err != nil {
		handleErr(conn, msg, err)
	}

	// populate the protobuf pipeline create msg
	createMsg := &pipeline.PipelineCreateRequest{
		Context:          &pipeline.SessionContext{SessionId: msg.Session},
		TrainDatasetUris: []string{datasetPath},
		Task:             pipeline.Task(pipeline.Task_value[strings.ToUpper(clientCreateMsg.Task)]),
		Output:           pipeline.Output(pipeline.Output_value[strings.ToUpper(clientCreateMsg.Output)]),
		Metric:           []pipeline.Metric{pipeline.Metric(pipeline.Metric_value[strings.ToUpper(clientCreateMsg.Metric)])},
		TargetFeatures:   []string{clientCreateMsg.Feature},
		MaxPipelines:     clientCreateMsg.MaxPipelines,
	}

	// kick off the pipeline creation, or re-attach to one that is already running
	if session, ok := client.GetSession(msg.Session); ok {
		requestInfo := pipeline.GeneratePipelineCreateRequest(createMsg)
		proxy, err := session.GetOrDispatch(context.Background(), requestInfo)
		if err != nil {
			handleErr(conn, msg, err)
			return
		}
		handleCreatePipelinesSuccess(conn, msg, proxy)
	} else {
		log.Warnf("Expected session %s does not exist", msg.Session)
	}
	return
}

func handleGetSessionSuccess(conn *Connection, msg *Message, session string, created bool, resumed bool, uuids []uuid.UUID) {
	// convert uuids to strings
	var strs []string
	for _, uid := range uuids {
		strs = append(strs, uid.String())
	}
	// send response
	handleSuccess(conn, msg, map[string]interface{}{
		"success": true,
		"session": session,
		"created": created,
		"resumed": resumed,
		"uuids":   uuids,
	})
}

func handleEndSessionSuccess(conn *Connection, msg *Message) {
	// send response
	handleSuccess(conn, msg, map[string]interface{}{
		"success": true,
	})
}

func handleCreatePipelinesSuccess(conn *Connection, msg *Message, proxy *pipeline.ResultProxy) {
	// process the result proxy, which is replicated for completed, pending requests
	for {
		select {
		case result := <-proxy.Results:
			res := (*result).(*pipeline.PipelineCreateResult)
			// check to see if the server is handling the request successfully
			if res.ResponseInfo.Status.Code == pipeline.StatusCode_OK {
				// extract the baseline pipeline status
				progress := pipeline.Progress_name[int32(res.ProgressInfo)]
				response := map[string]interface{}{
					"session":    res.ResponseInfo.Context.SessionId,
					"pipelineId": res.PipelineId,
					"progress":   progress,
				}
				log.Infof("Pipeline %s - %s", res.PipelineId, progress)

				// on complete, fetch results as well
				if res.ProgressInfo == pipeline.Progress_COMPLETE {
					scores := make([]map[string]interface{}, 0)
					for _, score := range res.PipelineInfo.Score {
						scores = append(scores, map[string]interface{}{
							"metric": pipeline.Metric_name[int32(score.Metric)],
							"value":  score.Value,
						})
					}
					response["pipeline"] = map[string]interface{}{
						"scores": scores,
						"output": pipeline.Output_name[int32(res.PipelineInfo.Output)],
					}
				}
				handleSuccess(conn, msg, response)
			} else {
				status := res.ResponseInfo.Status.Code
				statusDesc := res.ResponseInfo.Status.Details
				handleErr(conn, msg, errors.Errorf("pipeline create failed - %s: %s", status, statusDesc))
				return
			}
		case err := <-proxy.Errors:
			handleErr(conn, msg, err)
			return
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
