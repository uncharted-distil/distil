package main

import (
	"net/http"
	"os"
	"syscall"

	"github.com/unchartedsoftware/plog"
	"github.com/zenazn/goji/graceful"
	"goji.io"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/env"
	"github.com/unchartedsoftware/distil/api/middleware"
	"github.com/unchartedsoftware/distil/api/pipeline"
	"github.com/unchartedsoftware/distil/api/redis"
	"github.com/unchartedsoftware/distil/api/routes"
	"github.com/unchartedsoftware/distil/api/ws"
)

const (
	defaultEsEndpoint              = "http://localhost:9200"
	defaultRedisEndpoint           = "localhost:6379"
	defaultRedisExpiry             = -1 // no expiry
	defaultAppPort                 = "8080"
	defaultPipelineComputeEndPoint = "localhost:9500"
)

var (
	version   = "unset"
	timestamp = "unset"
)

func registerRoute(mux *goji.Mux, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	log.Infof("Registering route %s", pattern)
	mux.HandleFunc(pat.Get(pattern), handler)
}

func main() {
	log.Infof("version: %s built: %s", version, timestamp)

	// load elasticsearch endpoint
	esEndpoint := env.Load("ES_ENDPOINT", defaultEsEndpoint)
	// load application port
	redisEndpoint := env.Load("REDIS_ENDPOINT", defaultRedisEndpoint)
	// load redis endpoint
	httpPort := env.Load("PORT", defaultAppPort)
	// load compute server endpoint
	pipelineComputeEndpoint := env.Load("PIPELINE_COMPUTE_ENDPOINT", defaultPipelineComputeEndPoint)

	// instantiate elasticsearch client constructor
	esClientCtor := elastic.NewClient(esEndpoint, false)

	// instantiate the pipeline compute client
	pipelineClient, err := pipeline.NewClient(pipelineComputeEndpoint)
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
	defer pipelineClient.Close()

	// instantiate redis pool
	redisPool := redis.NewPool(redisEndpoint, defaultRedisExpiry)

	// register routes
	mux := goji.NewMux()

	mux.Use(middleware.Log)
	mux.Use(middleware.Gzip)
	mux.Use(middleware.Redis(redisPool))

	registerRoute(mux, "/distil/datasets/:index", routes.DatasetsHandler(esClientCtor))
	registerRoute(mux, "/distil/variables/:index/:dataset", routes.VariablesHandler(esClientCtor))
	registerRoute(mux, "/distil/variable-summaries/:index/:dataset/:variable", routes.VariableSummaryHandler(esClientCtor))
	registerRoute(mux, "/distil/filtered-data/:dataset", routes.FilteredDataHandler(esClientCtor))
	registerRoute(mux, "/distil/pipeline-session-id", routes.PipelineSessionHandler(pipelineClient))
	registerRoute(mux, "/distil/pipeline-end-session/:session-id", routes.PipelineEndSessionHandler(pipelineClient))
	registerRoute(mux, "/distil/pipeline-create/:session-id/:task/:output/:metric/:max-pipelines", routes.PipelineCreateHandler(pipelineClient))
	registerRoute(mux, "/distil/pipeline-execute/:session-id/:pipeline-id", routes.PipelineExecuteHandler(pipelineClient))
	registerRoute(mux, "/ws", ws.PipelineHandler(pipelineClient))
	registerRoute(mux, "/*", routes.FileHandler("./dist"))

	// catch kill signals for graceful shutdown
	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)

	// kick off the server listen loop
	log.Infof("Listening on port %s", httpPort)
	err = graceful.ListenAndServe(":"+httpPort, mux)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// wait until server gracefully exits
	graceful.Wait()
}
