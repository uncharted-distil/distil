//
//   Copyright © 2021 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

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

func isImage(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "image")
}

func isWebsocketUpgrade(r *http.Request) bool {
	return r.Header.Get("Upgrade") == "websocket"
}

// Gzip represents a middleware handler to support gzip compression.
func Gzip(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !isGzipSupported(r) || isWebsocketUpgrade(r) || isImage(r) {
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
