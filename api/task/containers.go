package task

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"

	"github.com/unchartedsoftware/distil-ingest/feature"
	"github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil-ingest/rest"
)

// FeaturizeContainer uses containers to obtain a featurized view of complex variables.
func ClusterContainer(index string, dataset string, config *IngestTaskConfig) error {
	client := rest.NewClient(config.ClusteringRESTEndpoint)

	// create required folders for outputPath
	createContainingDirs(config.getTmpAbsolutePath(config.ClusteringOutputDataRelative))
	createContainingDirs(config.getTmpAbsolutePath(config.ClusteringOutputSchemaRelative))

	// create featurizer
	featurizer := rest.NewFeaturizer(config.ClusteringFunctionName, client)

	// load metadata from original schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(config.getAbsolutePath(config.SchemaPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load original schema file")
	}

	// cluster data
	err = feature.ClusterDataset(meta, featurizer, config.ContainerDataPath,
		config.MediaPath, config.TmpDataPath,
		config.ClusteringOutputDataRelative, config.ClusteringOutputSchemaRelative, config.HasHeader)
	if err != nil {
		return errors.Wrap(err, "unable to cluster data")
	}

	log.Infof("Clustered data written to %s", config.getAbsolutePath(config.TmpDataPath))

	return nil
}

// FeaturizeContainer uses containers to obtain a featurized view of complex variables.
func FeaturizeContainer(index string, dataset string, config *IngestTaskConfig) error {
	client := rest.NewClient(config.FeaturizationRESTEndpoint)

	// create required folders for outputPath
	createContainingDirs(config.getTmpAbsolutePath(config.FeaturizationOutputDataRelative))
	createContainingDirs(config.getTmpAbsolutePath(config.FeaturizationOutputSchemaRelative))

	// create featurizer
	featurizer := rest.NewFeaturizer(config.FeaturizationFunctionName, client)

	// load metadata from cluster schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(config.getTmpAbsolutePath(config.ClusteringOutputSchemaRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load cluster schema file")
	}

	// featurize data
	err = feature.FeaturizeDataset(meta, featurizer, config.ContainerDataPath,
		config.MediaPath, config.TmpDataPath,
		config.FeaturizationOutputDataRelative, config.FeaturizationOutputSchemaRelative, config.HasHeader)
	if err != nil {
		return errors.Wrap(err, "unable to featurize data")
	}

	log.Infof("Featurized data written to %s", config.getAbsolutePath(config.TmpDataPath))

	return nil
}

// ClassifyContainer uses the merged datafile and determines the data types of
// every variable specified in the merged schema file.
func ClassifyContainer(index string, dataset string, config *IngestTaskConfig) error {
	var classification *rest.ClassificationResult
	var err error
	if config.ClassificationEnabled {
		client := rest.NewClient(config.ClassificationRESTEndpoint)

		// create classifier
		classifier := rest.NewClassifier(config.ClassificationFunctionName, client)

		// classify the file
		classification, err = classifier.ClassifyFile(config.getTmpAbsolutePath(config.MergedOutputPathRelative))
		if err != nil {
			return errors.Wrap(err, "unable to classify dataset")
		}

	} else {
		log.Infof("classification disabled, writing out schema types")
		meta, err := metadata.LoadMetadataFromMergedSchema(config.getTmpAbsolutePath(config.MergedOutputSchemaPathRelative))
		if err != nil {
			return errors.Wrap(err, "unable to load metadata")
		}
		if len(meta.DataResources) != 1 {
			return errors.Wrap(err, "loaded metadata not a merged schema")
		}
		classification = &rest.ClassificationResult{
			Labels:        make([][]string, len(meta.DataResources[0].Variables)),
			Probabilities: make([][]float64, len(meta.DataResources[0].Variables)),
			Path:          config.getTmpAbsolutePath(config.MergedOutputPathRelative),
		}
		for i, v := range meta.DataResources[0].Variables {
			classification.Labels[i] = []string{v.Type}
			classification.Probabilities[i] = []float64{float64(1)}
		}
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

// RankContainer the importance of the variables in the dataset.
func RankContainer(index string, dataset string, config *IngestTaskConfig) error {
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

// SummarizeContainer the contents of the dataset.
func SummarizeContainer(index string, dataset string, config *IngestTaskConfig) error {
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
