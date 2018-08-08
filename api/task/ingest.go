package task

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"gopkg.in/olivere/elastic.v5"

	"github.com/unchartedsoftware/distil-ingest/conf"
	"github.com/unchartedsoftware/distil-ingest/merge"
	"github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil-ingest/postgres"
	"github.com/unchartedsoftware/distil/api/model"
)

const (
	rankingFilename = "rank-no-missing.csv"
	baseTableSuffix = "_base"
)

var (
	classify  = ClassifyContainer
	rank      = RankContainer
	summarize = SummarizeContainer
	featurize = FeaturizeContainer
	cluster   = ClusterContainer
)

// Classify function that will classify variables within the dataset.
type Classify func(index string, dataset string, config *IngestTaskConfig) error

// Rank function that will rank relative variable importance within the dataset.
type Rank func(index string, dataset string, config *IngestTaskConfig) error

// Summarize function that will provide a summary of the dataset.
type Summarize func(index string, dataset string, config *IngestTaskConfig) error

// Featurize function that will extract features from dataset variables
// and add them to the dataset.
type Featurize func(index string, dataset string, config *IngestTaskConfig) error

// Cluster function that will cluster features from dataset variables
// and add them to the dataset.
type Cluster func(index string, dataset string, config *IngestTaskConfig) error

// SetClassify sets the classification function to use.
func SetClassify(classificationFunc Classify) {
	classify = classificationFunc
}

// SetRank sets the ranking function to use.
func SetRank(rankFunc Rank) {
	rank = rankFunc
}

// SetSummarize sets the summarization function to use.
func SetSummarize(summarizeFunc Summarize) {
	summarize = summarizeFunc
}

// SetFeaturize sets the featurization function to use.
func SetFeaturize(featurizeFunc Featurize) {
	featurize = featurizeFunc
}

// SetCluster sets the clustering function to use.
func SetCluster(clusterFunc Cluster) {
	cluster = clusterFunc
}

// IngestTaskConfig captures the necessary configuration for an data ingest.
type IngestTaskConfig struct {
	ContainerDataPath                  string
	TmpDataPath                        string
	DataPathRelative                   string
	DatasetFolderSuffix                string
	MediaPath                          string
	HasHeader                          bool
	ClusteringRESTEndpoint             string
	ClusteringFunctionName             string
	ClusteringOutputDataRelative       string
	ClusteringOutputSchemaRelative     string
	FeaturizationRESTEndpoint          string
	FeaturizationFunctionName          string
	FeaturizationOutputDataRelative    string
	FeaturizationOutputSchemaRelative  string
	MergedOutputPathRelative           string
	MergedOutputSchemaPathRelative     string
	SchemaPathRelative                 string
	ClassificationRESTEndpoint         string
	ClassificationFunctionName         string
	ClassificationOutputPathRelative   string
	ClassificationProbabilityThreshold float64
	ClassificationEnabled              bool
	RankingRESTEndpoint                string
	RankingFunctionName                string
	RankingOutputPathRelative          string
	RankingRowLimit                    int
	DatabasePassword                   string
	DatabaseUser                       string
	Database                           string
	DatabaseHost                       string
	DatabasePort                       int
	SummaryOutputPathRelative          string
	SummaryMachineOutputPathRelative   string
	SummaryRESTEndpoint                string
	SummaryFunctionName                string
	ESEndpoint                         string
	ESTimeout                          int
	ESDatasetPrefix                    string
}

func (c *IngestTaskConfig) getRootPath(dataset string) string {
	return fmt.Sprintf("%s/%s/%s%s", c.ContainerDataPath, dataset, dataset, c.DatasetFolderSuffix)
}

func (c *IngestTaskConfig) getAbsolutePath(relativePath string) string {
	return fmt.Sprintf("%s/%s", c.ContainerDataPath, relativePath)
}

func (c *IngestTaskConfig) getTmpAbsolutePath(relativePath string) string {
	return fmt.Sprintf("%s/%s", c.TmpDataPath, relativePath)
}

func (c *IngestTaskConfig) getRawDataPath() string {
	return fmt.Sprintf("%s/", c.ContainerDataPath)
}

