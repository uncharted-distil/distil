package ws

import (
	"github.com/unchartedsoftware/plog"
)

func handleSuccess(conn *Connection, id string) {
	// send error response
	err := conn.SendResponse(map[string]interface{}{
		"id":      id,
		"success": true,
	})
	if err != nil {
		log.Errorf("%+v", err)
	}
}
