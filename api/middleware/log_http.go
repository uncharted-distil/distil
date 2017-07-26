package middleware

import (
	"net/http"
	"time"
)

// Log is a middleware that logs each request. Inspired from the logging
// middleware from zenazn/goji:
// https://github.com/zenazn/goji/blob/master/web/middleware/logger.go
func Log(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if isWebsocketUpgrade(r) {
			// do not log websocket connections
			// TODO: intercept and log the beginning and the end of the
			// connection.
			h.ServeHTTP(w, r)
			return
		}
		lw := wrapWriter(w)
		t1 := time.Now()
		h.ServeHTTP(lw, r)
		if lw.Status() == 0 {
			lw.WriteHeader(http.StatusOK)
		}
		t2 := time.Now()
		newRequestLogger().
			requestType(r.Method).
			request(r.URL.String()).
			params(r.URL.String()).
			status(lw.Status()).
			duration(t2.Sub(t1)).
			log(lw.Status() < 500)
	}
	return http.HandlerFunc(fn)
}
