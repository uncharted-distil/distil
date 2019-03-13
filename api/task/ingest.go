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
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-ingest/conf"
	"github.com/uncharted-distil/distil-ingest/metadata"
	"github.com/uncharted-distil/distil-ingest/postgres"
	log "github.com/unchartedsoftware/plog"
	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
)

const (
	rankingFilename = "rank-no-missing.csv"
	baseTableSuffix = "_base"
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
	SummaryOutputPathRelative          string
	SummaryMachineOutputPathRelative   string
	SummaryEnabled                     bool
	ESEndpoint                         string
	ESTimeout                          int
	ESDatasetPrefix                    string
	HardFail                           bool
}

// IngestDataset executes the complete ingest process for the specified dataset.
func IngestDataset(datasetSource metadata.DatasetSource, metaCtor api.MetadataStorageCtor, index string, dataset string, config *IngestTaskConfig) error {
	// Set the probability threshold
	metadata.SetTypeProbabilityThreshold(config.ClassificationProbabilityThreshold)

	storage, err := metaCtor()
	if err != nil {
		return errors.Wrap(err, "unable to initialize metadata storage")
	}

	sourceFolder := env.ResolvePath(datasetSource, dataset)

	originalSchemaFile := path.Join(sourceFolder, config.SchemaPathRelative)
	latestSchemaOutput := originalSchemaFile

	output, err := Merge(datasetSource, latestSchemaOutput, index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to merge all data into a single file")
	}
	latestSchemaOutput = output
	log.Infof("finished merging the dataset")

	output, err = Clean(datasetSource, latestSchemaOutput, index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to clean all data")
	}
	latestSchemaOutput = output
	log.Infof("finished cleaning the dataset")

	if config.ClusteringEnabled {
		output, err = Cluster(datasetSource, latestSchemaOutput, index, dataset, config)
		if err != nil {
			if config.HardFail {
				return errors.Wrap(err, "unable to cluster all data")
			}
			log.Errorf("unable to cluster all data: %v", err)
		} else {
			latestSchemaOutput = output
		}
		log.Infof("finished clustering the dataset")
	}

	output, err = Featurize(datasetSource, latestSchemaOutput, index, dataset, config)
	if err != nil {
		if config.HardFail {
			return errors.Wrap(err, "unable to featurize all data")
		}
		log.Errorf("unable to featurize all data: %v", err)
	} else {
		latestSchemaOutput = output
	}
	log.Infof("finished featurizing the dataset")

	err = Classify(latestSchemaOutput, index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to classify fields")
	}
	log.Infof("finished classifying the dataset")

	err = Rank(latestSchemaOutput, index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to rank field importance")
	}
	log.Infof("finished ranking the dataset")

	if config.SummaryEnabled {
		err = Summarize(latestSchemaOutput, index, dataset, config)
		log.Infof("finished summarizing the dataset")
		if err != nil {
			if config.HardFail {
				return errors.Wrap(err, "unable to summarize the dataset")
			}
			log.Errorf("unable to summarize the dataset: %v", err)
		}
	} else {
		log.Infof("summarization disabled")
	}

	if config.GeocodingEnabled {
		output, err = GeocodeForwardDataset(datasetSource, latestSchemaOutput, index, dataset, config)
		if err != nil {
			return errors.Wrap(err, "unable to geocode all data")
		}
		latestSchemaOutput = output
		log.Infof("finished geocoding the dataset")
	}

	err = Ingest(originalSchemaFile, latestSchemaOutput, storage, index, dataset, datasetSource, config)
	if err != nil {
		return errors.Wrap(err, "unable to ingest ranked data")
	}
	log.Infof("finished ingesting the dataset")

	return nil
}

