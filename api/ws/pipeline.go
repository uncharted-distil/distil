package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/satori/go.uuid"
	"github.com/unchartedsoftware/plog"
	"golang.org/x/net/context"

	"github.com/unchartedsoftware/distil/api/pipeline"
)

const (
	getSession = "GET_SESSION"
	endSession = "END_SESSION"
)

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
		msg, err := NewMessage(bytes)
		if err != nil {
			// parsing error, send back a failure response
			err := fmt.Errorf("unable to parse pipeline request message: %s", string(bytes))
			// send error response
			handleErr(conn, nil, err)
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
	msg.Timestamp = time.Now()
	return msg, nil
}

func handleMessage(conn *Connection, client *pipeline.Client, msg *Message) {
	switch msg.Type {
	case getSession:
		// get session

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

	case endSession:
		// end session

		err := client.EndSession(context.Background(), msg.Session)
		if err != nil {
			handleErr(conn, msg, err)
			return
		}
		handleEndSessionSuccess(conn, msg)

	default:
		// unrecognized type
		handleErr(conn, msg, fmt.Errorf("unrecognized message type"))
		return
	}
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
