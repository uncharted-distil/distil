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
	"syscall"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"github.com/zenazn/goji/graceful"
	goji "goji.io/v3"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/metadata"
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
	ta2Version     = ""
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

	createOutputFolders(&config)

	// initialize the pipeline cache and
	pipelineCacheFilename := path.Join(env.GetTmpPath(), config.PipelineCacheFilename)
	err = api.InitializeCache(pipelineCacheFilename)
	if err != nil {
		log.Errorf("%+v", err)
		os.Exit(1)
	}
	api.InitializeQueue(&config)

	// initialize the user event logger - records user interactions with the system in a CSV file for post-run
	// analysis
	discoveryLogger, err := env.NewDiscoveryLogger("event-"+util.GenerateTimeFileNameStr()+".csv", &config)
	if err != nil {
		log.Errorf("%+v", err)
		os.Exit(1)
	}

	// set dataset directory
	api.SetDatasetDir(config.D3MOutputDir)

	// instantiate elastic client constructor.
	esClientCtor := elastic.NewClient(config.ElasticEndpoint, false)

	// instantiate the postgres client constructor.
	postgresClientCtor := postgres.NewClient(config.PostgresHost, config.PostgresPort, config.PostgresUser, config.PostgresPassword,
		config.PostgresDatabase, config.PostgresLogLevel, false)
	postgresBatchClientCtor := postgres.NewClient(config.PostgresHost, config.PostgresPort, config.PostgresUser, config.PostgresPassword,
		config.PostgresDatabase, "error", true)

	// instantiate the metadata storage (using ES).
	esMetadataStorageCtor := es.NewMetadataStorage(config.ESDatasetsIndex, false, esClientCtor)

	// instantiate the exported model storage (using ES).
	esExportedModelStorageCtor := es.NewExportedModelStorage(config.ESModelsIndex, false, esClientCtor)

	// instantiate the metadata storage (using filesystem).
	fileMetadataStorageCtor := file.NewMetadataStorage(config.D3MOutputDir)

	// instantiate the postgres data storage constructor.
	pgDataStorageCtor := pg.NewDataStorage(postgresClientCtor, postgresBatchClientCtor, esMetadataStorageCtor)

	// instantiate the postgres solution storage constructor.
	pgSolutionStorageCtor := pg.NewSolutionStorage(postgresClientCtor, esMetadataStorageCtor)

	// Instantiate the solution compute client
	solutionClient, err := task.NewDefaultClient(config, userAgent, discoveryLogger)
	if err != nil {
		log.Errorf("%+v", err)
		os.Exit(1)
	}
	defer solutionClient.Close()

	// wait for required services.
	servicesToWait["postgres"] = func() bool {
		_, err := postgresClientCtor()
		return err == nil
	}
	servicesToWait["elastic"] = func() bool {
		_, err := esClientCtor()
		return err == nil
	}
	servicesToWait["ta2"] = func() bool {
		versionNumber, err := solutionClient.Hello()
		ta2Version = versionNumber
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

	// NOTE: EVAL ONLY OVERRIDE SETUP FOR METRICS!
	problemPath = path.Join(config.D3MInputDir, "TRAIN", "problem_TRAIN", api.D3MProblem)

	// set the postgres random seed for data table reading
	pg.SetRandomSeed(config.PostgresRandomSeed)

	// set the ingest client to use
	task.SetClient(solutionClient)

	// build the ingest configuration.
	ingestConfig := task.NewConfig(config)

	// instantiate the metadata storage (using datamart).
	datamartCtors := make(map[string]model.MetadataStorageCtor)
	if config.DatamartNYUEnabled {
		nyuDatamartClientCtor := rest.NewClient(config.DatamartURINYU)
		datamartCtors[dm.ProvenanceNYU] = dm.NewNYUMetadataStorage(config.DatamartImportFolder, &config, ingestConfig, nyuDatamartClientCtor)
	}
	if config.DatamartISIEnabled {
		isiDatamartClientCtor := rest.NewClient(config.DatamartURIISI)
		datamartCtors[dm.ProvenanceISI] = dm.NewISIMetadataStorage(config.DatamartImportFolder, &config, ingestConfig, isiDatamartClientCtor)
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
		_, err = task.IngestDataset(metadata.Contrib, pgDataStorageCtor, esMetadataStorageCtor,
			"initial", nil, model.DatasetTypeModelling, ingestConfig, &task.IngestSteps{ClassificationOverwrite: true})
		if err != nil {
			log.Errorf("%+v", err)
			os.Exit(1)
		}
	}

	// register routes
	mux := goji.NewMux()
	mux.Use(middleware.Log)
	mux.Use(middleware.Gzip)

	routes.SetVerboseError(config.VerboseError)
	// GET
	registerRoute(mux, "/distil/datasets", routes.DatasetsHandler(datamartCtors))
	registerRoute(mux, "/distil/available", routes.AvailableDatasetsHandler(esMetadataStorageCtor))
	registerRoute(mux, "/distil/datasets/:dataset", routes.DatasetHandler(esMetadataStorageCtor))
	registerRoute(mux, "/distil/models", routes.ModelsHandler(esExportedModelStorageCtor))
	registerRoute(mux, "/distil/models/:model", routes.ModelHandler(esExportedModelStorageCtor))
	registerRoute(mux, "/distil/join-suggestions/:dataset", routes.JoinSuggestionHandler(esMetadataStorageCtor, datamartCtors))
	registerRoute(mux, "/distil/solution/:solution-id", routes.SolutionHandler(pgSolutionStorageCtor))
	registerRoute(mux, "/distil/solutions/:dataset/:target", routes.SolutionsHandler(pgSolutionStorageCtor))
	registerRoute(mux, "/distil/solution-requests/:dataset/:target", routes.SolutionRequestsHandler(pgSolutionStorageCtor))
	registerRoute(mux, "/distil/solution-request/:request-id", routes.SolutionRequestHandler(pgSolutionStorageCtor))
	registerRoute(mux, "/distil/prediction/:request-id", routes.PredictionHandler(pgSolutionStorageCtor))
	registerRoute(mux, "/distil/predictions/:fitted-solution-id", routes.PredictionsHandler(pgSolutionStorageCtor))
	registerRoute(mux, "/distil/variables/:dataset", routes.VariablesHandler(esMetadataStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/variable-rankings/:dataset/:target", routes.VariableRankingHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/residuals-extrema/:dataset/:target", routes.ResidualsExtremaHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/export/:solution-id", routes.ExportHandler(solutionClient, config.D3MOutputDir, discoveryLogger))
	registerRoute(mux, "/distil/config", routes.ConfigHandler(config, version, timestamp, problemPath, datasetDocPath, ta2Version))
	registerRoute(mux, "/distil/task/:dataset/:target/:variables", routes.TaskHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoute(mux, "/distil/multiband-image/:dataset/:image-id/:band-combination/:is-thumbnail/*", routes.MultiBandImageHandler(esMetadataStorageCtor))
	registerRoute(mux, "/distil/multiband-combinations/:dataset", routes.MultiBandCombinationsHandler(esMetadataStorageCtor))
	registerRoute(mux, "/distil/load/:solution-id/:fitted", routes.LoadHandler(esExportedModelStorageCtor, pgSolutionStorageCtor, esMetadataStorageCtor))
	registerRoute(mux, "/distil/solution-variable-rankings/:solution-id", routes.SolutionVariableRankingHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoute(mux, "/distil/export-results/:produce-request-id", routes.ExportResultHandler(pgSolutionStorageCtor, pgDataStorageCtor, esMetadataStorageCtor))
	registerRoute(mux, "/distil/model-metrics/:task", routes.ModelMetricsHandler(esMetadataStorageCtor))
	registerRoute(mux, "/ws", ws.SolutionHandler(solutionClient, esMetadataStorageCtor, pgDataStorageCtor, pgSolutionStorageCtor, esExportedModelStorageCtor))
	registerRoute(mux, "/distil/image-attention/:dataset/:result-id/:index/:opacity/:color-scale", routes.ImageAttentionHandler(pgSolutionStorageCtor, esMetadataStorageCtor))

	// POST
	registerRoutePost(mux, "/distil/grouping/:dataset", routes.GroupingHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/remove-grouping/:dataset/:variable", routes.RemoveGroupingHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/variables/:dataset", routes.VariableTypeHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/data/:dataset/:invert", routes.DataHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/import/:datasetID/:source/:provenance", routes.ImportHandler(pgDataStorageCtor, datamartCtors, fileMetadataStorageCtor, esMetadataStorageCtor, &config))
	registerRoutePost(mux, "/distil/delete/:dataset/:variable", routes.DeleteHandler(pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/results/:dataset/:solution-id", routes.ResultsHandler(pgSolutionStorageCtor, pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/prediction-results/:produce-request-id", routes.PredictionResultsHandler(pgSolutionStorageCtor, pgDataStorageCtor, esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/variable-summary/:dataset/:variable/:invert/:mode", routes.VariableSummaryHandler(esMetadataStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/training-summary/:dataset/:variable/:results-uuid/:mode", routes.TrainingSummaryHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/target-summary/:dataset/:target/:results-uuid/:mode", routes.TargetSummaryHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/residuals-summary/:dataset/:target/:results-uuid/:mode", routes.ResidualsSummaryHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/correctness-summary/:dataset/:results-uuid/:mode", routes.CorrectnessSummaryHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/confidence-summary/:dataset/:results-uuid/:mode", routes.ConfidenceSummaryHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/prediction-result-summary/:results-uuid/:mode", routes.PredictionResultSummaryHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/solution-result-summary/:results-uuid/:mode", routes.SolutionResultSummaryHandler(esMetadataStorageCtor, pgSolutionStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/geocode/:dataset/:variable", routes.GeocodingHandler(esMetadataStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/cluster/:dataset/:variable", routes.ClusteringHandler(esMetadataStorageCtor, pgDataStorageCtor, config))
	registerRoutePost(mux, "/distil/cluster/:result-id", routes.ClusteringExplainHandler(pgSolutionStorageCtor, esMetadataStorageCtor, pgDataStorageCtor, config))
	registerRoutePost(mux, "/distil/upload/:dataset", routes.UploadHandler(&config))
	registerRoutePost(mux, "/distil/update/:dataset", routes.UpdateHandler(esMetadataStorageCtor, pgDataStorageCtor, config))
	registerRoutePost(mux, "/distil/clone/:dataset", routes.CloningHandler(esMetadataStorageCtor, pgDataStorageCtor, config))
	registerRoutePost(mux, "/distil/add-field/:dataset", routes.AddFieldHandler(esMetadataStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/extract/:dataset/:invert", routes.ExtractHandler(esMetadataStorageCtor, pgDataStorageCtor, config))
	registerRoutePost(mux, "/distil/join", routes.JoinHandler(esMetadataStorageCtor))
	registerRoutePost(mux, "/distil/timeseries/:dataset/:timeseriesColName/:xColName/:yColName/:invert", routes.TimeseriesHandler(esMetadataStorageCtor, pgDataStorageCtor))
	registerRoutePost(mux, "/distil/timeseries-forecast/:truthDataset/:forecastDataset/:timeseriesColName/:xColName/:yColName/:result-uuid", routes.TimeseriesForecastHandler(esMetadataStorageCtor, pgDataStorageCtor, pgSolutionStorageCtor, config.TrainTestSplitTimeSeries))
	registerRoutePost(mux, "/distil/event", routes.UserEventHandler(discoveryLogger))
	registerRoutePost(mux, "/distil/save/:solution-id/:fitted", routes.SaveHandler(esExportedModelStorageCtor, pgSolutionStorageCtor, esMetadataStorageCtor))

	// static
	registerRoute(mux, "/distil/image/:dataset/:file", routes.ImageHandler(esMetadataStorageCtor, &config))
	registerRoute(mux, "/distil/graphs/:dataset/:file", routes.GraphsHandler(config.D3MInputDir))
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

//nolint
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

func createOutputFolders(config *env.Config) {
	// create the augmented data folder
	augmentPath := env.GetAugmentedPath()
	if err := os.MkdirAll(augmentPath, os.ModePerm); err != nil {
		log.Error(errors.Wrap(err, "failed to created output folder"))
	}

	// create the public data folder
	publicPath := env.GetPublicPath()
	if err := os.MkdirAll(publicPath, os.ModePerm); err != nil {
		log.Error(errors.Wrap(err, "failed to created public folder"))
	}
}
