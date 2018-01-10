package task

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v5"

	"github.com/unchartedsoftware/distil-ingest/conf"
	"github.com/unchartedsoftware/distil-ingest/merge"
	"github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil-ingest/postgres"
	"github.com/unchartedsoftware/distil-ingest/rest"
)

type ImportTaskConfig struct {
	ContainerDataPath                string
	Dataset                          string
	DataPathRelative                 string
	DatasetFolderSuffix              string
	HasHeader                        bool
	MergedOutputPathRelative         string
	MergedOutputSchemaPathRelative   string
	SchemaPathRelative               string
	RESTBaseEndpoint                 string
	ClassificationFunctionName       string
	ClassificationOutputPathRelative string
	RankingFunctionName              string
	RankingOutputPathRelative        string
	DatabasePassword                 string
	DatabaseUser                     string
	Database                         string
	DatabaseTable                    string
	SummaryOutputPathRelative        string
	ESEndpoint                       string
	ESTimeout                        int
	ESMetadataIndexName              string
	ESDatasetPrefix                  string
}

func (c *ImportTaskConfig) getRootPath() string {
	return fmt.Sprintf("%s/%s/%s%s", c.ContainerDataPath, c.Dataset, c.Dataset)
}

func (c *ImportTaskConfig) getAbsolutePath(relativePath string) string {
	return fmt.Sprintf("%s/%s", c.getRootPath(), relativePath)
}

func (c *ImportTaskConfig) getRawDataPath() string {
	return fmt.Sprintf("%s/", c.getRootPath())
}

func Merge(config *ImportTaskConfig) error {
	// load the metadata from schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(config.getAbsolutePath(config.SchemaPathRelative))
	if err != nil {
		errors.Wrap(err, "unable to load metadata schema")
	}

	// merge file links in dataset
	mergedDR, output, err := merge.InjectFileLinksFromFile(meta, config.getAbsolutePath(config.DataPathRelative), config.getRawDataPath(), config.HasHeader)
	if err != nil {
		errors.Wrap(err, "unable to merge linked files")
	}

	// write copy to disk
	err = ioutil.WriteFile(config.getAbsolutePath(config.MergedOutputPathRelative), output, 0644)
	if err != nil {
		errors.Wrap(err, "unable to write merged data")
	}

	// write merged metadata out to disk
	err = meta.WriteMergedSchema(config.getAbsolutePath(config.MergedOutputSchemaPathRelative), mergedDR)
	if err != nil {
		errors.Wrap(err, "unable to write merged schema")
	}

	return nil
}

func Classify(config *ImportTaskConfig) error {
	client := rest.NewClient(config.RESTBaseEndpoint)

	// create classifier
	classifier := rest.NewClassifier(config.ClassificationFunctionName, client)

	// classify the file
	classification, err := classifier.ClassifyFile(config.getAbsolutePath(config.MergedOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to classify dataset")
	}

	// marshall result
	bytes, err := json.MarshalIndent(classification, "", "    ")
	if err != nil {
		return errors.Wrap(err, "unable to serialize classification result")
	}
	// write to file
	err = ioutil.WriteFile(config.getAbsolutePath(config.ClassificationOutputPathRelative), bytes, 0644)
	if err != nil {
		errors.Wrap(err, "unable to store classification result")
	}

	return nil
}

func Rank(config *ImportTaskConfig) error {
	// create ranker
	client := rest.NewClient(config.RESTBaseEndpoint)
	ranker := rest.NewRanker(config.RankingFunctionName, client)

	// get the importance from the REST interface
	importance, err := ranker.RankFile(config.getAbsolutePath(config.MergedOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to rank importance file")
	}

	// marshall result
	bytes, err := json.MarshalIndent(importance, "", "    ")
	if err != nil {
		return errors.Wrap(err, "unable to marshall importance ranking result")
	}

	// write to file
	outputPath := config.getAbsolutePath(config.RankingOutputPathRelative)
	err = ioutil.WriteFile(outputPath, bytes, 0644)
	if err != nil {
		return errors.Wrapf(err, "unable to write importance ranking to '%s'", outputPath)
	}

	return nil
}

func Ingest(config *ImportTaskConfig) error {
	meta, err := metadata.LoadMetadataFromClassification(
		config.getAbsolutePath(config.MergedOutputSchemaPathRelative),
		config.getAbsolutePath(config.ClassificationOutputPathRelative))
	if err != nil {
		errors.Wrap(err, "unable to load metadata")
	}

	indices := make([]int, len(meta.DataResources[0].Variables))
	for i := 0; i < len(indices); i++ {
		indices[i] = i
	}
	err = meta.LoadImportance(config.getAbsolutePath(config.RankingOutputPathRelative), indices)
	if err != nil {
		errors.Wrap(err, "unable to load importance from file")
	}

	// load summary
	err = meta.LoadSummary(config.getAbsolutePath(config.SummaryOutputPathRelative), true)
	if err != nil {
		errors.Wrap(err, "unable to load summary")
	}

	// load stats
	err = meta.LoadDatasetStats(config.getAbsolutePath(config.MergedOutputPathRelative))
	if err != nil {
		errors.Wrap(err, "unable to load stats")
	}

	// create elasticsearch client
	elasticClient, err := elastic.NewClient(
		elastic.SetURL(config.ESEndpoint),
		elastic.SetHttpClient(&http.Client{Timeout: time.Second * time.Duration(config.ESTimeout)}),
		elastic.SetMaxRetries(10),
		elastic.SetSniff(false),
		elastic.SetGzip(true))
	if err != nil {
		errors.Wrap(err, "unable to initialize elastic client")
	}

	// ingest the metadata
	// Create the metadata index if it doesn't exist
	err = metadata.CreateMetadataIndex(elasticClient, config.ESMetadataIndexName, false)
	if err != nil {
		errors.Wrap(err, "unable to create metadata index")
	}

	// Ingest the dataset info into the metadata index
	err = metadata.IngestMetadata(elasticClient, config.ESMetadataIndexName, config.ESDatasetPrefix, meta)
	if err != nil {
		errors.Wrap(err, "unable to ingest metadata")
	}

	// Connect to the database.
	postgresConfig := &conf.Conf{
		DBPassword: config.DatabasePassword,
		DBUser:     config.DatabaseUser,
		Database:   config.Database,
	}
	pg, err := postgres.NewDatabase(postgresConfig)
	if err != nil {
		return errors.Wrap(err, "unable to initialize a new database")
	}

	// Drop the current table if requested.
	pg.DropTable(config.DatabaseTable)

	// Create the database table.
	ds, err := pg.InitializeDataset(meta)
	if err != nil {
		return err
	}

	err = pg.InitializeTable(config.DatabaseTable, ds)
	if err != nil {
		return err
	}

	err = pg.StoreMetadata(config.DatabaseTable)
	if err != nil {
		return err
	}

	err = pg.CreateResultTable(config.DatabaseTable)
	if err != nil {
		return err
	}

	err = pg.CreatePipelineMetadataTables()
	if err != nil {
		return err
	}

	// Load the data.
	reader, err := os.Open(config.getAbsolutePath(config.MergedOutputPathRelative))
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		err = pg.IngestRow(config.DatabaseTable, line)
		if err != nil {
			return errors.Wrap(err, "unable to ingest row")
		}
	}

	err = pg.InsertRemainingRows()
	if err != nil {
		errors.Wrap(err, "unable to ingest last rows")
	}

	return nil
}
