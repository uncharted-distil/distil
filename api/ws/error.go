package ws

import (
	"github.com/unchartedsoftware/plog"
)

func handleErr(conn *Connection, id string, err error) {
	log.Errorf("%+v", err)
	// send error response
	err = conn.SendResponse(map[string]interface{}{
		"id":    id,
		"error": err,
	})
	if err != nil {
		log.Errorf("%+v", err)
	}
}
