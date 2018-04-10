package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"github.com/fatih/structs"

	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

const (
	getSession        = "GET_SESSION"
	endSession        = "END_SESSION"
	createPipelines   = "CREATE_PIPELINES"
	streamClose       = "STREAM_CLOSE"
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
	case createPipelines:
		handleCreatePipelines(conn, client, metadataCtor, dataCtor, pipelineCtor, msg)
		return
	default:
		// unrecognized type
		handleErr(conn, msg, errors.New("unrecognized message type"))
		return
	}
}

func handleCreatePipelines(conn *Connection, client *pipeline.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, pipelineCtor model.PipelineStorageCtor, msg *Message) {

	// unmarshall the request data
	createMessage := &pipeline.CreateMessage{}
	err := json.Unmarshal(msg.Raw, createMessage)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// initialize the storage
	dataStorage, err := dataCtor()
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// initialize metadata storage
	metaStorage, err := metadataCtor()
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

	// persist the request information and dispatch the request
	statusChannels, err := createMessage.PersistAndDispatch(client, pipelineStorage, metaStorage, dataStorage)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	for _, c := range statusChannels {
		// TODO: listen and respond to client
		go func(statusChannel chan pipeline.PipelineStatus) {
			// read status from, channel
			status <- statusChannel
			// check for error
			if status.Error != nil {
				handleErr(conn, msg, err)
				return
			}
			// send status to client
			handleSuccess(conn, msg,  structs.Map(status))
		}(c)
	}
}
