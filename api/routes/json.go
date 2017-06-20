package routes

import (
	"net/http"

	"github.com/unchartedsoftware/distil/api/util/json"
)

func handleJSON(w http.ResponseWriter, data interface{}) error {
	// marshall data
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// send response
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	return nil
}
