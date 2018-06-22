package ws

import (
	"time"

	"github.com/unchartedsoftware/plog"
)

func handleErr(conn *Connection, msg *Message, err error) {
	if msg != nil {
		// log the response
		newMessageLogger().
			messageType(msg.Type).
			message(msg.Raw).
			duration(time.Since(msg.Timestamp)).
			log(err != nil)
		// send error response if we have an id
		errOther := conn.SendResponse(map[string]interface{}{
			"id":      msg.ID,
			"success": false,
			"error":   err.Error(),
		})
		// log error
		if errOther != nil {
			log.Errorf("%+v", errOther)
		}
	}
	// log error
	if err != nil {
		log.Errorf("%+v", err)
	}
}
