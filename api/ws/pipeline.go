package ws

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/satori/go.uuid"
	"github.com/unchartedsoftware/plog"
	"golang.org/x/net/context"

	"github.com/unchartedsoftware/distil/api/pipeline"
)

const (
	startSession  = "start"
	resumeSession = "resume"
	endSession    = "end"
)

// Message represents a websocket message.
type Message struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Session string `json:"session"`
}

// PipelineHandler represents a pipeline websocket handler.
func PipelineHandler(client *pipeline.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// create conn
		conn, err := NewConnection(w, r, handlePipelineMessage(client))
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

func handlePipelineMessage(client *pipeline.Client) func(conn *Connection, bytes []byte) {
	return func(conn *Connection, bytes []byte) {
		// parse the message
		msg, err := parseMessage(bytes)
		if err != nil {
			// parsing error, send back a failure response
			err := fmt.Errorf("unable to parse pipeline request message: %s", string(bytes))
			// send error response
			handleErr(conn, "missing", err)
			return
		}
		// handle message
		go handleMessage(conn, client, msg)
	}
}

func parseMessage(bytes []byte) (*Message, error) {
	var msg *Message
	err := json.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func handleMessage(conn *Connection, client *pipeline.Client, msg *Message) {
	switch msg.Type {
	case startSession:
		// start session
		log.Debugf("START SESSION REQ")
		session, err := client.StartSession(context.Background())
		if err != nil {
			handleErr(conn, msg.ID, err)
			return
		}
		handleStartSessionSuccess(conn, msg.ID, session.ID)

	case resumeSession:
		// resume session
		log.Debugf("RESUME SESSION REQ")
		session, err := client.GetSession(msg.Session)
		if err != nil {
			handleErr(conn, msg.ID, err)
			return
		}
		handleResumeSessionSuccess(conn, msg.ID, session.GetExistingUUIDs())

	case endSession:
		// end session
		log.Debugf("END SESSION REQ")
		err := client.EndSession(context.Background(), msg.Session)
		if err != nil {
			handleErr(conn, msg.ID, err)
			return
		}
		handleEndSessionSuccess(conn, msg.ID)
	}
}

func handleStartSessionSuccess(conn *Connection, id string, session string) {
	// send error response
	err := conn.SendResponse(map[string]interface{}{
		"id":      id,
		"success": true,
		"session": session,
	})
	if err != nil {
		log.Errorf("%+v", err)
	}
}

func handleResumeSessionSuccess(conn *Connection, id string, uuids []uuid.UUID) {
	var strs []string
	for _, uid := range uuids {
		strs = append(strs, uid.String())
	}
	// send error response
	err := conn.SendResponse(map[string]interface{}{
		"id":      id,
		"success": true,
		"uuids":   strs,
	})
	if err != nil {
		log.Errorf("%+v", err)
	}
}

func handleEndSessionSuccess(conn *Connection, id string) {
	// send error response
	err := conn.SendResponse(map[string]interface{}{
		"id":      id,
		"success": true,
	})
	if err != nil {
		log.Errorf("%+v", err)
	}
}

/*
func handleRequestSync(req *pipeline.RequestInfo) ([]*pipeline.Response, error) {
	// gets an existing request or dispatch a new one
	proxy, err := pipelineService.GetOrDispatch(context.Background(), req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dispatch request")
	}

	var results []*pipeline.Response

	// process the result proxy
	for {
		select {
		case result := <-proxy.Results:
			// A session requests is single call/response, so if we already
			// have a result, process it and we're done.
			res, ok := result.(*pipeline.Response)
			if !ok {
				return nil, errors.Wrap(err, "failed to parse response")
			}
			results = append(res, results)

		case err := <-proxy.Errors:
			// handle error
			handleError(w, )
			return nil, errors.Wrap(err, "request failed")

		case <-proxy.Done:
			// finished
			return results, nil
		}
	}
}
*/
