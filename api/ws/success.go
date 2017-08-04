package ws

import (
	"time"

	"github.com/unchartedsoftware/plog"
)

func handleSuccess(conn *Connection, msg *Message, response map[string]interface{}) {
	// append msg id
	response["id"] = msg.ID
	// log the response
	newMessageLogger().
		messageType(msg.Type).
		message(msg.Raw).
		duration(time.Since(msg.Timestamp)).
		log(true)
	// send response
	err := conn.SendResponse(response)
	if err != nil {
		log.Errorf("%+v", err)
	}
}
