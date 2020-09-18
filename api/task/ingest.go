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
	"fmt"
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
	ClusteringKMeans                   bool
	FeaturizationEnabled               bool
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
	SampleRowLimit                     int
}

// IngestSteps is a collection of parameters that specify ingest behaviour.
type IngestSteps struct {
	ClassificationOverwrite bool
	RawGrouping             map[string]interface{}
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
		ClusteringKMeans:                   config.ClusteringKMeans,
		FeaturizationEnabled:               config.FeaturizationEnabled,
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
		SampleRowLimit:                     config.IngestSampleRowLimit,
	}
}

// IngestResult captures the result of a dataset ingest process.
type IngestResult struct {
	DatasetID string
	Sampled   bool
	RowCount  int
}

// IngestDataset executes the complete ingest process for the specified dataset.
func IngestDataset(datasetSource metadata.DatasetSource, dataCtor api.DataStorageCtor, metaCtor api.MetadataStorageCtor,
	dataset string, origins []*model.DatasetOrigin, datasetType api.DatasetType, config *IngestTaskConfig, steps *IngestSteps) (*IngestResult, error) {
	// Set the probability threshold
	metadata.SetTypeProbabilityThreshold(config.ClassificationProbabilityThreshold)

	metaStorage, err := metaCtor()
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize metadata storage")
	}

	dataStorage, err := dataCtor()
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize data storage")
	}

	sourceFolder := env.ResolvePath(datasetSource, dataset)

	originalSchemaFile := path.Join(sourceFolder, config.SchemaPathRelative)
	latestSchemaOutput := originalSchemaFile

	output := latestSchemaOutput
	if config.ClusteringEnabled {
		output, err = ClusterDataset(latestSchemaOutput, dataset, config)
		if err != nil {
			if config.HardFail {
				return nil, errors.Wrap(err, "unable to cluster all data")
			}
			log.Errorf("unable to cluster all data: %v", err)
		} else {
			latestSchemaOutput = output
		}
		log.Infof("finished clustering the dataset")
	}

	output, err = Merge(latestSchemaOutput, dataset, config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to merge all data into a single file")
	}
	latestSchemaOutput = output
	log.Infof("finished merging the dataset")

	output, err = Clean(latestSchemaOutput, dataset, config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to clean all data")
	}
	latestSchemaOutput = output
	log.Infof("finished cleaning the dataset")

	definitiveClassification := false
	if config.ClassificationEnabled {
		if steps.ClassificationOverwrite || !classificationExists(latestSchemaOutput, config) {
			_, err = Classify(latestSchemaOutput, dataset, config)
			if err != nil {
				if config.HardFail {
					return nil, errors.Wrap(err, "unable to classify fields")
				}
				log.Errorf("unable to classify fields: %+v", err)
			}
			log.Infof("finished classifying the dataset")
		} else {
			definitiveClassification = true
			log.Infof("skipping classification because it already exists")
		}
	} else {
		log.Infof("classification disabled")
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
				return nil, errors.Wrap(err, "unable to summarize the dataset")
			}
			log.Errorf("unable to summarize the dataset: %v", err)
		}
	} else {
		log.Infof("summarization disabled")
	}

	if config.GeocodingEnabled {
		output, err = GeocodeForwardDataset(latestSchemaOutput, dataset, config)
		if err != nil {
			return nil, errors.Wrap(err, "unable to geocode all data")
		}
		latestSchemaOutput = output
		log.Infof("finished geocoding the dataset")
	}

	// not sure if better to call canSample here, or as the first part of the sample step
	sampled := false
	rowCount := 0
	if canSample(latestSchemaOutput, config) {
		log.Infof("sampling dataset")
		latestSchemaOutput, sampled, rowCount, err = Sample(originalSchemaFile, latestSchemaOutput, dataset, config)
		if err != nil {
			return nil, errors.Wrap(err, "unable to sample dataset")
		}
		log.Infof("finished sampling dataset")
	}

	datasetID, err := Ingest(originalSchemaFile, latestSchemaOutput, dataStorage, metaStorage, dataset, datasetSource, origins, datasetType, config, true, !definitiveClassification, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to ingest ranked data")
	}
	log.Infof("finished ingesting the dataset")

	// set the known grouping information
	if steps.RawGrouping != nil {
		log.Infof("creating groupings in metadata")
		err = SetGroups(datasetID, steps.RawGrouping, metaStorage, config)
		if err != nil {
			return nil, errors.Wrap(err, "unable to set grouping")
		}
		log.Infof("done creating groupings in metadata")
	}

	// featurize dataset for downstream efficiencies
	if config.FeaturizationEnabled && canFeaturize(dataset, metaStorage) {
		_, featurizedDatasetPath, err := FeaturizeDataset(originalSchemaFile, latestSchemaOutput, dataset, metaStorage, config)
		if err != nil {
			return nil, errors.Wrap(err, "unable to featurize dataset")
		}
		log.Infof("finished featurizing the dataset")
		ingestedDataset, err := metaStorage.FetchDataset(dataset, true, true)
		if err != nil {
			return nil, errors.Wrap(err, "unable to load metadata")
		}
		ingestedDataset.LearningDataset = featurizedDatasetPath
		err = metaStorage.UpdateDataset(ingestedDataset)
		if err != nil {
			return nil, errors.Wrap(err, "unable to store updated metadata")
		}
	}

	// updating extremas is optional
	err = UpdateExtremas(datasetID, metaStorage, dataStorage)
	if err != nil {
		log.Errorf("unable to update extremas ranked data: %v", err)
	}
	log.Infof("finished updating extremas")

	return &IngestResult{
		DatasetID: datasetID,
		Sampled:   sampled,
		RowCount:  rowCount,
	}, nil
}

