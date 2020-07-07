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

package task

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/middleware"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/postgres"
)

const (
	baseTableSuffix    = "_base"
	explainTableSuffix = "_explain"
)

// IngestTaskConfig captures the necessary configuration for an data ingest.
type IngestTaskConfig struct {
	HasHeader                          bool
	ClusteringOutputDataRelative       string
	ClusteringOutputSchemaRelative     string
	ClusteringEnabled                  bool
	FeaturizationOutputDataRelative    string
	FeaturizationOutputSchemaRelative  string
	FormatOutputDataRelative           string
	FormatOutputSchemaRelative         string
	CleanOutputDataRelative            string
	CleanOutputSchemaRelative          string
	GeocodingOutputDataRelative        string
	GeocodingOutputSchemaRelative      string
	GeocodingEnabled                   bool
	MergedOutputPathRelative           string
	MergedOutputSchemaPathRelative     string
	SchemaPathRelative                 string
	ClassificationOutputPathRelative   string
	ClassificationProbabilityThreshold float64
	ClassificationEnabled              bool
	RankingOutputPathRelative          string
	RankingRowLimit                    int
	DatabasePassword                   string
	DatabaseUser                       string
	Database                           string
	DatabaseHost                       string
	DatabasePort                       int
	DatabaseBatchSize                  int
	DatabaseLogLevel                   string
	SummaryOutputPathRelative          string
	SummaryMachineOutputPathRelative   string
	SummaryEnabled                     bool
	ESEndpoint                         string
	ESTimeout                          int
	ESDatasetPrefix                    string
	HardFail                           bool
	IngestOverwrite                    bool
}

type IngestSteps struct {
	ClassificationOverwrite bool
}

// NewDefaultClient creates a new client to use when submitting pipelines.
func NewDefaultClient(config env.Config, userAgent string, discoveryLogger middleware.MethodLogger) (*compute.Client, error) {
	return compute.NewClient(
		config.SolutionComputeEndpoint,
		config.SolutionComputeTrace,
		userAgent,
		"TA2",
		time.Duration(config.SolutionComputePullTimeout)*time.Second,
		config.SolutionComputePullMax,
		config.SkipPreprocessing,
		discoveryLogger)
}

// NewConfig creates an ingest config based on a distil config.
func NewConfig(config env.Config) *IngestTaskConfig {
	return &IngestTaskConfig{
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
		DatabaseBatchSize:                  config.PostgresBatchSize,
		DatabaseLogLevel:                   config.PostgresLogLevel,
		SummaryOutputPathRelative:          config.SummaryPath,
		SummaryMachineOutputPathRelative:   config.SummaryMachinePath,
		SummaryEnabled:                     config.SummaryEnabled,
		ESEndpoint:                         config.ElasticEndpoint,
		ESTimeout:                          config.ElasticTimeout,
		ESDatasetPrefix:                    config.ElasticDatasetPrefix,
		HardFail:                           config.IngestHardFail,
		IngestOverwrite:                    config.IngestOverwrite,
	}
}

