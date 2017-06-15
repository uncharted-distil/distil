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
	"github.com/unchartedsoftware/distil/api/routes"
)

const (
	defaultEsEndpoint = "http://localhost:9200"
	defaultAppPort    = "8080"
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
	port := env.Load("PORT", defaultAppPort)

	// instantiate elasticsearch client
	client, err := elastic.NewClient(esEndpoint, false)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// register routes
	mux := goji.NewMux()
	registerRoute(mux, "/distil/echo/:echo", routes.EchoHandler())
	registerRoute(mux, "/distil/datasets/:index", routes.DatasetsHandler(client))
	registerRoute(mux, "/distil/variables/:index/:dataset", routes.VariablesHandler(client))
	registerRoute(mux, "/distil/variable-summaries/:index/:dataset", routes.VariableSummariesHandler(client))
	registerRoute(mux, "/*", routes.FileHandler("./dist"))

	// catch kill signals for graceful shutdown
	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)

	// kick off the server listen loop
	log.Infof("Listening on port %s", port)
	err = graceful.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// wait until server gracefully exits
	graceful.Wait()
}
