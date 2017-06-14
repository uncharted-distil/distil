package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
)

func handleServerError(err error, w http.ResponseWriter) {
	log.Error(errors.Cause(err))
	http.Error(w, errors.Cause(err).Error(), http.StatusInternalServerError)
}
