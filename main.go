package main

import (
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"github.com/zenazn/goji/graceful"
	"goji.io"
	"goji.io/pat"
	"gopkg.in/olivere/elastic.v2"

	"github.com/unchartedsoftware/distil/routes"
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

// Wraps calls to plog in the elastic.Logger interface
type elasticPlogAdapter struct{}

func (elasticPlogAdapter) Printf(format string, v ...interface{}) {
	log.Infof(format, v)
}

func createEsClient(endpoint string, debug bool) (*elastic.Client, error) {
	// turn on trace logs if necessary
	var adapter *elasticPlogAdapter
	if debug {
		adapter = new(elasticPlogAdapter)
	}

	client, err := elastic.NewClient(
		elastic.SetURL(endpoint),
		elastic.SetHttpClient(&http.Client{Timeout: defaultEsTimeout}),
		elastic.SetMaxRetries(10),
		elastic.SetSniff(false),
		elastic.SetGzip(false),
		elastic.SetTraceLog(adapter))
	if err != nil {
		return nil, errors.Wrap(err, "ES client init failed")
	}

	log.Infof("Connected to endpoint %s", endpoint)
	return client, nil
}

func registerRoute(pattern string, handler func(http.ResponseWriter, *http.Request), mux *goji.Mux) {
	log.Infof("Registering route %s", pattern)
	mux.HandleFunc(pat.Get(pattern), handler)
}

func main() {
	// Creates an endpoint for locally managed ES data
	esEndpoint := getEnv("DATASET_ENDPOINT", defaultEsEndpoint)
	dataSetClient, err := createEsClient(esEndpoint, true)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// route registration
	mux := goji.NewMux()
	registerRoute("/distil/echo/:echo", routes.EchoHandler(), mux)
	registerRoute("/distil/datasets", routes.DatasetsHandler(dataSetClient), mux)
	registerRoute("/distil/variables/:dataset", routes.VariablesHandler(dataSetClient), mux)
	registerRoute("/distil/variable-summaries/:dataset", routes.VariableSummariesHandler(dataSetClient), mux)
	registerRoute("/*", routes.FileHandler("./dist"), mux)

	// catch kill signals for graceful shutdown
	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)

	// kick off the server listen loop
	log.Infof("Listening on port %d", defaultAppPort)
	err = graceful.ListenAndServe(":"+strconv.Itoa(defaultAppPort), mux)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// wait until server gracefully exits
	graceful.Wait()
}