// Ingest the metadata to ES and the data to Postgres.
func Ingest(originalSchemaFile string, schemaFile string, data api.DataStorage, storage api.MetadataStorage, dataset string, source metadata.DatasetSource,
	origins []*model.DatasetOrigin, datasetType api.DatasetType, config *IngestTaskConfig, checkMatch bool, verifyMetadata bool, fallbackMerged bool) (string, error) {
	_, meta, err := loadMetadataForIngest(originalSchemaFile, schemaFile, source, nil, config, verifyMetadata, fallbackMerged)
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
			err = datasetStorage.WriteMetadata(originalSchemaFile, meta, extendedOutput)
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
	_, err = IngestMetadata(originalSchemaFile, schemaFile, data, storage, source, origins, datasetType, config, verifyMetadata, fallbackMerged)
	if err != nil {
		return "", err
	}

	// ingest the data
	err = IngestPostgres(originalSchemaFile, schemaFile, source, config, verifyMetadata, false, fallbackMerged, storage)
	
	if err != nil {
		return "", err
	}

	return meta.ID, nil
}

// IngestMetadata ingests the data to ES.
func IngestMetadata(originalSchemaFile string, schemaFile string, data api.DataStorage, storage api.MetadataStorage, source metadata.DatasetSource,
	origins []*model.DatasetOrigin, datasetType api.DatasetType, config *IngestTaskConfig, verifyMetadata bool, fallbackMerged bool) (string, error) {
	_, meta, err := loadMetadataForIngest(originalSchemaFile, schemaFile, source, origins, config, verifyMetadata, fallbackMerged)
	if err != nil {
		return "", err
	}
	meta.Type = string(datasetType)

	if data != nil {
		storageName, err := data.GetStorageName(meta.ID)
		if err != nil {
			return "", err
		}
		if meta.StorageName != storageName {
			log.Infof("updating storage name in metadata from %s to %s", meta.StorageName, storageName)
			meta.StorageName = storageName
			err = datasetStorage.WriteMetadata(schemaFile, meta, true)
			if err != nil {
				return "", err
			}
		}
	}

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
	config *IngestTaskConfig, verifyMetadata bool, createMetadataTables bool, fallbackMerged bool, storage api.MetadataStorage) error {
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
	data, err := datasetStorage.ReadData(dataDir)
	if err != nil {
		return errors.Wrap(err, "unable to read input data")
	}

	// skip header
	data = data[1:]
	count := 0
	for _, line := range data {
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
	// verfiy the data type for the columns
	verifyData(pg, meta, storage, dbTable)
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
	dataDir := path.Join(datasetDir, mainDR.ResPath)
	log.Infof("using %s as data directory (built from %s and %s)", dataDir, datasetDir, mainDR.ResPath)

	// check and fix metadata issues
	if verifyMetadata {
		updated, err := metadata.VerifyAndUpdate(meta, dataDir, source)
		if err != nil {
			return "", nil, errors.Wrap(err, "unable to fix metadata")
		}

		// store the updated metadata
		if updated {
			extendedOutput := source == metadata.Augmented
			log.Infof("storing updated (extended: %v) metadata to %s", extendedOutput, originalSchemaFile)
			err = datasetStorage.WriteMetadata(originalSchemaFile, meta, extendedOutput)
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

	// load stats and adjust for header row (may want to do in the LoadDatasetStats function)
	err = metadata.LoadDatasetStats(meta, dataDir)
	if err != nil {
		log.Warnf("unable to load stats: %v", err)
	}
	meta.NumRows = meta.NumRows - 1

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

func verifyData(pg *postgres.Database, meta *model.Metadata, storage api.MetadataStorage, tableName string) {
	validTypes := postgres.GetValidTypes()
	ds, err:=storage.FetchDataset(meta.Name, false, true)
	if err != nil{
		return
	}
	//removing double and geometry for now
	double:="double precision"
	geometry:="geometry"
	mainValidTypes := []string{}
	provenance:="postgres-valid"
	for i := range validTypes{
		if validTypes[i] != double && validTypes[i] != geometry{
			mainValidTypes = append(mainValidTypes, validTypes[i])
		}
	}
	// if view can succeed on column add potential type to column
	for i := range ds.Variables{
		suggestedMap := make(map[string]bool)
		for t:=range ds.Variables[i].SuggestedTypes{
			suggestedMap[ds.Variables[i].SuggestedTypes[t].Type]=true
		}
		for j := range mainValidTypes{
			isValid, err := pg.IsColumnType(tableName, ds.Variables[i], mainValidTypes[j])
			if err != nil{
				continue
			}
			if isValid {
				d3mTypes,err:=postgres.MapPostgresTypeToD3MType(mainValidTypes[j])
				if err != nil{
					continue
				}
				for k := range d3mTypes{
					// this could be moved up to an exit case above but a lot of the upconversion from pg to d3m types involves multiple results
					if suggestedMap[d3mTypes[k]] {
						continue
					}
					suggestedType := model.SuggestedType{Probability:0, Type:d3mTypes[k], Provenance:provenance}
					ds.Variables[i].SuggestedTypes = append(ds.Variables[i].SuggestedTypes, &suggestedType)
				}
			}
		}
	}
	// save changes
	storage.UpdateDataset(ds)
	
	return
}