// Ingest the metadata to ES and the data to Postgres.
func Ingest(originalSchemaFile string, schemaFile string, storage api.MetadataStorage, index string, dataset string, source metadata.DatasetSource, config *IngestTaskConfig) error {
	datasetDir := path.Dir(schemaFile)
	meta, err := metadata.LoadMetadataFromClassification(schemaFile, path.Join(datasetDir, config.ClassificationOutputPathRelative), true)
	if err != nil {
		return errors.Wrap(err, "unable to load original schema file")
	}
	meta.DatasetFolder = path.Base(path.Dir(originalSchemaFile))
	dataDir := path.Join(datasetDir, meta.DataResources[0].ResPath)

	err = metadata.LoadImportance(meta, path.Join(datasetDir, config.RankingOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load importance from file")
	}

	// load stats
	err = metadata.LoadDatasetStats(meta, dataDir)
	if err != nil {
		return errors.Wrap(err, "unable to load stats")
	}

	// load summary
	err = metadata.LoadSummaryFromDescription(meta, path.Join(datasetDir, config.SummaryOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load summary")
	}

	// load machine summary
	err = metadata.LoadSummaryMachine(meta, path.Join(datasetDir, config.SummaryMachineOutputPathRelative))
	// NOTE: For now ignore summary errors!
	if err != nil {
		log.Errorf("unable to load machine summary: %v", err)
	}

	// create elasticsearch client
	elasticClient, err := elastic.NewClient(
		elastic.SetURL(config.ESEndpoint),
		elastic.SetHttpClient(&http.Client{Timeout: time.Second * time.Duration(config.ESTimeout)}),
		elastic.SetMaxRetries(10),
		elastic.SetSniff(false),
		elastic.SetGzip(true))
	if err != nil {
		return errors.Wrap(err, "unable to initialize elastic client")
	}

	// Connect to the database.
	postgresConfig := &conf.Conf{
		DBPassword:  config.DatabasePassword,
		DBUser:      config.DatabaseUser,
		Database:    config.Database,
		DBHost:      config.DatabaseHost,
		DBPort:      config.DatabasePort,
		DBBatchSize: 1000,
	}
	pg, err := postgres.NewDatabase(postgresConfig)
	if err != nil {
		return errors.Wrap(err, "unable to initialize a new database")
	}

	// Check for existing dataset
	match, err := matchDataset(storage, meta, index)
	// Ignore the error for now as if this fails we still want ingest to succeed.
	if err != nil {
		log.Error(err)
	}
	if match != "" {
		log.Infof("Matched %s to dataset %s", meta.Name, match)
		err = deleteDataset(match, index, pg, elasticClient)
		log.Infof("Deleted dataset %s", match)
	}

	// ingest the metadata
	// Create the metadata index if it doesn't exist
	err = metadata.CreateMetadataIndex(elasticClient, index, false)
	if err != nil {
		return errors.Wrap(err, "unable to create metadata index")
	}

	// Ingest the dataset info into the metadata index
	err = metadata.IngestMetadata(elasticClient, index, config.ESDatasetPrefix, source, meta)
	if err != nil {
		return errors.Wrap(err, "unable to ingest metadata")
	}

	dbTable := meta.StorageName

	// Drop the current table if requested.
	// Hardcoded the base table name for now.
	pg.DropView(dbTable)
	pg.DropTable(fmt.Sprintf("%s%s", dbTable, baseTableSuffix))

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

	err = pg.CreateSolutionMetadataTables()
	if err != nil {
		return errors.Wrap(err, "unable to create solution metadata tables")
	}

	// Load the data.
	log.Infof("inserting rows into database based on data found in %s", dataDir)
	reader, err := os.Open(dataDir)
	scanner := bufio.NewScanner(reader)

	// skip header
	scanner.Scan()
	count := 0
	for scanner.Scan() {
		line := scanner.Text()

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

func matchDataset(storage api.MetadataStorage, meta *model.Metadata, index string) (string, error) {
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

func deleteDataset(name string, index string, pg *postgres.Database, es *elastic.Client) error {
	id := name
	success := false
	for i := 0; i < 10 && !success; i++ {
		_, err := es.Delete().Index(index).Id(id).Type("metadata").Do(context.Background())
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
