package ws

import (
	"time"

	"github.com/unchartedsoftware/plog"
)

func handleSuccess(conn *Connection, msg *Message, response map[string]interface{}) {
	// append msg id
	response["id"] = msg.ID
	// log the response
	logger := newMessageLogger()
	logger.messageType(msg.Type)
	logger.message(msg.Raw)
	logger.duration(time.Now().Sub(msg.Timestamp))
	logger.log(true)
	// send response
	err := conn.SendResponse(response)
	if err != nil {
		log.Errorf("%+v", err)
	}
}
