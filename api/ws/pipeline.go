package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/unchartedsoftware/distil/api/pipeline"
	"github.com/unchartedsoftware/plog"
	"golang.org/x/net/context"

	es "gopkg.in/olivere/elastic.v5"
)

const (
	getSession       = "GET_SESSION"
	endSession       = "END_SESSION"
	createPipelines  = "CREATE_PIPELINES"
	streamClose      = "STREAM_CLOSE"
	datasetDir       = "datasets"
	categoricalType  = "categorical"
	numericalType    = "numerical"
	datasetSizeLimit = 10000
)

// PipelineHandler represents a pipeline websocket handler.
func PipelineHandler(client *pipeline.Client, esCtor elastic.ClientCtor, storageCtor model.StorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// create conn
		conn, err := NewConnection(w, r, handlePipelineMessage(client, esCtor, storageCtor))
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

func handlePipelineMessage(client *pipeline.Client, esCtor elastic.ClientCtor, storageCtor model.StorageCtor) func(conn *Connection, bytes []byte) {
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
		go handleMessage(conn, client, esCtor, storageCtor, msg)
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

func handleMessage(conn *Connection, client *pipeline.Client, esCtor elastic.ClientCtor, storageCtor model.StorageCtor, msg *Message) {
	switch msg.Type {
	case getSession:
		handleGetSession(conn, client, msg, storageCtor)
		return
	case endSession:
		handleEndSession(conn, client, msg)
		return
	case createPipelines:
		handleCreatePipelines(conn, client, esCtor, storageCtor, msg)
		return
	default:
		// unrecognized type
		handleErr(conn, msg, errors.New("unrecognized message type"))
		return
	}
}

func loadSessionRequests(msg *Message, session *pipeline.Session, storage model.Storage) error {
	// load the stored session information.
	log.Infof("Loading requests for session %v.", msg.Session)
	reqs, err := storage.FetchRequests(msg.Session)
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

func handleGetSession(conn *Connection, client *pipeline.Client, msg *Message, storageCtor model.StorageCtor) {
	// get the storage instance
	storage, err := storageCtor()
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

	// store the sessions
	err = storage.PersistSession(session.ID)
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
	Index        string          `json:"index"`
	Feature      string          `json:"feature"`
	Task         string          `json:"task"`
	Output       string          `json:"output"`
	MaxPipelines int32           `json:"maxPipelines"`
	Filters      json.RawMessage `json:"filters"`
	Metrics      []string        `json:"metric"`
}

func handleCreatePipelines(conn *Connection, client *pipeline.Client, esCtor elastic.ClientCtor, storageCtor model.StorageCtor, msg *Message) {
	// unmarshall the request data
	clientCreateMsg := &pipelineCreateMsg{}
	err := json.Unmarshal(msg.Raw, clientCreateMsg)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// parse the features out of the create msg - done as a separate step because their structure isn't entirely
	// fixed
	filters, err := parseDatasetFilters(clientCreateMsg.Filters)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// initialize the storage
	storage, err := storageCtor()
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// initialize ES client
	esClient, err := esCtor()
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// persist the filtered dataset if necessary
	fetchFilteredData := func(dataset string, index string, filters *model.FilterParams, inclusive bool) (*model.FilteredData, error) {
		return model.FetchFilteredData(storage, dataset, index, filters, inclusive)
	}
	fetchVariable := func(dataset string, index string) ([]*model.Variable, error) {
		return model.FetchVariables(esClient, index, dataset)
	}
	datasetPath, err := pipeline.PersistFilteredData(fetchFilteredData, fetchVariable, client.DataDir, clientCreateMsg.Dataset, clientCreateMsg.Index, clientCreateMsg.Feature, filters, true)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// make sure the path is absolute
	datasetPath, err = filepath.Abs(datasetPath)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// Create the set of training features - we already filtered that out when we persist, but needs to be specified
	// to satisfy ta3ta2 API.
	trainFeatures := []*pipeline.Feature{}
	filteredVars, err := fetchFilteredVariables(esClient, clientCreateMsg.Index, clientCreateMsg.Dataset, filters)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}
	for _, featureName := range filteredVars {
		feature := &pipeline.Feature{
			FeatureId: featureName,
			DataUri:   datasetPath,
		}
		trainFeatures = append(trainFeatures, feature)
	}

	// convert received metrics into the ta3ta2 format
	metrics := []pipeline.Metric{}
	for _, msgMetric := range clientCreateMsg.Metrics {
		metric := pipeline.Metric(pipeline.Metric_value[strings.ToUpper(msgMetric)])
		metrics = append(metrics, metric)
	}

	// populate the protobuf pipeline create msg
	createMsg := &pipeline.PipelineCreateRequest{
		Context:       &pipeline.SessionContext{SessionId: msg.Session},
		TrainFeatures: trainFeatures,
		Task:          pipeline.TaskType(pipeline.TaskType_value[strings.ToUpper(clientCreateMsg.Task)]),
		Output:        pipeline.OutputType(pipeline.OutputType_value[strings.ToUpper(clientCreateMsg.Output)]),
		Metrics:       metrics,
		TargetFeatures: []*pipeline.Feature{
			{
				FeatureId: clientCreateMsg.Feature,
				DataUri:   datasetPath,
			},
		},
		MaxPipelines: clientCreateMsg.MaxPipelines,
	}

	// kick off the pipeline creation, or re-attach to one that is already running
	if session, ok := client.GetSession(msg.Session); ok {
		requestInfo := pipeline.GeneratePipelineCreateRequest(createMsg)
		proxy, err := session.GetOrDispatch(context.Background(), requestInfo)
		if err != nil {
			handleErr(conn, msg, err)
			return
		}

		// store the request using the initial progress value
		requestID := fmt.Sprintf("%s", requestInfo.RequestID)
		err = storage.PersistRequest(session.ID, requestID, clientCreateMsg.Dataset, pipeline.Progress_name[0])
		if err != nil {
			handleErr(conn, msg, err)
			return
		}

		// store the request features
		for _, f := range trainFeatures {
			err = storage.PersistRequestFeature(requestID, f.FeatureId, model.FeatureTypeTrain)
			if err != nil {
				handleErr(conn, msg, err)
				return
			}
		}

		for _, f := range createMsg.TargetFeatures {
			err = storage.PersistRequestFeature(requestID, f.FeatureId, model.FeatureTypeTarget)
			if err != nil {
				handleErr(conn, msg, err)
				return
			}
		}

		// handle the request
		handleCreatePipelinesSuccess(conn, msg, proxy, storage, clientCreateMsg.Dataset)
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

func handleCreatePipelinesSuccess(conn *Connection, msg *Message, proxy *pipeline.ResultProxy, storage model.Storage, dataset string) {
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
					"requestId":  proxy.RequestID,
					"pipelineId": res.PipelineId,
					"progress":   progress,
					"dataset":    dataset,
				}
				log.Infof("Pipeline %s - %s", res.PipelineId, progress)

				// update the request progress
				err := storage.UpdateRequest(fmt.Sprintf("%s", proxy.RequestID), progress)
				if err != nil {
					handleErr(conn, msg, errors.Wrap(err, "Unable to store request update"))
				}

				// on complete, fetch results as well
				if res.ProgressInfo == pipeline.Progress_COMPLETED || res.ProgressInfo == pipeline.Progress_UPDATED {
					scores := make([]map[string]interface{}, 0)
					for _, score := range res.PipelineInfo.Scores {
						s := map[string]interface{}{
							"metric": pipeline.Metric_name[int32(score.Metric)],
							"value":  score.Value,
						}
						scores = append(scores, s)

						// store the result score
						if res.ProgressInfo == pipeline.Progress_COMPLETED {
							storage.PersistResultScore(res.PipelineId, s["metric"].(string), float64(s["value"].(float32)))
						}
					}
					response["pipeline"] = map[string]interface{}{
						"scores":    scores,
						"output":    pipeline.OutputType_name[int32(res.PipelineInfo.Output)],
						"resultUri": res.PipelineInfo.PredictResultUris[0],
					}

					// store the result data & metadata
					err = storage.PersistResultMetadata(fmt.Sprintf("%s", proxy.RequestID), res.PipelineId, "", res.PipelineInfo.PredictResultUris[0], progress)
					if err != nil {
						handleErr(conn, msg, errors.Wrap(err, "Unable to store result metadata"))
					}

					err = storage.PersistResult(dataset, res.PipelineInfo.PredictResultUris[0])
					if err != nil {
						handleErr(conn, msg, errors.Wrap(err, "Unable to store pipeline results"))
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

// TODO: We don't store this anywhere, so we end up running an ES query to get the var list.  This should
// be cached by Redis, but still worth looking into storing some of the dataset info.
func fetchFilteredVariables(esClient *es.Client, index string, dataset string, filters *model.FilterParams) ([]string, error) {
	// put the filtered variables into a set for quick lookup
	nameSet := map[string]bool{}
	for _, varName := range filters.None {
		nameSet[varName] = true
	}

	// fetch the variable set from es
	variables, err := model.FetchVariables(esClient, index, dataset)
	if err != nil {
		return nil, err
	}

	// create a list minus those that are in the filtered list
	filteredVars := []string{}
	for _, variable := range variables {
		if _, ok := nameSet[variable.Name]; !ok {
			filteredVars = append(filteredVars, variable.Name)
		}
	}
	return filteredVars, nil
}

// pointers used to support optional field pattern
type filter struct {
	Name       string
	Enabled    bool
	Type       *string
	Min        *float64
	Max        *float64
	Categories *[]string
}

// parse filter parameters out of JSON
func parseDatasetFilters(rawFilters json.RawMessage) (*model.FilterParams, error) {
	// filter params for subsequent store query
	filterParams := model.FilterParams{}
	filterParams.Size = datasetSizeLimit

	// unmarshall from params porition of message
	var filters map[string]filter
	err := json.Unmarshal(rawFilters, &filters)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse filters")
	}

	// sort the filter values by var name to ensure consistent hashing
	//
	// TODO: this can possibly be circumvented by having the client pass
	// the filter params up as a sorted list rather than a map
	filterValues := make([]*filter, 0, len(filters))
	for k := range filters {
		f := filters[k]
		filterValues = append(filterValues, &f)
	}
	sort.SliceStable(filterValues, func(i, j int) bool {
		return filterValues[i].Name < filterValues[j].Name
	})

	for _, filter := range filterValues {
		// parse out filter parameters
		if filter.Type != nil {
			if *filter.Type == numericalType {
				if filter.Min == nil || filter.Max == nil {
					return nil, errors.New("numerical filter missing min/max value")
				}
				varRange := model.VariableRange{Name: filter.Name, Min: *filter.Min, Max: *filter.Max}
				filterParams.Ranged = append(filterParams.Ranged, varRange)
			} else if *filter.Type == categoricalType {
				if filter.Categories == nil {
					return nil, errors.New("categorical filter missing categories set")
				}
				sort.Strings(*filter.Categories)
				varCategories := model.VariableCategories{Name: filter.Name, Categories: *filter.Categories}
				filterParams.Categorical = append(filterParams.Categorical, varCategories)
			} else {
				return nil, errors.Errorf("unknown filter type %s", *filter.Type)
			}
		} else {
			filterParams.None = append(filterParams.None, filter.Name)
		}
		sort.Strings(filterParams.None)
	}
	return &filterParams, nil
}
