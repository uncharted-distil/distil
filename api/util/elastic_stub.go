package util

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/olivere/elastic.v3"
)

// ElasticStub represents a handler function to mock the elasticsearch response.
type ElasticStub func(w http.ResponseWriter, r *http.Request)

// ElasticHandler represents the elasticsearch based handler to test.
type ElasticHandler func(*elastic.Client) func(http.ResponseWriter, *http.Request)

// TestElasticRoute allows for mocking an elasticsearch based handler function.
func TestElasticRoute(
	esHandler ElasticStub,
	testRequest *http.Request,
	testRoute ElasticHandler) (*httptest.ResponseRecorder, error) {
	// create a new test server to handle elastic search rest requests
	testServer := httptest.NewServer(http.HandlerFunc(esHandler))
	// create an elastic search rest client that will route requests to our test
	// server
	client, err := elastic.NewSimpleClient(elastic.SetURL(testServer.URL))
	if err != nil {
		return nil, err
	}
	// wrap the route under test as an http service
	routeHandler := http.HandlerFunc(testRoute(client))
	// forward the caller supplied test request to the route http service and
	// return the recorded results
	responseRec := httptest.NewRecorder()
	routeHandler.ServeHTTP(responseRec, testRequest)
	return responseRec, nil
}

// MockElasticResponse mocks the elasticsearch response handler with the
// provided slice of responses returned in order.
func MockElasticResponse(t *testing.T, responseFiles []string) ElasticStub {
	c := 0
	// mock elasticsearch request handler
	return func(w http.ResponseWriter, r *http.Request) {
		if c < len(responseFiles) {
			// load ES result json the test server will return
			res, err := ioutil.ReadFile(responseFiles[c])
			assert.NoError(t, err)
			w.Write(res)
			c++
			return
		}
		t.Errorf("no response file found for index %d", c)
	}

}
