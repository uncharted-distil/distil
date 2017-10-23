package main

import (
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/unchartedsoftware/plog"
	"github.com/zenazn/goji/graceful"
	"goji.io"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/env"
	"github.com/unchartedsoftware/distil/api/middleware"
	pg "github.com/unchartedsoftware/distil/api/model/storage/postgres"
	"github.com/unchartedsoftware/distil/api/pipeline"
	"github.com/unchartedsoftware/distil/api/postgres"
	"github.com/unchartedsoftware/distil/api/redis"
	"github.com/unchartedsoftware/distil/api/routes"
	"github.com/unchartedsoftware/distil/api/ws"
)

var (
	version   = "unset"
	timestamp = "unset"
)

func registerRoute(mux *goji.Mux, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	log.Infof("Registering GET route %s", pattern)
	mux.HandleFunc(pat.Get(pattern), handler)
}

func registerRoutePost(mux *goji.Mux, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	log.Infof("Registering POST route %s", pattern)
	mux.HandleFunc(pat.Post(pattern), handler)
}

func main() {
	log.Infof("version: %s built: %s", version, timestamp)

	// load config from env
	config, err := env.LoadConfig()
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}

	// instantiate elastic client constructor.
	esClientCtor := elastic.NewClient(config.ElasticEndpoint, false)

	// instantiate the postgres client constructor.
	postgresClientCtor := postgres.NewClient(config.PostgresHost, config.PostgresPort, config.PostgresUser, config.PostgresPassword, config.PostgresDatabase)

	// make sure a connection can be made to postgres - doesn't appear to be thread safe and
	// causes panic if deferred, so we'll do it an a retry loop here.  We need to provide
	// flexibility on startup because we can't guarantee the DB will be up before the server.
	for i := 0; i < config.PostgresRetryCount; i++ {
		_, err = postgresClientCtor()
		if err == nil {
			break
		} else if i == config.PostgresRetryCount {
			log.Errorf("%v", err)
			os.Exit(1)
		}
		log.Errorf("%v", err)
		time.Sleep(time.Duration(config.PostgresRetryTimeout) * time.Millisecond)
	}

	// instantiate the postgres storage constructor.
	pgStorageCtor := pg.NewStorage(postgresClientCtor, esClientCtor)

	// Instantiate the pipeline compute client
	pipelineClient, err := pipeline.NewClient(config.PipelineComputeEndpoint, config.PipelineDataDir, config.PipelineComputeTrace)
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
	defer pipelineClient.Close()

	// instantiate redis pool
	redisPool := redis.NewPool(config.RedisEndpoint, config.RedisExpiry)

	// register routes
	mux := goji.NewMux()

	mux.Use(middleware.Log)
	mux.Use(middleware.Gzip)
	mux.Use(middleware.Redis(redisPool))

	registerRoute(mux, "/distil/datasets/:index", routes.DatasetsHandler(esClientCtor))
	registerRoute(mux, "/distil/variables/:index/:dataset", routes.VariablesHandler(esClientCtor))
	registerRoutePost(mux, "/distil/variables/:index/:dataset/update", routes.VariableTypeHandler(pgStorageCtor, esClientCtor))
	registerRoute(mux, "/distil/variable-summaries/:index/:dataset/:variable", routes.VariableSummaryHandler(pgStorageCtor, esClientCtor))
	registerRoute(mux, "/distil/filtered-data/:esIndex/:dataset/:inclusive", routes.FilteredDataHandler(pgStorageCtor))
	registerRoute(mux, "/distil/results/:index/:dataset/:results-uuid/:inclusive", routes.ResultsHandler(pgStorageCtor))
	registerRoute(mux, "/distil/results-summary/:index/:dataset/:results-uuid", routes.ResultsSummaryHandler(pgStorageCtor))
	registerRoute(mux, "/distil/session/:session", routes.SessionHandler(pgStorageCtor))
	registerRoute(mux, "/distil/abort", routes.AbortHandler())
	registerRoute(mux, "/distil/export/:session/:pipeline-id", routes.ExportHandler(pipelineClient, config.ExportPath))

	registerRoute(mux, "/ws", ws.PipelineHandler(pipelineClient, esClientCtor, pgStorageCtor))
	registerRoute(mux, "/*", routes.FileHandler("./dist"))

	// catch kill signals for graceful shutdown
	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)

	// kick off the server listen loop
	log.Infof("Listening on port %s", config.AppPort)
	err = graceful.ListenAndServe(":"+config.AppPort, mux)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// wait until server gracefully exits
	graceful.Wait()
}
