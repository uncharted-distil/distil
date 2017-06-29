package mock

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"goji.io/pattern"
)

// HTTPResponse simplifies mocking a HTTP handler function.
func HTTPResponse(t *testing.T, req *http.Request, handler http.HandlerFunc) *httptest.ResponseRecorder {
	// wrap the route under test as an http service and forward the provided
	// test request to the route http service and return the recorded results
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// HTTPRequest simplifies mocking a HTTP request.
func HTTPRequest(t *testing.T, method string, url string, params map[string]string, query map[string]string) *http.Request {
	// create query params
	var queryParams []string
	for key, val := range query {
		queryParams = append(queryParams, key+"="+val)
	}
	if len(queryParams) > 0 {
		url += "?" + strings.Join(queryParams, "&")
	}
	// create request
	req, err := http.NewRequest(method, url, nil)
	assert.NoError(t, err)
	// add params
	ctx := req.Context()
	for key, val := range params {
		ctx = context.WithValue(ctx, pattern.Variable(key), val)
	}
	// add context to req
	return req.WithContext(ctx)
}
