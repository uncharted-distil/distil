package task

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-ingest/conf"
	"github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil-ingest/postgres"
	log "github.com/unchartedsoftware/plog"
	elastic "gopkg.in/olivere/elastic.v5"

	api "github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util"
)

const (
	rankingFilename = "rank-no-missing.csv"
	baseTableSuffix = "_base"
)

// IngestTaskConfig captures the necessary configuration for an data ingest.
type IngestTaskConfig struct {
	Resolver                           *util.PathResolver
	HasHeader                          bool
	ClusteringOutputDataRelative       string
	ClusteringOutputSchemaRelative     string
	ClusteringEnabled                  bool
	FeaturizationOutputDataRelative    string
	FeaturizationOutputSchemaRelative  string
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
	ESEndpoint                         string
	ESTimeout                          int
	ESDatasetPrefix                    string
	HardFail                           bool
}

func (c *IngestTaskConfig) getAbsolutePath(relativePath string) string {
	return c.Resolver.ResolveInputAbsolute(relativePath)
}

func (c *IngestTaskConfig) getTmpAbsolutePath(relativePath string) string {
	return c.Resolver.ResolveOutputAbsolute(relativePath)
}

// IngestDataset executes the complete ingest process for the specified dataset.
func IngestDataset(metaCtor api.MetadataStorageCtor, index string, dataset string, config *IngestTaskConfig) error {
	// Set the probability threshold
	metadata.SetTypeProbabilityThreshold(config.ClassificationProbabilityThreshold)

	storage, err := metaCtor()
	if err != nil {
		return errors.Wrap(err, "unable to initialize metadata storage")
	}

	latestSchemaOutput := config.getAbsolutePath(config.SchemaPathRelative)

	if config.ClusteringEnabled {
		err := Cluster(index, dataset, config)
		if err != nil {
			if config.HardFail {
				return errors.Wrap(err, "unable to cluster all data")
			}
			log.Errorf("unable to cluster all data: %v", err)
		} else {
			latestSchemaOutput = config.getTmpAbsolutePath(config.ClusteringOutputSchemaRelative)
		}
		log.Infof("finished clustering the dataset")
	}

	err = Featurize(latestSchemaOutput, index, dataset, config)
	if err != nil {
		if config.HardFail {
			return errors.Wrap(err, "unable to featurize all data")
		}
		log.Errorf("unable to featurize all data: %v", err)
	} else {
		latestSchemaOutput = config.getTmpAbsolutePath(config.FeaturizationOutputSchemaRelative)
	}
	log.Infof("finished featurizing the dataset")

	err = Merge(latestSchemaOutput, index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to merge all data into a single file")
	}
	log.Infof("finished merging the dataset")

	err = Classify(index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to classify fields")
	}
	log.Infof("finished classifying the dataset")

	err = Rank(index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to rank field importance")
	}
	log.Infof("finished ranking the dataset")

	err = Summarize(index, dataset, config)
	log.Infof("finished summarizing the dataset")
	// NOTE: For now ignore summary errors!
	if err != nil {
		if config.HardFail {
			return errors.Wrap(err, "unable to summarize the dataset")
		}
		log.Errorf("unable to summarize the dataset: %v", err)
	}

	err = Ingest(storage, index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to ingest ranked data")
	}
	log.Infof("finished ingestig the dataset")

	return nil
}

// Ingest the metadata to ES and the data to Postgres.
func Ingest(storage api.MetadataStorage, index string, dataset string, config *IngestTaskConfig) error {
	meta, err := metadata.LoadMetadataFromClassification(
		config.getTmpAbsolutePath(config.MergedOutputSchemaPathRelative),
		config.getTmpAbsolutePath(config.ClassificationOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load metadata")
	}

	err = metadata.LoadImportance(meta, config.getTmpAbsolutePath(config.RankingOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load importance from file")
	}

	// load stats
	err = metadata.LoadDatasetStats(meta, config.getTmpAbsolutePath(config.MergedOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load stats")
	}

	// load summary
	err = metadata.LoadSummaryFromDescription(meta, config.getTmpAbsolutePath(config.SummaryOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load summary")
	}

	// load machine summary
	err = metadata.LoadSummaryMachine(meta, config.getTmpAbsolutePath(config.SummaryMachineOutputPathRelative))
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
	err = metadata.IngestMetadata(elasticClient, index, config.ESDatasetPrefix, meta)
	if err != nil {
		return errors.Wrap(err, "unable to ingest metadata")
	}

	dbTable := meta.ID
	dbTable = fmt.Sprintf("%s%s", config.ESDatasetPrefix, dbTable)

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
	log.Infof("inserting rows into database")
	reader, err := os.Open(config.getTmpAbsolutePath(config.MergedOutputPathRelative))
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

func fixDatasetIDName(meta *model.Metadata) {
	// Train dataset ID & name need to be adjusted to fit the expected format.
	// The ID MUST end in _dataset, and the name should be representative.
	if isTrainDataset(meta) {
		meta.ID = strings.TrimSuffix(meta.ID, "_TRAIN")
	}
}

func isTrainDataset(meta *model.Metadata) bool {
	return strings.HasSuffix(meta.ID, "_TRAIN")
}

func matchDataset(storage api.MetadataStorage, meta *model.Metadata, index string) (string, error) {
	// load the datasets from ES.
	datasets, err := storage.FetchDatasets(true, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to fetch datasets for matching")
	}

	// See if any of the loaded datasets match.
	for _, dataset := range datasets {
		variables := make([]string, 0)
		for _, v := range dataset.Variables {
			variables = append(variables, v.Name)
		}
		if metadata.DatasetMatches(meta, variables) {
			// Return the name of the matching set.
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

func copyFileContents(source string, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return errors.Wrap(err, "unable to open source")
	}
	defer in.Close()
	out, err := os.Create(destination)
	if err != nil {
		return errors.Wrap(err, "unable to open destination")
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return errors.Wrap(err, "unable to copy data")
	}
	err = out.Sync()
	if err != nil {
		return errors.Wrap(err, "unable to finalize copy")
	}

	return nil
}

func translateSchemaRelativeToAbsoluteFilename(schemalFilename string, dataFilename string) string {
	return path.Join(path.Dir(schemalFilename), dataFilename)
}

func createContainingDirs(filePath string) error {
	dirToCreate := filepath.Dir(filePath)
	if dirToCreate != "/" && dirToCreate != "." {
		err := os.MkdirAll(dirToCreate, 0777)
		if err != nil {
			return errors.Wrap(err, "unable to create containing directory")
		}
	}

	return nil
}