// IngestDataset executes the complete ingest process for the specified dataset.
func IngestDataset(datasetSource metadata.DatasetSource, dataCtor api.DataStorageCtor, metaCtor api.MetadataStorageCtor,
	dataset string, origins []*model.DatasetOrigin, datasetType api.DatasetType, config *IngestTaskConfig, steps *IngestSteps) (string, error) {
	// Set the probability threshold
	metadata.SetTypeProbabilityThreshold(config.ClassificationProbabilityThreshold)

	metaStorage, err := metaCtor()
	if err != nil {
		return "", errors.Wrap(err, "unable to initialize metadata storage")
	}

	dataStorage, err := dataCtor()
	if err != nil {
		return "", errors.Wrap(err, "unable to initialize data storage")
	}

	sourceFolder := env.ResolvePath(datasetSource, dataset)

	originalSchemaFile := path.Join(sourceFolder, config.SchemaPathRelative)
	latestSchemaOutput := originalSchemaFile

	output := latestSchemaOutput
	if config.ClusteringEnabled {
		output, err = ClusterDataset(latestSchemaOutput, dataset, config)
		if err != nil {
			if config.HardFail {
				return "", errors.Wrap(err, "unable to cluster all data")
			}
			log.Errorf("unable to cluster all data: %v", err)
		} else {
			latestSchemaOutput = output
		}
		log.Infof("finished clustering the dataset")
	}

	output, err = Merge(latestSchemaOutput, dataset, config)
	if err != nil {
		return "", errors.Wrap(err, "unable to merge all data into a single file")
	}
	latestSchemaOutput = output
	log.Infof("finished merging the dataset")

	output, err = Clean(latestSchemaOutput, dataset, config)
	if err != nil {
		return "", errors.Wrap(err, "unable to clean all data")
	}
	latestSchemaOutput = output
	log.Infof("finished cleaning the dataset")

	if steps.ClassificationOverwrite || !classificationExists(dataset, config) {
		_, err = Classify(latestSchemaOutput, dataset, config)
		if err != nil {
			return "", errors.Wrap(err, "unable to classify fields")
		}
		log.Infof("finished classifying the dataset")
	} else {
		log.Infof("skipping classification because it already exists")
	}

	_, err = Rank(latestSchemaOutput, dataset, config)
	if err != nil {
		log.Errorf("unable to rank field importance: %v", err)
	}
	log.Infof("finished ranking the dataset")

	if config.SummaryEnabled {
		_, err = Summarize(latestSchemaOutput, dataset, config)
		log.Infof("finished summarizing the dataset")
		if err != nil {
			if config.HardFail {
				return "", errors.Wrap(err, "unable to summarize the dataset")
			}
			log.Errorf("unable to summarize the dataset: %v", err)
		}
	} else {
		log.Infof("summarization disabled")
	}

	if config.GeocodingEnabled {
		output, err = GeocodeForwardDataset(latestSchemaOutput, dataset, config)
		if err != nil {
			return "", errors.Wrap(err, "unable to geocode all data")
		}
		latestSchemaOutput = output
		log.Infof("finished geocoding the dataset")
	}

	datasetID, err := Ingest(originalSchemaFile, latestSchemaOutput, metaStorage, dataset, datasetSource, origins, datasetType, config, true, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to ingest ranked data")
	}
	log.Infof("finished ingesting the dataset")

	// updating extremas is optional
	err = UpdateExtremas(datasetID, metaStorage, dataStorage)
	if err != nil {
		log.Errorf("unable to update extremas ranked data: %v", err)
	}
	log.Infof("finished updating extremas")

	return datasetID, nil
}

// Ingest the metadata to ES and the data to Postgres.
func Ingest(originalSchemaFile string, schemaFile string, storage api.MetadataStorage, dataset string, source metadata.DatasetSource,
	origins []*model.DatasetOrigin, datasetType api.DatasetType, config *IngestTaskConfig, checkMatch bool, fallbackMerged bool) (string, error) {
	_, meta, err := loadMetadataForIngest(originalSchemaFile, schemaFile, source, nil, config, true, fallbackMerged)
	if err != nil {
		return "", err
	}

	if !config.IngestOverwrite {
		// get the unique name, and if it is different then write out the updated metadata
		uniqueName, err := getUniqueDatasetName(meta, storage)
		if err != nil {
			return "", err
		}

		if uniqueName != meta.Name {
			extendedOutput := source == metadata.Augmented
			log.Infof("storing (extended: %v) metadata with new name to %s (new: '%s', old: '%s')", extendedOutput, originalSchemaFile, uniqueName, meta.Name)
			meta.Name = uniqueName
			meta.ID = model.NormalizeDatasetID(uniqueName)
			err = metadata.WriteSchema(meta, originalSchemaFile, extendedOutput)
			if err != nil {
				return "", errors.Wrap(err, "unable to store updated metadata")
			}
			log.Infof("updated metadata with new name written to %s", originalSchemaFile)
		}
	}

	// Connect to the database.
	postgresConfig := &postgres.Config{
		Password:         config.DatabasePassword,
		User:             config.DatabaseUser,
		Database:         config.Database,
		Host:             config.DatabaseHost,
		Port:             config.DatabasePort,
		BatchSize:        config.DatabaseBatchSize,
		PostgresLogLevel: "error",
	}
	pg, err := postgres.NewDatabase(postgresConfig, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to initialize a new database")
	}

	// Check for existing dataset
	if checkMatch && config.IngestOverwrite {
		match, err := matchDataset(storage, meta)
		// Ignore the error for now as if this fails we still want ingest to succeed.
		if err != nil {
			log.Error(err)
		}
		if match != "" {
			log.Infof("Matched %s to dataset %s", meta.Name, match)
			err = deleteDataset(match, pg, storage)
			if err != nil {
				log.Errorf("error deleting dataset: %v", err)
			}
			log.Infof("Deleted dataset %s", match)
		}
	}

	// ingest the metadata
	_, err = IngestMetadata(originalSchemaFile, schemaFile, storage, source, origins, datasetType, config, true, fallbackMerged)
	if err != nil {
		return "", err
	}

	// ingest the data
	err = IngestPostgres(originalSchemaFile, schemaFile, source, config, true, false, fallbackMerged)
	if err != nil {
		return "", err
	}

	return meta.ID, nil
}

