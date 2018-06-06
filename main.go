package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/davecgh/go-spew/spew"
	"github.com/unchartedsoftware/plog"
	"github.com/zenazn/goji/graceful"
	"goji.io"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil-ingest/rest"
	"github.com/unchartedsoftware/distil/api/compute"
	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/env"
	"github.com/unchartedsoftware/distil/api/middleware"
	es "github.com/unchartedsoftware/distil/api/model/storage/elastic"
	pg "github.com/unchartedsoftware/distil/api/model/storage/postgres"
	"github.com/unchartedsoftware/distil/api/postgres"
	"github.com/unchartedsoftware/distil/api/routes"
	"github.com/unchartedsoftware/distil/api/service"
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
	servicesToWait := make(map[string]service.Heartbeat)

	userAgent := fmt.Sprintf("uncharted-distil-%s-%s", version, timestamp)
	apiVersion := compute.GetAPIVersion()
	log.Infof("user agent: %s api version: %s", userAgent, apiVersion)

	// load config from env
	config, err := env.LoadConfig()
	if err != nil {
		log.Errorf("%+v", err)
		os.Exit(1)
	}
	log.Infof("%+v", spew.Sdump(config))

	// set dataset directory
	compute.SetDatasetDir(config.SolutionDataDir)

	// instantiate elastic client constructor.
	esClientCtor := elastic.NewClient(config.ElasticEndpoint, false)

	// instantiate the postgres client constructor.
	postgresClientCtor := postgres.NewClient(config.PostgresHost, config.PostgresPort, config.PostgresUser, config.PostgresPassword,
		config.PostgresDatabase, config.PostgresLogLevel)

	// wait for required services.
	servicesToWait["postgres"] = func() bool {
		_, err := postgresClientCtor()
		return err == nil
	}
	servicesToWait["elastic"] = func() bool {
		_, err := esClientCtor()
		return err == nil
	}
	if config.ClassificationWait {
		servicesToWait["classification"] = func() bool {
			return waitForPostEndpoint(fmt.Sprintf("%s%s", config.ClassificationEndpoint, "/aaaa"))
		}
	}
	if config.RankingWait {
		servicesToWait["ranking"] = func() bool {
			return waitForPostEndpoint(fmt.Sprintf("%s%s", config.RankingEndpoint, "/aaaa"))
		}
	}

	// set the ingest functions to use
	if config.IngestPrimitive {
		task.SetClassify(task.ClassifyPrimmitive)
		task.SetRank(task.RankPrimmitive)
		task.SetSummarize(task.SummarizePrimitive)
		task.SetFeaturize(task.FeaturizePrimitive)
	}

	// make sure a connection can be made to postgres - doesn't appear to be thread safe and
	// causes panic if deferred, so we'll do it an a retry loop here.  We need to provide
	// flexibility on startup because we can't guarantee the DB will be up before the server.
	for name, test := range servicesToWait {
		log.Infof("Waiting for service '%s'", name)
		err = service.WaitForService(name, &config, test)
		if err == nil {
			log.Infof("Service '%s' is up", name)
		} else {
			log.Error(err)
			os.Exit(1)
		}
	}

	// instantiate the metadata storage (using ES).
	metadataStorageCtor := es.NewMetadataStorage(config.ESDatasetsIndex, esClientCtor)

	// instantiate the postgres data storage constructor.
	pgDataStorageCtor := pg.NewDataStorage(postgresClientCtor, metadataStorageCtor)

	// instantiate the postgres solution storage constructor.
	pgSolutionStorageCtor := pg.NewSolutionStorage(postgresClientCtor, metadataStorageCtor)

	// Instantiate the solution compute client
	solutionClient, err := compute.NewClient(config.SolutionComputeEndpoint, config.SolutionDataDir, config.SolutionComputeTrace, userAgent)
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
	defer solutionClient.Close()

	// instantiate the REST client for primitives.
	restClient := rest.NewClient(config.PrimitiveEndPoint)

	// build the ingest configuration.
	ingestConfig := &task.IngestTaskConfig{
		ContainerDataPath:                  config.DataFolderPath,
		TmpDataPath:                        config.TmpDataPath,
		DataPathRelative:                   config.DataFilePath,
		DatasetFolderSuffix:                config.DatasetFolderSuffix,
		MediaPath:                          config.MediaPath,
		HasHeader:                          true,
		FeaturizationRESTEndpoint:          config.FeaturizationRESTEndpoint,
		FeaturizationFunctionName:          config.FeaturizationFunctionName,
		FeaturizationOutputDataRelative:    config.FeaturizationOutputDataRelative,
		FeaturizationOutputSchemaRelative:  config.FeaturizationOutputSchemaRelative,
		MergedOutputPathRelative:           config.MergedOutputDataPath,
		MergedOutputSchemaPathRelative:     config.MergedOutputSchemaPath,
		SchemaPathRelative:                 config.SchemaPath,
		ClassificationRESTEndpoint:         config.ClassificationEndpoint,
		ClassificationFunctionName:         config.ClassificationFunctionName,
		ClassificationOutputPathRelative:   config.ClassificationOutputPath,
		ClassificationProbabilityThreshold: config.ClassificationProbabilityThreshold,
		RankingRESTEndpoint:                config.RankingEndpoint,
		RankingFunctionName:                config.RankingFunctionName,
		RankingOutputPathRelative:          config.RankingOutputPath,
		RankingRowLimit:                    config.RankingRowLimit,
		DatabasePassword:                   config.PostgresPassword,
		DatabaseUser:                       config.PostgresUser,
		Database:                           config.PostgresDatabase,
		DatabaseHost:                       config.PostgresHost,
		DatabasePort:                       config.PostgresPort,
		SummaryOutputPathRelative:          config.SummaryPath,
		SummaryRESTEndpoint:                config.SummaryEndpoint,
		SummaryFunctionName:                config.SummaryFunctionName,
		SummaryMachineOutputPathRelative:   config.SummaryMachinePath,
		ESEndpoint:                         config.ElasticEndpoint,
		ESTimeout:                          config.ElasticTimeout,
		ESDatasetPrefix:                    config.ElasticDatasetPrefix,
	}

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

	routes.SetVerboseError(config.VerboseError)

	// GET
	registerRoute(mux, "/distil/datasets/:index", routes.DatasetsHandler(metadataStorageCtor))
	registerRoute(mux, "/distil/solutions/:dataset/:target/:solution-id", routes.SolutionHandler(pgSolutionStorageCtor))
	registerRoute(mux, "/distil/variables/:index/:dataset", routes.VariablesHandler(metadataStorageCtor))
	registerRoute(mux, "/distil/results-variable-extrema/:index/:dataset/:variable/:results-uuid", routes.ResultVariableExtremaHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/results-extrema/:index/:dataset/:results-uuid", routes.ResultsExtremaHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/residuals-extrema/:index/:dataset/:results-uuid", routes.ResidualsExtremaHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/ranking/:index/:dataset/:target", routes.RankingHandler(pgDataStorageCtor, restClient, config.SolutionDataDir))
	registerRoute(mux, "/distil/abort", routes.AbortHandler())
	registerRoute(mux, "/distil/export/:solution-id", routes.ExportHandler(pgSolutionStorageCtor, metadataStorageCtor, solutionClient, config.ExportPath))
	registerRoute(mux, "/distil/ingest/:index/:dataset", routes.IngestHandler(metadataStorageCtor, ingestConfig))
	registerRoute(mux, "/distil/version", routes.VersionHandler(version, timestamp))
	registerRoute(mux, "/ws", ws.SolutionHandler(solutionClient, metadataStorageCtor, pgDataStorageCtor, pgSolutionStorageCtor))

	// POST
	registerRoutePost(mux, "/distil/variables/:index/:dataset", routes.VariableTypeHandler(pgDataStorageCtor, metadataStorageCtor))
	registerRoutePost(mux, "/distil/discovery/:index/:dataset/:target", routes.ProblemDiscoveryHandler(pgDataStorageCtor, metadataStorageCtor, config.UserProblemPath))
	registerRoutePost(mux, "/distil/data/:esIndex/:dataset/:invert", routes.DataHandler(pgDataStorageCtor, metadataStorageCtor))
	registerRoutePost(mux, "/distil/results/:index/:dataset/:solution-id", routes.ResultsHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/variable-summary/:index/:dataset/:variable", routes.VariableSummaryHandler(pgDataStorageCtor))
	registerRoutePost(mux, "/distil/results-variable-summary/:index/:dataset/:variable/:min/:max/:results-uuid", routes.ResultVariableSummaryHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/residuals-summary/:index/:dataset/:min/:max/:results-uuid", routes.ResidualsSummaryHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/results-summary/:index/:dataset/:min/:max/:results-uuid", routes.ResultsSummaryHandler(pgSolutionStorageCtor, pgDataStorageCtor))

	// static
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

func waitForPostEndpoint(endpoint string) bool {
	up := false
	resp, err := http.Post(endpoint, "application/json", strings.NewReader("test"))
	log.Infof("Sent request to %s", endpoint)
	log.Infof("response error: %v", err)
	if err != nil {
		// If the error indicates the service is up, then stop waiting.
		if !strings.Contains(err.Error(), "connection refused") {
			up = true
		}
	} else {
		up = true
	}
	if resp != nil {
		resp.Body.Close()
	}

	return up
}
