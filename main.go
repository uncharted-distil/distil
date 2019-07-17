//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

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
	"github.com/uncharted-distil/distil-ingest/metadata"
	log "github.com/unchartedsoftware/plog"
	"github.com/zenazn/goji/graceful"
	goji "goji.io"
	"goji.io/pat"

	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/compute"
	"github.com/uncharted-distil/distil/api/elastic"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/middleware"
	"github.com/uncharted-distil/distil/api/model"
	dm "github.com/uncharted-distil/distil/api/model/storage/datamart"
	es "github.com/uncharted-distil/distil/api/model/storage/elastic"
	"github.com/uncharted-distil/distil/api/model/storage/file"
	pg "github.com/uncharted-distil/distil/api/model/storage/postgres"
	"github.com/uncharted-distil/distil/api/postgres"
	"github.com/uncharted-distil/distil/api/rest"
	"github.com/uncharted-distil/distil/api/routes"
	"github.com/uncharted-distil/distil/api/service"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util"
	"github.com/uncharted-distil/distil/api/ws"
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

	err = env.Initialize(&config)
	if err != nil {
		log.Errorf("%+v", err)
		os.Exit(1)
	}

	// set dataset directory
	api.SetDatasetDir(config.TmpDataPath)

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
	esMetadataStorageCtor := es.NewMetadataStorage(config.ESDatasetsIndex, esClientCtor)

	// instantiate the metadata storage (using filesystem).
	fileMetadataStorageCtor := file.NewMetadataStorage(config.TmpDataPath)

	// instantiate the postgres data storage constructor.
	pgDataStorageCtor := pg.NewDataStorage(postgresClientCtor, esMetadataStorageCtor)

	// instantiate the postgres solution storage constructor.
	pgSolutionStorageCtor := pg.NewSolutionStorage(postgresClientCtor, esMetadataStorageCtor)

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
		err = os.MkdirAll(config.UserProblemPath, os.ModePerm)
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
	ingestConfig := &task.IngestTaskConfig{
		HasHeader:                          true,
		ClusteringOutputDataRelative:       config.ClusteringOutputDataRelative,
		ClusteringOutputSchemaRelative:     config.ClusteringOutputSchemaRelative,
		ClusteringEnabled:                  config.ClusteringEnabled,
		FeaturizationOutputDataRelative:    config.FeaturizationOutputDataRelative,
		FeaturizationOutputSchemaRelative:  config.FeaturizationOutputSchemaRelative,
		FormatOutputDataRelative:           config.FormatOutputDataRelative,
		FormatOutputSchemaRelative:         config.FormatOutputSchemaRelative,
		CleanOutputDataRelative:            config.CleanOutputDataRelative,
		CleanOutputSchemaRelative:          config.CleanOutputSchemaRelative,
		GeocodingOutputDataRelative:        config.GeocodingOutputDataRelative,
		GeocodingOutputSchemaRelative:      config.GeocodingOutputSchemaRelative,
		GeocodingEnabled:                   config.GeocodingEnabled,
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
		SummaryEnabled:                     config.SummaryEnabled,
		ESEndpoint:                         config.ElasticEndpoint,
		ESTimeout:                          config.ElasticTimeout,
		ESDatasetPrefix:                    config.ElasticDatasetPrefix,
		HardFail:                           config.IngestHardFail,
	}
	sourceFolder := config.DataFolderPath

	// instantiate the metadata storage (using datamart).
	datamartCtors := make(map[string]model.MetadataStorageCtor)
	if config.DatamartNYUEnabled {
		nyuDatamartClientCtor := rest.NewClient(config.DatamartURINYU)
		datamartCtors[dm.ProvenanceNYU] = dm.NewNYUMetadataStorage(config.DatamartImportFolder, ingestConfig, nyuDatamartClientCtor)
	}
	if config.DatamartISIEnabled {
		isiDatamartClientCtor := rest.NewClient(config.DatamartURIISI)
		datamartCtors[dm.ProvenanceISI] = dm.NewISIMetadataStorage(config.DatamartImportFolder, ingestConfig, isiDatamartClientCtor)
	}
	datamartCtors[es.Provenance] = esMetadataStorageCtor

	// set extremas
	//esStorage, err := esMetadataStorageCtor()
	//if err != nil {
	//	log.Errorf("%+v", err)
	//	os.Exit(1)
	//}

	//pgStorage, err := pgDataStorageCtor()
	//if err != nil {
	//	log.Errorf("%+v", err)
	//	os.Exit(1)
	//}

	//err = updateExtremas(esStorage, pgStorage)
	//if err != nil {
	//	log.Errorf("%+v", err)
	//	os.Exit(1)
	//}

	// Ingest the data specified by the environment
	if config.InitialDataset != "" && !config.SkipIngest {
		log.Infof("Loading initial dataset '%s'", config.InitialDataset)
		err = util.Copy(path.Join(config.InitialDataset, "TRAIN", "dataset_TRAIN"), path.Join(config.DatamartImportFolder, "initial"))
		if err != nil {
			log.Errorf("%+v", err)
			os.Exit(1)
		}
		err = task.IngestDataset(metadata.Contrib, pgDataStorageCtor, esMetadataStorageCtor, config.ESDatasetsIndex, "initial", nil, ingestConfig)
		if err != nil {
			log.Errorf("%+v", err)
			os.Exit(1)
		}

		sourceFolder = env.ResolvePath(metadata.Contrib, ingestConfig.GeocodingOutputSchemaRelative)
		sourceFolder = path.Dir(sourceFolder)
	}

	// register routes
	mux := goji.NewMux()
	mux.Use(middleware.Log)
	mux.Use(middleware.Gzip)

	routes.SetVerboseError(config.VerboseError)

	// GET
	registerRoute(mux, "/distil/datasets", routes.DatasetsHandler(datamartCtors))
	registerRoute(mux, "/distil/datasets/:dataset", routes.DatasetHandler(esMetadataStorageCtor))
	registerRoute(mux, "/distil/join-suggestions/:dataset", routes.JoinSuggestionHandler(esMetadataStorageCtor, datamartCtors))
	registerRoute(mux, "/distil/solutions/:dataset/:target/:solution-id", routes.SolutionHandler(pgSolutionStorageCtor))
	registerRoute(mux, "/distil/variables/:dataset", routes.VariablesHandler(esMetadataStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/variable-rankings/:dataset/:target", routes.VariableRankingHandler(esMetadataStorageCtor))
	registerRoute(mux, "/distil/residuals-extrema/:dataset/:target", routes.ResidualsExtremaHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/abort", routes.AbortHandler())
	registerRoute(mux, "/distil/export/:solution-id", routes.ExportHandler(solutionClient, config.D3MOutputDir))
	registerRoute(mux, "/distil/config", routes.ConfigHandler(config, version, timestamp, problemPath, datasetDocPath))
	registerRoute(mux, "/ws", ws.SolutionHandler(solutionClient, esMetadataStorageCtor, pgDataStorageCtor, pgSolutionStorageCtor))

	// POST
	registerRoutePost(mux, "/distil/grouping/:dataset", routes.GroupingHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/remove-grouping/:dataset", routes.RemoveGroupingHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/variables/:dataset", routes.VariableTypeHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/discovery/:dataset/:target", routes.ProblemDiscoveryHandler(pgDataStorageCtor, esMetadataStorageCtor, config.UserProblemPath, userAgent, config.SkipPreprocessing))
	registerRoutePost(mux, "/distil/data/:dataset/:invert", routes.DataHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/import/:datasetID/:source/:provenance", routes.ImportHandler(pgDataStorageCtor, datamartCtors, fileMetadataStorageCtor, esMetadataStorageCtor, ingestConfig))
	registerRoutePost(mux, "/distil/compose/:dataset", routes.ComposeHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/delete/:dataset/:variable", routes.DeleteHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/results/:dataset/:solution-id", routes.ResultsHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/variable-summary/:dataset/:variable/:invert", routes.VariableSummaryHandler(pgDataStorageCtor))
	registerRoutePost(mux, "/distil/timeseries-summary/:dataset/:xColName/:yColName/:binningInterval/:invert", routes.TimeseriesSummaryHandler(pgDataStorageCtor))
	registerRoutePost(mux, "/distil/training-summary/:dataset/:variable/:results-uuid", routes.TrainingSummaryHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/training-timeseries-summary/:dataset/:xColName/:yColName/:binningInterval/:results-uuid", routes.TrainingTimeseriesSummaryHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/target-timeseries-summary/:dataset/:xColName/:yColName/:binningInterval/:results-uuid", routes.TargetTimeseriesSummaryHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/forecasting-summary/:dataset/:xColName/:yColName/:binningInterval/:results-uuid", routes.ForecastingSummaryHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/target-summary/:dataset/:target/:results-uuid", routes.TargetSummaryHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/residuals-summary/:dataset/:target/:results-uuid", routes.ResidualsSummaryHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/correctness-summary/:dataset/:results-uuid", routes.CorrectnessSummaryHandler(pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/predicted-summary/:dataset/:target/:results-uuid", routes.PredictedSummaryHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/geocode/:dataset/:variable", routes.GeocodingHandler(esMetadataStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/cluster/:dataset/:variable", routes.ClusteringHandler(esMetadataStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/upload/:dataset", routes.UploadHandler(path.Join(config.TmpDataPath, config.AugmentedSubFolder), ingestConfig))
	registerRoutePost(mux, "/distil/join/:dataset-left/:column-left/:source-left/:dataset-right/:column-right/:source-right", routes.JoinHandler(esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/timeseries/:dataset/:timeseriesColName/:xColName/:yColName/:timeseriesURI/:invert", routes.TimeseriesHandler(pgDataStorageCtor))
	registerRoutePost(mux, "/distil/timeseries-forecast/:dataset/:timeseriesColName/:xColName/:yColName/:timeseriesURI/:result-uuid", routes.TimeseriesForecastHandler(pgDataStorageCtor, pgSolutionStorageCtor))

	// static
	registerRoute(mux, "/distil/image/:dataset/:source/:file", routes.ImageHandler(esMetadataStorageCtor, &config))
	registerRoute(mux, "/distil/graphs/:dataset/:file", routes.GraphsHandler(config.DataFolderPath))
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

func updateExtremas(metaStorage model.MetadataStorage, dataStorage model.DataStorage) error {
	log.Infof("updating extremas on startup")
	datasets, err := metaStorage.FetchDatasets(false, false)
	if err != nil {
		return err
	}

	for _, d := range datasets {
		log.Infof("updating extremas for dataset %s", d.Name)
		err = task.UpdateExtremas(d.ID, metaStorage, dataStorage)
		if err != nil {
			return err
		}
		log.Infof("done updating extremas for %s", d.Name)
	}

	log.Infof("done updating all extremas")

	return nil
}