// IngestMetadata ingests the data to ES.
func IngestMetadata(originalSchemaFile string, schemaFile string, storage api.MetadataStorage, source metadata.DatasetSource,
	origins []*model.DatasetOrigin, datasetType api.DatasetType, config *IngestTaskConfig, verifyMetadata bool, fallbackMerged bool) (string, error) {
	_, meta, err := loadMetadataForIngest(originalSchemaFile, schemaFile, source, origins, config, verifyMetadata, fallbackMerged)
	if err != nil {
		return "", err
	}
	meta.Type = string(datasetType)

	// Ingest the dataset info into the metadata storage
	err = storage.IngestDataset(source, meta)
	if err != nil {
		return "", errors.Wrap(err, "unable to ingest metadata")
	}

	log.Infof("ingested metadata for dataset")

	return meta.ID, nil
}

// IngestPostgres ingests a dataset to PG storage.
func IngestPostgres(originalSchemaFile string, schemaFile string, source metadata.DatasetSource,
	config *IngestTaskConfig, verifyMetadata bool, createMetadataTables bool, fallbackMerged bool) error {
	datasetDir, meta, err := loadMetadataForIngest(originalSchemaFile, schemaFile, source, nil, config, verifyMetadata, fallbackMerged)
	if err != nil {
		return err
	}
	mainDR := meta.GetMainDataResource()
	dataDir := path.Join(datasetDir, mainDR.ResPath)

	// Connect to the database.
	postgresConfig := &postgres.Config{
		Password:         config.DatabasePassword,
		User:             config.DatabaseUser,
		Database:         config.Database,
		Host:             config.DatabaseHost,
		Port:             config.DatabasePort,
		BatchSize:        config.DatabaseBatchSize,
		PostgresLogLevel: "error",
	}
	pg, err := postgres.NewDatabase(postgresConfig, true)
	if err != nil {
		return errors.Wrap(err, "unable to initialize a new database")
	}

	dbTable := meta.StorageName
	if createMetadataTables {
		err = pg.CreateSolutionMetadataTables()
		if err != nil {
			return err
		}
	}

	// Drop the current table if requested.
	// Hardcoded the base table name for now.
	pg.DropView(dbTable)
	pg.DropTable(fmt.Sprintf("%s%s", dbTable, baseTableSuffix))
	pg.DropTable(fmt.Sprintf("%s%s", dbTable, explainTableSuffix))

	// Create the database table.
	ds, err := pg.InitializeDataset(meta)
	if err != nil {
		return errors.Wrap(err, "unable to initialize a new dataset")
	}

	err = pg.InitializeTable(dbTable, ds)
	if err != nil {
		return errors.Wrap(err, "unable to initialize a table")
	}

	err = pg.StoreMetadata(dbTable)
	if err != nil {
		return errors.Wrap(err, "unable to store the metadata")
	}

	err = pg.CreateResultTable(dbTable)
	if err != nil {
		return errors.Wrap(err, "unable to create the result table")
	}

	// Load the data.
	log.Infof("inserting rows into database based on data found in %s", dataDir)
	csvFile, err := os.Open(dataDir)
	if err != nil {
		return errors.Wrap(err, "unable to open input data")
	}
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)

	// skip header
	reader.Read()
	count := 0
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "unable to read input line")
		}

		err = pg.AddWordStems(line)
		if err != nil {
			log.Warn(fmt.Sprintf("%v", err))
		}

		err = pg.IngestRow(dbTable, line)
		if err != nil {
			return errors.Wrap(err, "unable to ingest row")
		}

		count = count + 1
		if count%10000 == 0 {
			log.Infof("inserted %d rows so far", count)
		}
	}

	log.Infof("ingesting final rows")
	err = pg.InsertRemainingRows()
	if err != nil {
		return errors.Wrap(err, "unable to ingest last rows")
	}

	log.Infof("all data ingested")

	return nil
}

