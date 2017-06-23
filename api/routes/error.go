package routes

import (
	"net/http"

	"github.com/unchartedsoftware/plog"
)

func handleError(w http.ResponseWriter, err error) {
	log.Errorf("%+v", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
