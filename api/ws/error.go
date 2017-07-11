package ws

func handleErr(conn *Connection, err error) error {
	// send error response
	return conn.SendResponse(map[string]interface{}{
		"success": false,
		"error":   err,
	})
}
