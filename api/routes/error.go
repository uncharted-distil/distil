package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
)

func handleError(w http.ResponseWriter, err error) {
	log.Error(errors.Cause(err))
	http.Error(w, errors.Cause(err).Error(), http.StatusInternalServerError)
}
