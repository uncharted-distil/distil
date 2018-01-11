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
	ESDatasetPrefix                  string
}

func (c *ImportTaskConfig) getRootPath(dataset string) string {
	return fmt.Sprintf("%s/%s/%s%s", c.ContainerDataPath, dataset, dataset, c.DatasetFolderSuffix)
}

func (c *ImportTaskConfig) getAbsolutePath(dataset string, relativePath string) string {
	return fmt.Sprintf("%s/%s", c.getRootPath(dataset), relativePath)
}

func (c *ImportTaskConfig) getRawDataPath(dataset string) string {
	return fmt.Sprintf("%s/", c.getRootPath(dataset))
}

func Merge(index string, dataset string, config *ImportTaskConfig) error {
	// load the metadata from schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(config.getAbsolutePath(dataset, config.SchemaPathRelative))
	if err != nil {
		errors.Wrap(err, "unable to load metadata schema")
	}

	// merge file links in dataset
	mergedDR, output, err := merge.InjectFileLinksFromFile(meta, config.getAbsolutePath(dataset, config.DataPathRelative), config.getRawDataPath(dataset), config.HasHeader)
	if err != nil {
		errors.Wrap(err, "unable to merge linked files")
	}

	// write copy to disk
	err = ioutil.WriteFile(config.getAbsolutePath(dataset, config.MergedOutputPathRelative), output, 0644)
	if err != nil {
		errors.Wrap(err, "unable to write merged data")
	}

	// write merged metadata out to disk
	err = meta.WriteMergedSchema(config.getAbsolutePath(dataset, config.MergedOutputSchemaPathRelative), mergedDR)
	if err != nil {
		errors.Wrap(err, "unable to write merged schema")
	}

	return nil
}

func Classify(index string, dataset string, config *ImportTaskConfig) error {
	client := rest.NewClient(config.RESTBaseEndpoint)

	// create classifier
	classifier := rest.NewClassifier(config.ClassificationFunctionName, client)

	// classify the file
	classification, err := classifier.ClassifyFile(config.getAbsolutePath(dataset, config.MergedOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to classify dataset")
	}

	// marshall result
	bytes, err := json.MarshalIndent(classification, "", "    ")
	if err != nil {
		return errors.Wrap(err, "unable to serialize classification result")
	}
	// write to file
	err = ioutil.WriteFile(config.getAbsolutePath(dataset, config.ClassificationOutputPathRelative), bytes, 0644)
	if err != nil {
		errors.Wrap(err, "unable to store classification result")
	}

	return nil
}

func Rank(index string, dataset string, config *ImportTaskConfig) error {
	// create ranker
	client := rest.NewClient(config.RESTBaseEndpoint)
	ranker := rest.NewRanker(config.RankingFunctionName, client)

	// get the importance from the REST interface
	importance, err := ranker.RankFile(config.getAbsolutePath(dataset, config.MergedOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to rank importance file")
	}

	// marshall result
	bytes, err := json.MarshalIndent(importance, "", "    ")
	if err != nil {
		return errors.Wrap(err, "unable to marshall importance ranking result")
	}

	// write to file
	outputPath := config.getAbsolutePath(dataset, config.RankingOutputPathRelative)
	err = ioutil.WriteFile(outputPath, bytes, 0644)
	if err != nil {
		return errors.Wrapf(err, "unable to write importance ranking to '%s'", outputPath)
	}

	return nil
}

func Ingest(index string, dataset string, config *ImportTaskConfig) error {
	meta, err := metadata.LoadMetadataFromClassification(
		config.getAbsolutePath(dataset, config.MergedOutputSchemaPathRelative),
		config.getAbsolutePath(dataset, config.ClassificationOutputPathRelative))
	if err != nil {
		errors.Wrap(err, "unable to load metadata")
	}

	indices := make([]int, len(meta.DataResources[0].Variables))
	for i := 0; i < len(indices); i++ {
		indices[i] = i
	}
	err = meta.LoadImportance(config.getAbsolutePath(dataset, config.RankingOutputPathRelative), indices)
	if err != nil {
		errors.Wrap(err, "unable to load importance from file")
	}

	// load summary
	err = meta.LoadSummary(config.getAbsolutePath(dataset, config.SummaryOutputPathRelative), true)
	if err != nil {
		errors.Wrap(err, "unable to load summary")
	}

	// load stats
	err = meta.LoadDatasetStats(config.getAbsolutePath(dataset, config.MergedOutputPathRelative))
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
	err = metadata.CreateMetadataIndex(elasticClient, index, false)
	if err != nil {
		errors.Wrap(err, "unable to create metadata index")
	}

	// Ingest the dataset info into the metadata index
	err = metadata.IngestMetadata(elasticClient, index, config.ESDatasetPrefix, meta)
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
		return errors.Wrap(err, "unable to initialize a new dataset")
	}

	err = pg.InitializeTable(config.DatabaseTable, ds)
	if err != nil {
		return errors.Wrap(err, "unable to initialize a table")
	}

	err = pg.StoreMetadata(config.DatabaseTable)
	if err != nil {
		return errors.Wrap(err, "unable to store the metadata")
	}

	err = pg.CreateResultTable(config.DatabaseTable)
	if err != nil {
		return errors.Wrap(err, "unable to create the result table")
	}

	err = pg.CreatePipelineMetadataTables()
	if err != nil {
		return errors.Wrap(err, "unable to create pipeline metadata tables")
	}

	// Load the data.
	reader, err := os.Open(config.getAbsolutePath(dataset, config.MergedOutputPathRelative))
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
