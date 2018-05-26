package ws

import (
	"time"

	"github.com/unchartedsoftware/plog"
)

func handleComplete(conn *Connection, msg *Message) {
	// append msg id
	response := map[string]interface{}{
		"id":       msg.ID,
		"complete": true,
	}
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
