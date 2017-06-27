package middleware

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/unchartedsoftware/distil/api/redis"
)

type redisResponseWriter struct {
	http.ResponseWriter
	key  string
	conn *redis.Conn
	code int
}

// WriteHeader write the header.
func (w *redisResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

// Write to write the response.
func (w *redisResponseWriter) Write(b []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")
	if w.code == 0 || w.code == http.StatusOK {
		// only cache if success
		bs := encodeResponse(contentType, b)
		w.conn.Set(w.key, bs)
	}
	return w.ResponseWriter.Write(b)
}

func hash(r *http.Request) string {
	buf, _ := ioutil.ReadAll(r.Body)
	// reset body with original
	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	// hash body
	return fmt.Sprintf("%s:%v:%s", r.Method, r.URL, string(buf))
}

func encodeResponse(contentType string, res []byte) []byte {
	// allocate bytes
	bs := make([]byte, 2+len(contentType)+len(res))
	// write content-type length
	binary.LittleEndian.PutUint16(
		bs[0:2],
		uint16(len(contentType)))
	// write content-type
	copy(bs[2:2+len(contentType)], []byte(contentType))
	// write response
	copy(bs[2+len(contentType):], res)
	return bs
}

func decodeResponse(bs []byte) (string, []byte) {
	contentTypeLen := binary.LittleEndian.Uint16(bs[0:2])
	contentType := string(bs[2 : 2+contentTypeLen])
	res := bs[2+contentTypeLen:]
	return contentType, res
}

func ignoreRedisCache(r *http.Request) bool {
	// check if root
	path := r.URL.EscapedPath()
	if path == "/" {
		return true
	}
	// check if file
	ext := filepath.Ext(path)
	if ext != "" {
		return true
	}
	return false
}

// Redis represents a middleware handler to support redis caching.
func Redis(pool *redis.Pool) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			// check if we should not cache
			if ignoreRedisCache(r) {
				// do not use redis
				h.ServeHTTP(w, r)
				return
			}

			// generate request key
			key := hash(r)

			// get redis connection from the pool
			conn := pool.NewConn()

			// check if it exists in redis
			exists, err := conn.Exists(key)
			if err != nil {
				// do not use redis
				h.ServeHTTP(w, r)
				return
			}

			if exists {
				// get response
				bs, err := conn.Get(key)
				if err != nil {
					// do not use redis
					h.ServeHTTP(w, r)
					return
				}
				// decode response
				contentType, res := decodeResponse(bs)
				// send response
				w.Header().Set("Content-Type", contentType)
				w.Write(res)
				// close redis conn
				conn.Close()
				return
			}

			// no response cached in redis

			// create redis response writer
			rw := &redisResponseWriter{
				ResponseWriter: w,
				conn:           conn,
				key:            key,
			}
			// forward the writer
			h.ServeHTTP(rw, r)
			conn.Close()
		}
		return http.HandlerFunc(fn)
	}
}
