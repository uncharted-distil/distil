package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/unchartedsoftware/distil-ingest/metadata"
	log "github.com/unchartedsoftware/plog"
	"github.com/zenazn/goji/graceful"
	goji "goji.io"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil-compute/primitive/compute"
	api "github.com/unchartedsoftware/distil/api/compute"
	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/env"
	"github.com/unchartedsoftware/distil/api/middleware"
	"github.com/unchartedsoftware/distil/api/model"
	dm "github.com/unchartedsoftware/distil/api/model/storage/datamart"
	es "github.com/unchartedsoftware/distil/api/model/storage/elastic"
	pg "github.com/unchartedsoftware/distil/api/model/storage/postgres"
	"github.com/unchartedsoftware/distil/api/postgres"
	"github.com/unchartedsoftware/distil/api/rest"
	"github.com/unchartedsoftware/distil/api/routes"
	"github.com/unchartedsoftware/distil/api/service"
	"github.com/unchartedsoftware/distil/api/task"
	"github.com/unchartedsoftware/distil/api/util"
	"github.com/unchartedsoftware/distil/api/ws"
)

var (
	version        = "unset"
	timestamp      = "unset"
	problemPath    = ""
	datasetDocPath = ""
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
	api.SetDatasetDir(config.TmpDataPath)
	api.SetInputDir(config.D3MInputDirRoot)

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

	// make sure a connection can be made to postgres - doesn't appear to be thread safe and
	// causes panic if deferred, so we'll do it an a retry loop here.  We need to provide
	// flexibility on startup because we can't guarantee the DB will be up before the server.
	for name, test := range servicesToWait {
		log.Infof("Waiting for service '%s'", name)
		err = service.WaitForService(name, &config, test)
		if err == nil {
			log.Infof("Service '%s' is up", name)
		} else {
			log.Errorf("%+v", err)
			os.Exit(1)
		}
	}

	// instantiate the metadata storage (using ES).
	metadataStorageCtor := es.NewMetadataStorage(config.ESDatasetsIndex, esClientCtor)

	// instantiate the postgres data storage constructor.
	pgDataStorageCtor := pg.NewDataStorage(postgresClientCtor, metadataStorageCtor)

	// instantiate the postgres solution storage constructor.
	pgSolutionStorageCtor := pg.NewSolutionStorage(postgresClientCtor, metadataStorageCtor)

	var solutionClient *compute.Client
	if config.UseTA2Runner {
		// Instantiate the solution compute client mock
		solutionClient, err = compute.NewClientWithRunner(
			config.SolutionComputeEndpoint,
			config.SolutionComputeMockEndpoint,
			config.SolutionComputeTrace,
			userAgent,
			time.Duration(config.SolutionComputePullTimeout)*time.Second,
			config.SolutionComputePullMax,
			config.SkipPreprocessing)
		if err != nil {
			log.Errorf("%+v", err)
			os.Exit(1)
		}
	} else {
		// Instantiate the solution compute client
		solutionClient, err = compute.NewClient(
			config.SolutionComputeEndpoint,
			config.SolutionComputeTrace,
			userAgent,
			time.Duration(config.SolutionComputePullTimeout)*time.Second,
			config.SolutionComputePullMax,
			config.SkipPreprocessing)
		if err != nil {
			log.Errorf("%+v", err)
			os.Exit(1)
		}
	}
	defer solutionClient.Close()

	// reset the exported problem list
	if config.IsTask1 {
		problemListingFile := path.Join(config.UserProblemPath, routes.ProblemLabelFile)
		err = os.MkdirAll(config.UserProblemPath, 0755)
		if err != nil {
			log.Errorf("%+v", err)
			os.Exit(1)
		}

		err = util.WriteFileWithDirs(problemListingFile, []byte("problem_id,system,meaningful\n"), 0777)
		if err != nil {
			log.Errorf("%+v", err)
			os.Exit(1)
		}
		datasetDocPath = path.Join(config.D3MInputDir, "TRAIN", "dataset_TRAIN", compute.D3MDataSchema)
	} else {
		// NOTE: EVAL ONLY OVERRIDE SETUP FOR METRICS!
		problemPath = path.Join(config.D3MInputDir, "TRAIN", "problem_TRAIN", api.D3MProblem)
		ws.SetProblemFile(problemPath)
	}

	// set the ingest client to use
	task.SetClient(solutionClient)

	// build the ingest configuration.
	resolver := util.NewPathResolver(&util.PathConfig{
		InputFolder:     config.DataFolderPath,
		InputSubFolders: path.Join("TRAIN", "dataset_TRAIN"),
		OutputFolder:    config.TmpDataPath,
	})
	ingestConfig := &task.IngestTaskConfig{
		Resolver:                           resolver,
		HasHeader:                          true,
		ClusteringOutputDataRelative:       config.ClusteringOutputDataRelative,
		ClusteringOutputSchemaRelative:     config.ClusteringOutputSchemaRelative,
		ClusteringEnabled:                  config.ClusteringEnabled,
		FeaturizationOutputDataRelative:    config.FeaturizationOutputDataRelative,
		FeaturizationOutputSchemaRelative:  config.FeaturizationOutputSchemaRelative,
		FormatOutputDataRelative:           config.FormatOutputDataRelative,
		FormatOutputSchemaRelative:         config.FormatOutputSchemaRelative,
		GeocodingOutputDataRelative:        config.GeocodingOutputDataRelative,
		GeocodingOutputSchemaRelative:      config.GeocodingOutputSchemaRelative,
		MergedOutputPathRelative:           config.MergedOutputDataPath,
		MergedOutputSchemaPathRelative:     config.MergedOutputSchemaPath,
		SchemaPathRelative:                 config.SchemaPath,
		ClassificationOutputPathRelative:   config.ClassificationOutputPath,
		ClassificationProbabilityThreshold: config.ClassificationProbabilityThreshold,
		ClassificationEnabled:              config.ClassificationEnabled,
		RankingOutputPathRelative:          config.RankingOutputPath,
		RankingRowLimit:                    config.RankingRowLimit,
		DatabasePassword:                   config.PostgresPassword,
		DatabaseUser:                       config.PostgresUser,
		Database:                           config.PostgresDatabase,
		DatabaseHost:                       config.PostgresHost,
		DatabasePort:                       config.PostgresPort,
		SummaryOutputPathRelative:          config.SummaryPath,
		SummaryMachineOutputPathRelative:   config.SummaryMachinePath,
		ESEndpoint:                         config.ElasticEndpoint,
		ESTimeout:                          config.ElasticTimeout,
		ESDatasetPrefix:                    config.ElasticDatasetPrefix,
		HardFail:                           config.IngestHardFail,
	}
	sourceFolder := config.DataFolderPath

	// instantiate the metadata storage (using datamart).
	datamartClientCtor := rest.NewClient(config.DatamartURI)
	datamartMetadataStorageCtor := dm.NewMetadataStorage(config.DatamartImportFolder, ingestConfig, datamartClientCtor)

	// Ingest the data specified by the environment
	if config.InitialDataset != "" && !config.SkipIngest {
		log.Infof("Loading initial dataset '%s'", config.InitialDataset)
		err = task.IngestDataset(metadataStorageCtor, config.ESDatasetsIndex, config.InitialDataset, metadata.Seed, ingestConfig)
		if err != nil {
			log.Errorf("%+v", err)
			os.Exit(1)
		}
		sourceFolder = path.Dir(ingestConfig.GetTmpAbsolutePath(ingestConfig.GeocodingOutputSchemaRelative))
	}
	datasetsToProxy := parseResourceProxy(config.ResourceProxy)

	// register routes
	mux := goji.NewMux()
	mux.Use(middleware.Log)
	mux.Use(middleware.Gzip)

	routes.SetVerboseError(config.VerboseError)

	// GET
	// ** TEMPORARILY COMMENTED OUT DATAMART STORAGE DUE TO BREAKING API CHANGE.
	registerRoute(mux, "/distil/datasets", routes.DatasetsHandler([]model.MetadataStorageCtor{metadataStorageCtor, datamartMetadataStorageCtor}))
	registerRoute(mux, "/distil/datasets/:dataset", routes.DatasetHandler(metadataStorageCtor))
	registerRoute(mux, "/distil/solutions/:dataset/:target/:solution-id", routes.SolutionHandler(pgSolutionStorageCtor))
	registerRoute(mux, "/distil/variables/:dataset", routes.VariablesHandler(metadataStorageCtor))
	registerRoute(mux, "/distil/variable-rankings/:dataset/:target", routes.VariableRankingHandler(metadataStorageCtor))
	registerRoute(mux, "/distil/residuals-extrema/:dataset/:target", routes.ResidualsExtremaHandler(metadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/abort", routes.AbortHandler())
	registerRoute(mux, "/distil/export/:solution-id", routes.ExportHandler(pgSolutionStorageCtor, metadataStorageCtor, solutionClient, config.D3MOutputDir))
	registerRoute(mux, "/distil/config", routes.ConfigHandler(config, version, timestamp, problemPath, datasetDocPath))
	registerRoute(mux, "/ws", ws.SolutionHandler(solutionClient, metadataStorageCtor, pgDataStorageCtor, pgSolutionStorageCtor))

	// POST
	registerRoutePost(mux, "/distil/variables/:dataset", routes.VariableTypeHandler(pgDataStorageCtor, metadataStorageCtor))
	registerRoutePost(mux, "/distil/discovery/:dataset/:target", routes.ProblemDiscoveryHandler(pgDataStorageCtor, metadataStorageCtor, config.UserProblemPath, userAgent, config.SkipPreprocessing))
	registerRoutePost(mux, "/distil/data/:dataset/:invert", routes.DataHandler(pgDataStorageCtor, metadataStorageCtor))
	registerRoutePost(mux, "/distil/import/:dataset/:source/:index", routes.ImportHandler(datamartMetadataStorageCtor, metadataStorageCtor, ingestConfig))
	registerRoutePost(mux, "/distil/results/:dataset/:solution-id", routes.ResultsHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/variable-summary/:dataset/:variable", routes.VariableSummaryHandler(pgDataStorageCtor))
	registerRoutePost(mux, "/distil/training-summary/:dataset/:variable/:results-uuid", routes.TrainingSummaryHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/target-summary/:dataset/:target/:results-uuid", routes.TargetSummaryHandler(metadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/residuals-summary/:dataset/:target/:results-uuid", routes.ResidualsSummaryHandler(metadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/correctness-summary/:dataset/:results-uuid", routes.CorrectnessSummaryHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/predicted-summary/:dataset/:target/:results-uuid", routes.PredictedSummaryHandler(metadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/geocode/:dataset/:variable", routes.GeocodingHandler(metadataStorageCtor, pgDataStorageCtor, sourceFolder))
	registerRoutePost(mux, "/distil/upload/:dataset", routes.UploadHandler(config.DatamartImportFolder))
	registerRoutePost(mux, "/distil/join/:dataset-left/:column-left/:source-left/:dataset-right/:column-right/:source-right", routes.JoinHandler(metadataStorageCtor))

	// static
	registerRoute(mux, "/distil/image/:dataset/:file", routes.ImageHandler(config.DataFolderPath, config.RootResourceDirectory, datasetsToProxy))
	registerRoute(mux, "/distil/timeseries/:dataset/:file", routes.TimeseriesHandler(config.DataFolderPath, config.RootResourceDirectory, datasetsToProxy))
	registerRoute(mux, "/distil/graphs/:dataset/:file", routes.GraphsHandler(config.DataFolderPath, config.RootResourceDirectory, datasetsToProxy))
	registerRoute(mux, "/*", routes.FileHandler("./dist"))

	// catch kill signals for graceful shutdown
	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)

	// kick off the server listen loop
	log.Infof("Listening on port %s", config.AppPort)
	err = graceful.ListenAndServe(":"+config.AppPort, mux)
	if err != nil {
		log.Errorf("%+v", err)
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

func parseResourceProxy(datasets string) map[string]bool {
	toProxy := make(map[string]bool)
	datasetIds := strings.Split(datasets, ",")
	for _, d := range datasetIds {
		toProxy[d] = true
	}

	return toProxy
}
