package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// Write to write the gzip response.
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func isGzipSupported(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

func isWebsocketUpgrade(r *http.Request) bool {
	return r.Header.Get("Upgrade") == "websocket"
}

// Gzip represents a middleware handler to support gzip compression.
func Gzip(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !isGzipSupported(r) || isWebsocketUpgrade(r) {
			// do not use gzip
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		gzw := gzipResponseWriter{
			Writer:         gz,
			ResponseWriter: w,
		}
		h.ServeHTTP(gzw, r)
		gz.Close()
	}
	return http.HandlerFunc(fn)
}
