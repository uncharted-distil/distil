package main

import (
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/unchartedsoftware/plog"
	"github.com/zenazn/goji/graceful"
	"goji.io"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil-ingest/rest"
	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/env"
	"github.com/unchartedsoftware/distil/api/middleware"
	es "github.com/unchartedsoftware/distil/api/model/storage/elastic"
	pg "github.com/unchartedsoftware/distil/api/model/storage/postgres"
	"github.com/unchartedsoftware/distil/api/pipeline"
	"github.com/unchartedsoftware/distil/api/postgres"
	"github.com/unchartedsoftware/distil/api/routes"
	"github.com/unchartedsoftware/distil/api/task"
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
		log.Errorf("%+v", err)
		os.Exit(1)
	}
	log.Infof("%+v", spew.Sdump(config))

	// instantiate elastic client constructor.
	esClientCtor := elastic.NewClient(config.ElasticEndpoint, false)

	// instantiate the postgres client constructor.
	postgresClientCtor := postgres.NewClient(config.PostgresHost, config.PostgresPort, config.PostgresUser, config.PostgresPassword,
		config.PostgresDatabase, config.PostgresLogLevel)

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

	// instantiate the metadata storage (using ES).
	metadataStorageCtor := es.NewMetadataStorage(esClientCtor)

	// instantiate the postgres data storage constructor.
	pgDataStorageCtor := pg.NewDataStorage(postgresClientCtor, metadataStorageCtor)

	// instantiate the postgres pipeline storage constructor.
	pgPipelineStorageCtor := pg.NewPipelineStorage(postgresClientCtor, metadataStorageCtor)

	// Instantiate the pipeline compute client
	pipelineClient, err := pipeline.NewClient(config.PipelineComputeEndpoint, config.PipelineDataDir, config.PipelineComputeTrace)
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
	defer pipelineClient.Close()

	// instantiate the REST client for primitives.
	restClient := rest.NewClient(config.PrimitiveEndPoint)

	// build the ingest configuration.
	ingestConfig := &task.IngestTaskConfig{
		ContainerDataPath:                config.DataFolderPath,
		TmpDataPath:                      config.TmpDataPath,
		DataPathRelative:                 config.DataFilePath,
		DatasetFolderSuffix:              config.DatasetFolderSuffix,
		HasHeader:                        true,
		MergedOutputPathRelative:         config.MergedOutputDataPath,
		MergedOutputSchemaPathRelative:   config.MergedOutputSchemaPath,
		SchemaPathRelative:               config.SchemaPath,
		ClassificationRESTEndpoint:       config.ClassificationEndpoint,
		ClassificationFunctionName:       config.ClassificationFunctionName,
		ClassificationOutputPathRelative: config.ClassificationOutputPath,
		RankingRESTEndpoint:              config.RankingEndpoint,
		RankingFunctionName:              config.RankingFunctionName,
		RankingOutputPathRelative:        config.RankingOutputPath,
		RankingRowLimit:                  config.RankingRowLimit,
		DatabasePassword:                 config.PostgresPassword,
		DatabaseUser:                     config.PostgresUser,
		Database:                         config.PostgresDatabase,
		SummaryOutputPathRelative:        config.SummaryPath,
		SummaryRESTEndpoint:              config.SummaryEndpoint,
		SummaryFunctionName:              config.SummaryFunctionName,
		SummaryMachineOutputPathRelative: config.SummaryMachinePath,
		ESEndpoint:                       config.ElasticEndpoint,
		ESTimeout:                        config.ElasticTimeout,
		ESDatasetPrefix:                  config.ElasticDatasetPrefix,
	}
	waitForEndpoints(config)

	// Ingest the data specified by the environment
	if config.InitialDataset != "" && !config.SkipIngest {
		log.Infof("Loading initial dataset '%s'", config.InitialDataset)
		err = task.IngestDataset(metadataStorageCtor, config.ESDatasetsIndex, config.InitialDataset, ingestConfig)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}
	// register routes
	mux := goji.NewMux()
	mux.Use(middleware.Log)
	mux.Use(middleware.Gzip)

	registerRoute(mux, "/distil/datasets/:index", routes.DatasetsHandler(metadataStorageCtor))
	registerRoute(mux, "/distil/variables/:index/:dataset", routes.VariablesHandler(metadataStorageCtor))
	registerRoutePost(mux, "/distil/variables/:index/:dataset", routes.VariableTypeHandler(pgDataStorageCtor, metadataStorageCtor))
	registerRoutePost(mux, "/distil/discovery/:index/:dataset/:target", routes.ProblemDiscoveryHandler(pgDataStorageCtor, metadataStorageCtor, config.UserProblemPath))
	registerRoute(mux, "/distil/variable-summaries/:index/:dataset/:variable", routes.VariableSummaryHandler(pgDataStorageCtor))
	registerRoute(mux, "/distil/filtered-data/:esIndex/:dataset/:inclusive", routes.FilteredDataHandler(pgDataStorageCtor))
	registerRoute(mux, "/distil/results/:index/:dataset/:pipeline-id/:inclusive", routes.ResultsHandler(pgPipelineStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/results-summary/:index/:dataset/:results-uuid", routes.ResultsSummaryHandler(pgPipelineStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/results-variable-summary/:index/:dataset/:variable/:results-uuid", routes.ResultVariableSummaryHandler(pgPipelineStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/residuals-summary/:index/:dataset/:results-uuid", routes.ResidualsSummaryHandler(pgPipelineStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/ranking/:index/:dataset/:target", routes.RankingHandler(pgDataStorageCtor, restClient, config.PipelineDataDir))
	registerRoute(mux, "/distil/session/:session/:dataset/:target/:pipeline-id", routes.SessionHandler(pgPipelineStorageCtor))
	registerRoute(mux, "/distil/abort", routes.AbortHandler())
	registerRoute(mux, "/distil/export/:session/:pipeline-id", routes.ExportHandler(pgPipelineStorageCtor, metadataStorageCtor, pipelineClient, config.ExportPath))
	registerRoute(mux, "/distil/ingest/:index/:dataset", routes.IngestHandler(metadataStorageCtor, ingestConfig))
	registerRoute(mux, "/ws", ws.PipelineHandler(pipelineClient, metadataStorageCtor, pgDataStorageCtor, pgPipelineStorageCtor))

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

func waitForEndpoints(config env.Config) {
	log.Info("Waiting for services as needed")
	if config.ClassificationWait {
		log.Infof("Waiting for classification service at %s", config.ClassificationEndpoint)
		waitForPostEndpoint(config.ClassificationEndpoint, config.ServiceRetryCount)
		log.Infof("Classification service is up")
	}

	if config.RankingWait {
		log.Infof("Waiting for ranking service at %s", config.RankingEndpoint)
		waitForPostEndpoint(config.RankingEndpoint, config.ServiceRetryCount)
		log.Infof("Ranking service is up")
	}
	log.Info("All required services are up")
}

func waitForPostEndpoint(endpoint string, retryCount int) {
	up := false
	i := 0
	for ; i < retryCount && !up; i++ {
		resp, err := http.Post(endpoint, "application/json", strings.NewReader("test"))
		if err != nil {
			log.Infof("Sent request to %s", endpoint)

			// If the error indicates the service is up, then stop waiting.
			if !strings.Contains(err.Error(), "connection refused") {
				up = true
			}
			time.Sleep(10 * time.Second)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	if i == retryCount {
		log.Errorf("Shutting down since unable to connect to %s after %d retries", endpoint, retryCount)
		os.Exit(1)
	}
}