// IngestDataset executes the complete ingest process for the specified dataset.
func IngestDataset(metaCtor model.MetadataStorageCtor, index string, dataset string, config *IngestTaskConfig) error {
	// Make sure the temp data directory exists.
	tmpPath := path.Dir(config.getTmpAbsolutePath(config.MergedOutputSchemaPathRelative))
	os.MkdirAll(tmpPath, os.ModePerm)

	// Set the probability threshold
	metadata.SetTypeProbabilityThreshold(config.ClassificationProbabilityThreshold)

	storage, err := metaCtor()
	if err != nil {
		return errors.Wrap(err, "unable to initialize metadata storage")
	}

	err = cluster(index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to cluster all data")
	}
	log.Infof("finished clustering the dataset")

	err = featurize(index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to featurize all data")
	}
	log.Infof("finished featurizing the dataset")

	err = Merge(index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to merge all data into a single file")
	}
	log.Infof("finished merging the dataset")

	err = classify(index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to classify fields")
	}
	log.Infof("finished classifying the dataset")

	err = rank(index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to rank field importance")
	}
	log.Infof("finished ranking the dataset")

	err = summarize(index, dataset, config)
	log.Infof("finished summarizing the dataset")
	// NOTE: For now ignore summary errors!
	//if err != nil {
	//	return errors.Wrap(err, "unable to summarize the dataset")
	//}

	err = Ingest(storage, index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to ingest ranked data")
	}
	log.Infof("finished ingestig the dataset")

	return nil
}

// Merge combines all the source data files into a single datafile.
func Merge(index string, dataset string, config *IngestTaskConfig) error {
	// load the metadata from schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(config.getTmpAbsolutePath(config.FeaturizationOutputSchemaRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load metadata schema")
	}

	// merge file links in dataset
	mergedDR, output, err := merge.InjectFileLinksFromFile(meta, config.getTmpAbsolutePath(config.FeaturizationOutputDataRelative), config.getRawDataPath(), config.MergedOutputPathRelative, config.HasHeader)
	if err != nil {
		return errors.Wrap(err, "unable to merge linked files")
	}

	// write copy to disk
	err = ioutil.WriteFile(config.getTmpAbsolutePath(config.MergedOutputPathRelative), output, 0644)
	if err != nil {
		return errors.Wrap(err, "unable to write merged data")
	}

	// write merged metadata out to disk
	err = meta.WriteMergedSchema(config.getTmpAbsolutePath(config.MergedOutputSchemaPathRelative), mergedDR)
	if err != nil {
		return errors.Wrap(err, "unable to write merged schema")
	}

	return nil
}

// Ingest the metadata to ES and the data to Postgres.
func Ingest(storage model.MetadataStorage, index string, dataset string, config *IngestTaskConfig) error {
	meta, err := metadata.LoadMetadataFromClassification(
		config.getTmpAbsolutePath(config.MergedOutputSchemaPathRelative),
		config.getTmpAbsolutePath(config.ClassificationOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load metadata")
	}

	err = meta.LoadImportance(config.getTmpAbsolutePath(config.RankingOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load importance from file")
	}

	// load stats
	err = meta.LoadDatasetStats(config.getTmpAbsolutePath(config.MergedOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load stats")
	}

	// load summary
	err = meta.LoadSummaryFromDescription(config.getTmpAbsolutePath(config.SummaryOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load summary")
	}

	// load machine summary
	err = meta.LoadSummaryMachine(config.getTmpAbsolutePath(config.SummaryMachineOutputPathRelative))
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

func fixDatasetIDName(meta *metadata.Metadata) {
	// Train dataset ID & name need to be adjusted to fit the expected format.
	// The ID MUST end in _dataset, and the name should be representative.
	if isTrainDataset(meta) {
		meta.ID = strings.TrimSuffix(meta.ID, "_TRAIN")
	}
}

func isTrainDataset(meta *metadata.Metadata) bool {
	return strings.HasSuffix(meta.ID, "_TRAIN")
}

func matchDataset(storage model.MetadataStorage, meta *metadata.Metadata, index string) (string, error) {
	// load the datasets from ES.
	datasets, err := storage.FetchDatasets(true, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to fetch datasets for matching")
	}

	// See if any of the loaded datasets match.
	for _, dataset := range datasets {
		variables := make([]string, 0)
		for _, v := range dataset.Variables {
			variables = append(variables, v.Key)
		}
		if meta.DatasetMatches(variables) {
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
