package mock

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"goji.io/pattern"
)

// HTTPResponse simplifies mocking an HTTP handler function.
func HTTPResponse(t *testing.T, req *http.Request, handler http.HandlerFunc) *httptest.ResponseRecorder {
	// wrap the route under test as an http service and forward the provided
	// test request to the route http service and return the recorded results
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// HTTPRequest simplifies mockeing an HTTP request with proper parameter
// context.
func HTTPRequest(t *testing.T, method string, url string, params map[string]string) *http.Request {
	// put together a stub dataset request - need to manually account for goji's
	// parameter extraction
	req, err := http.NewRequest(method, url, nil)
	assert.NoError(t, err)
	// add params
	ctx := req.Context()
	for key, val := range params {
		ctx = context.WithValue(ctx, pattern.Variable(key), val)
	}
	// add context to req
	req = req.WithContext(ctx)
	return req
}