func loadMetadataForIngest(originalSchemaFile string, schemaFile string, source metadata.DatasetSource,
	origins []*model.DatasetOrigin, config *IngestTaskConfig, verifyMetadata bool, mergedFallback bool) (string, *model.Metadata, error) {
	datasetDir := path.Dir(schemaFile)
	meta, err := metadata.LoadMetadataFromClassification(schemaFile, path.Join(datasetDir, config.ClassificationOutputPathRelative), true, mergedFallback)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to load original schema file")
	}

	if source == metadata.Seed {
		meta.DatasetFolder = path.Base(path.Dir(path.Dir(originalSchemaFile)))
	} else {
		meta.DatasetFolder = path.Base(path.Dir(originalSchemaFile))
	}

	mainDR := meta.GetMainDataResource()
	log.Infof("main DR: %v", mainDR)
	if mainDR == nil {
		for _, dr := range meta.DataResources {
			log.Infof("DR: %v", dr)
		}
	}
	dataDir := path.Join(datasetDir, mainDR.ResPath)
	log.Infof("using %s as data directory (built from %s and %s)", dataDir, datasetDir, mainDR.ResPath)

	// check and fix metadata issues
	if verifyMetadata {
		updated, err := metadata.VerifyAndUpdate(meta, dataDir)
		if err != nil {
			return "", nil, errors.Wrap(err, "unable to fix metadata")
		}

		// store the updated metadata
		if updated {
			extendedOutput := source == metadata.Augmented
			log.Infof("storing updated (extended: %v) metadata to %s", extendedOutput, originalSchemaFile)
			err = metadata.WriteSchema(meta, originalSchemaFile, extendedOutput)
			if err != nil {
				return "", nil, errors.Wrap(err, "unable to store updated metadata")
			}
			log.Infof("updated metadata written to %s", originalSchemaFile)
		}
	}

	err = metadata.LoadImportance(meta, path.Join(datasetDir, config.RankingOutputPathRelative))
	if err != nil {
		log.Warnf("unable to load importance from file: %v", err)
	}

	// load stats
	err = metadata.LoadDatasetStats(meta, dataDir)
	if err != nil {
		log.Warnf("unable to load stats: %v", err)
	}

	// load summary
	metadata.LoadSummaryFromDescription(meta, path.Join(datasetDir, config.SummaryOutputPathRelative))

	// load machine summary
	err = metadata.LoadSummaryMachine(meta, path.Join(datasetDir, config.SummaryMachineOutputPathRelative))
	// NOTE: For now ignore summary errors!
	if err != nil {
		log.Warnf("unable to load machine summary: %v", err)
	}

	// set the origin
	if origins != nil {
		meta.DatasetOrigins = origins
	}

	return datasetDir, meta, nil
}

func matchDataset(storage api.MetadataStorage, meta *model.Metadata) (string, error) {
	// load the datasets from ES.
	datasets, err := storage.FetchDatasets(true, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to fetch datasets for matching")
	}

	// See if any of the loaded datasets match.
	for _, dataset := range datasets {
		if dataset.ID == meta.ID {
			return dataset.Name, nil
		}
		variables := make([]string, 0)
		for _, v := range dataset.Variables {
			variables = append(variables, v.Name)
		}
		if metadata.DatasetMatches(meta, variables) {
			return dataset.Name, nil
		}
	}

	// No matching set.
	return "", nil
}

func deleteDataset(name string, pg *postgres.Database, meta api.MetadataStorage) error {
	success := false
	for i := 0; i < 10 && !success; i++ {
		err := meta.DeleteDataset(name)
		if err != nil {
			log.Error(err)
		} else {
			success = true
		}
	}

	if success {
		pg.DeleteDataset(name)
	}

	return nil
}

func getUniqueDatasetName(meta *model.Metadata, storage api.MetadataStorage) (string, error) {
	// create a unique name if the current name is already in use
	datasets, err := storage.FetchDatasets(false, false)
	if err != nil {
		return "", err
	}

	datasetNames := make([]string, 0)
	for _, ds := range datasets {
		datasetNames = append(datasetNames, ds.Name)
	}

	return getUniqueString(meta.Name, datasetNames), nil
}
