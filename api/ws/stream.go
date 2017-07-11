package ws

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/unchartedsoftware/plog"
)

const (
	// MetaRoute represents the HTTP route for the resource.
	MetaRoute = "/ws/meta"
)

// StreamHandler represents the HTTP route response handler.
func StreamHandler(w http.ResponseWriter, r *http.Request) {
	// create conn
	conn, err := NewConnection(w, r, handleStream)
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

func parseRequestJSON(bytes []byte) (map[string]interface{}, error) {
	var req map[string]interface{}
	err := json.Unmarshal(bytes, &req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func handleStream(conn *Connection, msg []byte) {
	// parse the meta request into JSON
	_, err := parseRequestJSON(msg)
	if err != nil {
		// parsing error, send back a failure response
		err := fmt.Errorf("unable to parse meta request message: %s", string(msg))
		// log error
		log.Warn(err)
		// send error response
		err = handleErr(conn, err)
		if err != nil {
			log.Warn(err)
		}
		return
	}

	// TODO: send initial request here
	err = conn.SendResponse([]byte("setup"))
	if err != nil {
		log.Warn(err)
	}

	// TODO: hook up here to external API
	for {
		// TODO: read update from external API
		// ...

		// TODO: write update to client
		err = conn.SendResponse([]byte("update message"))
		if err != nil {
			log.Warn(err)
		}
		break
	}

	// TODO: shutdown message to client
	err = conn.SendResponse([]byte("finish"))
	if err != nil {
		log.Warn(err)
	}
}
