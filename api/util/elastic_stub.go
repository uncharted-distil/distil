package util

import (
	"net/http"
	"net/http/httptest"

	"gopkg.in/olivere/elastic.v2"
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
