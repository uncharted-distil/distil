package task

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v5"

	"github.com/unchartedsoftware/distil-ingest/conf"
	"github.com/unchartedsoftware/distil-ingest/merge"
	"github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil-ingest/postgres"
	"github.com/unchartedsoftware/distil-ingest/rest"
)

const (
	rankingFilename = "rank-no-missing.csv"
	baseTableSuffix = "_base"
)

// IngestTaskConfig captures the necessary configuration for an data ingest.
type IngestTaskConfig struct {
	ContainerDataPath                string
	TmpDataPath                      string
	DataPathRelative                 string
	DatasetFolderSuffix              string
	HasHeader                        bool
	MergedOutputPathRelative         string
	MergedOutputSchemaPathRelative   string
	SchemaPathRelative               string
	ClassificationRESTEndpoint       string
	ClassificationFunctionName       string
	ClassificationOutputPathRelative string
	RankingRESTEndpoint              string
	RankingFunctionName              string
	RankingOutputPathRelative        string
	RankingRowLimit                  int
	DatabasePassword                 string
	DatabaseUser                     string
	Database                         string
	SummaryOutputPathRelative        string
	SummaryMachineOutputPathRelative string
	SummaryRESTEndpoint              string
	SummaryFunctionName              string
	ESEndpoint                       string
	ESTimeout                        int
	ESDatasetPrefix                  string
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
func IngestDataset(index string, dataset string, config *IngestTaskConfig) error {
	// Make sure the temp data directory exists.
	tmpPath := path.Dir(config.getTmpAbsolutePath(config.MergedOutputSchemaPathRelative))
	os.MkdirAll(tmpPath, os.ModePerm)

	err := Merge(index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to merge all data into a single file")
	}

	err = Classify(index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to classify fields")
	}

	err = Rank(index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to rank field importance")
	}

	err = Summarize(index, dataset, config)
	// NOTE: For now ignore summary errors!
	//if err != nil {
	//	return errors.Wrap(err, "unable to summarize the dataset")
	//}

	err = Ingest(index, dataset, config)
	if err != nil {
		return errors.Wrap(err, "unable to ingest ranked data")
	}

	return nil
}

// Merge combines all the source data files into a single datafile.
func Merge(index string, dataset string, config *IngestTaskConfig) error {
	// load the metadata from schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(config.getAbsolutePath(config.SchemaPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load metadata schema")
	}

	// merge file links in dataset
	mergedDR, output, err := merge.InjectFileLinksFromFile(meta, config.getAbsolutePath(config.DataPathRelative), config.getRawDataPath(), config.HasHeader)
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

// Classify uses the merged datafile and determines the data types of
// every variable specified in the merged schema file.
func Classify(index string, dataset string, config *IngestTaskConfig) error {
	client := rest.NewClient(config.ClassificationRESTEndpoint)

	// create classifier
	classifier := rest.NewClassifier(config.ClassificationFunctionName, client)

	// classify the file
	classification, err := classifier.ClassifyFile(config.getTmpAbsolutePath(config.MergedOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to classify dataset")
	}

	// marshall result
	bytes, err := json.MarshalIndent(classification, "", "    ")
	if err != nil {
		return errors.Wrap(err, "unable to serialize classification result")
	}
	// write to file
	err = ioutil.WriteFile(config.getTmpAbsolutePath(config.ClassificationOutputPathRelative), bytes, 0644)
	if err != nil {
		return errors.Wrap(err, "unable to store classification result")
	}

	return nil
}

// Rank the importance of the variables in the dataset.
func Rank(index string, dataset string, config *IngestTaskConfig) error {
	// get the header for the rank data
	meta, err := metadata.LoadMetadataFromClassification(
		config.getTmpAbsolutePath(config.MergedOutputSchemaPathRelative),
		config.getTmpAbsolutePath(config.ClassificationOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load metadata")
	}

	header, err := meta.GenerateHeaders()
	if err != nil {
		return errors.Wrap(err, "unable to load metadata")
	}

	if len(header) != 1 {
		return errors.Errorf("merge data should only have one header but found %d", len(header))
	}

	// need to ignore rows with missing
	// ranking requires a header
	err = removeMissingValues(
		config.getTmpAbsolutePath(config.MergedOutputPathRelative),
		config.getTmpAbsolutePath(rankingFilename),
		config.HasHeader, header[0], config.RankingRowLimit)
	if err != nil {
		return errors.Wrap(err, "unable to ignore missing values")
	}

	// create ranker
	client := rest.NewClient(config.RankingRESTEndpoint)
	ranker := rest.NewRanker(config.RankingFunctionName, client)

	// get the importance from the REST interface
	importance, err := ranker.RankFile(config.getTmpAbsolutePath(rankingFilename))
	if err != nil {
		return errors.Wrap(err, "unable to rank importance file")
	}

	// marshall result
	bytes, err := json.MarshalIndent(importance, "", "    ")
	if err != nil {
		return errors.Wrap(err, "unable to marshall importance ranking result")
	}

	// write to file
	outputPath := config.getTmpAbsolutePath(config.RankingOutputPathRelative)
	err = ioutil.WriteFile(outputPath, bytes, 0644)
	if err != nil {
		return errors.Wrapf(err, "unable to write importance ranking to '%s'", outputPath)
	}

	return nil
}

// Summarize the contents of the dataset.
func Summarize(index string, dataset string, config *IngestTaskConfig) error {
	// create ranker
	client := rest.NewClient(config.SummaryRESTEndpoint)
	summarizer := rest.NewSummarizer(config.SummaryFunctionName, client)

	// get the importance from the REST interface
	summary, err := summarizer.SummarizeFile(config.getTmpAbsolutePath(config.MergedOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to summarize merged file")
	}

	// marshall result
	bytes, err := json.MarshalIndent(summary, "", "    ")
	if err != nil {
		return errors.Wrap(err, "unable to marshall summary result")
	}

	// write to file
	outputPath := config.getTmpAbsolutePath(config.SummaryMachineOutputPathRelative)
	err = ioutil.WriteFile(outputPath, bytes, 0644)
	if err != nil {
		return errors.Wrapf(err, "unable to write summary to '%s'", outputPath)
	}

	return nil
}

// Ingest the metadata to ES and the data to Postgres.
func Ingest(index string, dataset string, config *IngestTaskConfig) error {
	meta, err := metadata.LoadMetadataFromClassification(
		config.getTmpAbsolutePath(config.MergedOutputSchemaPathRelative),
		config.getTmpAbsolutePath(config.ClassificationOutputPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load metadata")
	}

	// Adjust the ID & name of the dataset as needed
	fixDatasetIDName(meta)

	indices := make([]int, len(meta.DataResources[0].Variables))
	for i := 0; i < len(indices); i++ {
		indices[i] = i
	}
	err = meta.LoadImportance(config.getTmpAbsolutePath(config.RankingOutputPathRelative), indices)
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

	// load stats
	err = meta.LoadSummaryMachine(config.getTmpAbsolutePath(config.SummaryMachineOutputPathRelative))
	// NOTE: For now ignore summary errors!
	//if err != nil {
	//	return errors.Wrap(err, "unable to load stats")
	//}

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

	dbTable := strings.Replace(meta.ID, "_dataset", "", -1)
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

	err = pg.CreatePipelineMetadataTables()
	if err != nil {
		return errors.Wrap(err, "unable to create pipeline metadata tables")
	}

	// Load the data.
	reader, err := os.Open(config.getTmpAbsolutePath(config.MergedOutputPathRelative))
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		err = pg.IngestRow(dbTable, line)
		if err != nil {
			return errors.Wrap(err, "unable to ingest row")
		}
	}

	err = pg.InsertRemainingRows()
	if err != nil {
		return errors.Wrap(err, "unable to ingest last rows")
	}

	return nil
}

func removeMissingValues(sourceFile string, destinationFile string, hasHeader bool, headerToWrite []string, rowLimit int) error {
	// Copy source to destination, removing rows that have missing values.
	file, err := os.Open(sourceFile)
	if err != nil {
		return errors.Wrap(err, "failed to open source file")
	}

	reader := csv.NewReader(file)

	// output writer
	output := &bytes.Buffer{}
	writer := csv.NewWriter(output)
	if headerToWrite != nil && len(headerToWrite) > 0 {
		err := writer.Write(headerToWrite)
		if err != nil {
			return errors.Wrap(err, "failed to write header to file")
		}
	}

	count := 0
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "failed to read line from file")
		}
		if (count > 0 || !hasHeader) && (count < rowLimit) {
			// write the csv line back out
			err := writer.Write(line)
			if err != nil {
				return errors.Wrap(err, "failed to write line to file")
			}
		}
		count++
	}
	// flush writer
	writer.Flush()

	err = ioutil.WriteFile(destinationFile, output.Bytes(), 0644)
	if err != nil {
		return errors.Wrap(err, "failed to close output file")
	}

	// close left
	err = file.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close input file")
	}
	return nil
}

func fixDatasetIDName(meta *metadata.Metadata) {
	// Train dataset ID & name need to be adjusted to fit the expected format.
	// The ID MUST end in _dataset, and the name should be representative.
	if isTrainDataset(meta) {
		meta.ID = strings.TrimSuffix(meta.ID, "_TRAIN")
		if !strings.HasSuffix(meta.ID, "_dataset") {
			meta.ID = fmt.Sprintf("%s%s", meta.ID, "_dataset")
		}
		meta.Name = strings.TrimSuffix(meta.ID, "_dataset")
	}
}

func isTrainDataset(meta *metadata.Metadata) bool {
	return strings.HasSuffix(meta.ID, "_TRAIN")
}
