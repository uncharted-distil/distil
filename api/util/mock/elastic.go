package mock

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/olivere/elastic.v5"
)

// ElasticClient mocks an elasticsearch client with the provided response
// handler.
func ElasticClient(t *testing.T, handler http.HandlerFunc) *elastic.Client {
	// create a new test server to handle elasticsearch rest requests
	server := httptest.NewServer(http.HandlerFunc(handler))
	// create an elasticsearch rest client that will route requests to our test
	// server
	client, err := elastic.NewSimpleClient(elastic.SetURL(server.URL))
	assert.NoError(t, err)
	return client
}

// ElasticClientCtor mocks an elasticsearch client constructor with the provided
// response handler.
func ElasticClientCtor(t *testing.T, handler http.HandlerFunc) func() (*elastic.Client, error) {
	// create a new test server to handle elasticsearch rest requests
	server := httptest.NewServer(http.HandlerFunc(handler))
	// create an elasticsearch rest client that will route requests to our test
	// server
	client, err := elastic.NewSimpleClient(elastic.SetURL(server.URL))
	assert.NoError(t, err)
	return func() (*elastic.Client, error) {
		return client, nil
	}
}

// ElasticHandler mocks the elasticsearch response handler with the provided
// slice of response files which will be read and returned in order of
// execution.
func ElasticHandler(t *testing.T, responseFiles []string) http.HandlerFunc {
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
