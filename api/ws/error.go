package ws

import (
	"time"

	"github.com/unchartedsoftware/plog"
)

func handleErr(conn *Connection, msg *Message, err error) {
	if msg != nil {
		// log the response
		logger := newMessageLogger()
		logger.messageType(msg.Type)
		logger.message(msg.Raw)
		logger.duration(time.Now().Sub(msg.Timestamp))
		logger.log(err != nil)
		// send error response if we have an id
		errOther := conn.SendResponse(map[string]interface{}{
			"id":      msg.ID,
			"success": false,
			"error":   err,
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
