package mock

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// HTTPResponse simplifies mocking a HTTP handler function.
func HTTPResponse(t *testing.T, req *http.Request, handler http.HandlerFunc) *httptest.ResponseRecorder {
	// wrap the route under test as an http service and forward the provided
	// test request to the route http service and return the recorded results
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
