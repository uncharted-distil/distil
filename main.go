package main

import (
	"net/http"
	"os"
	"time"

	elastic "gopkg.in/olivere/elastic.v2"

	"github.com/unchartedsoftware/plog"

	"github.com/unchartedsoftware/distil-server/routes"

	"goji.io"
	"goji.io/pat"
)

const (
	defaultEsEndpoint = "http://localhost:9200"
	defaultEsTimeout  = time.Second * 60 * 5
)

var (
	dataSetClient *elastic.Client
	marvinClient  *elastic.Client
)

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		val = fallback
	}
	log.Infof("%s = %s", key, val)
	return val
}

func createEsClient(endpoint string) (*elastic.Client, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(endpoint),
		elastic.SetHttpClient(&http.Client{Timeout: defaultEsTimeout}),
		elastic.SetMaxRetries(10),
		elastic.SetSniff(false),
		elastic.SetGzip(false))
	if err != nil {
		return nil, err
	}
	log.Infof("Connected to endpoint %s", endpoint)
	return client, nil
}

func registerRoute(pattern string, handler routes.Route, mux *goji.Mux) {
	log.Infof("Registering route %s", pattern)
	mux.HandleFunc(pat.Get(pattern), handler)
}

func main() {

	var err error

	// Creates an endpoint for locally managed ES data
	datasetEndpoint := getEnv("DATASET_ENDPOINT", defaultEsEndpoint)
	dataSetClient, err = createEsClient(datasetEndpoint)
	if err != nil {
		log.Error("Failed to initialize dataset client")
		log.Error(err)
		os.Exit(1)
	}

	// Creates an ES endpoint for shared D3M ES data
	marvinEndpoint := getEnv("MARVIN_ENDPOINT", defaultEsEndpoint)
	marvinClient, err = createEsClient(marvinEndpoint)
	if err != nil {
		log.Error("Failed to initialize marvin client")
		log.Error(err)
		os.Exit(1)
	}

	// route registration
	mux := goji.NewMux()
	registerRoute("/distil/echo/:echo", routes.EchoHandler(), mux)
	registerRoute("/distil/datasets", routes.DatasetsHandler(marvinClient), mux)
	registerRoute("/distil/variables/:dataset", routes.VariablesHandler(marvinClient), mux)

	// kick off the serer listen loop
	http.ListenAndServe("localhost:8000", mux)
}
