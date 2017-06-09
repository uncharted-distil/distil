package main

import (
	"net/http"
	"os"
	"syscall"
	"time"

	elastic "gopkg.in/olivere/elastic.v2"

	"github.com/unchartedsoftware/plog"
	"github.com/zenazn/goji/graceful"

	"github.com/unchartedsoftware/distil-server/routes"

	"strconv"

	"goji.io"
	"goji.io/pat"
)

const (
	defaultEsEndpoint = "http://localhost:9200"
	defaultAppPort    = 8080
	defaultEsTimeout  = time.Second * 60 * 5
)

var (
	dataSetClient *elastic.Client
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

func registerRoute(pattern string, handler func(http.ResponseWriter, *http.Request), mux *goji.Mux) {
	log.Infof("Registering route %s", pattern)
	mux.HandleFunc(pat.Get(pattern), handler)
}

func main() {

	var err error

	// Creates an endpoint for locally managed ES data
	datasetEndpoint := getEnv("DATASET_ENDPOINT", defaultEsEndpoint)
	dataSetClient, err = createEsClient(datasetEndpoint)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// route registration
	mux := goji.NewMux()
	registerRoute("/distil/echo/:echo", routes.EchoHandler(), mux)
	registerRoute("/distil/datasets", routes.DatasetsHandler(dataSetClient), mux)
	registerRoute("/distil/variables/:dataset", routes.VariablesHandler(dataSetClient), mux)
	registerRoute("/*", routes.FileHandler("./dist"), mux)

	// catch kill signals for graceful shutdown
	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)

	// kick off the server listen loop
	err = graceful.ListenAndServe(":"+strconv.Itoa(defaultAppPort), mux)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// wait until server gracefully exits
	graceful.Wait()
}
