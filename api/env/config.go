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

package env

import (
	"sync"

	"github.com/caarlos0/env"
)

var (
	cfg  *Config
	once sync.Once
)

// Config represents the application configuration state loaded from env vars.
type Config struct {
	AppPort                            string  `env:"PORT" envDefault:"8080"`
	AugmentedSubFolder                 string  `env:"AUGMENTED_SUBFOLDER" envDefault:"augmented"`
	ClassificationOutputPath           string  `env:"CLASSIFICATION_OUTPUT_PATH" envDefault:"classification.json"`
	ClassificationProbabilityThreshold float64 `env:"CLASSIFICATION_PROBABILITY_THRESHOLD" envDefault:"0.8"`
	ClassificationEnabled              bool    `env:"CLASSIFICATION_ENABLED" envDefault:"true"`
	CleanOutputDataRelative            string  `env:"CLEAN_OUTPUT_DATA" envDefault:"clean/tables/learningData.csv"`
	CleanOutputSchemaRelative          string  `env:"CLEAN_OUTPUT_SCHEMA" envDefault:"clean/datasetDoc.json"`
	ClusteringEnabled                  bool    `env:"CLUSTERING_ENABLED" envDefault:"true"`
	ClusteringOutputDataRelative       string  `env:"CLUSTERING_OUTPUT_DATA" envDefault:"clusters/tables/learningData.csv"`
	ClusteringOutputSchemaRelative     string  `env:"CLUSTERING_OUTPUT_SCHEMA" envDefault:"clusters/datasetDoc.json"`
	ClusteringKMeans                   bool    `env:"CLUSTERING_KMEANS" envDefault:"true"`
	D3MInputDir                        string  `env:"D3MINPUTDIR" envDefault:"datasets"`
	D3MOutputDir                       string  `env:"D3MOUTPUTDIR" envDefault:"outputs"`
	DatamartURIISI                     string  `env:"DATAMART_ISI_URL" envDefault:"https://dsbox02.isi.edu:9000"`
	DatamartURINYU                     string  `env:"DATAMART_NYU_URL" envDefault:"https://auctus.vida-nyu.org"`
	DatamartISIEnabled                 bool    `env:"DATAMART_ISI_ENABLED" envDefault:"false"`
	DatamartNYUEnabled                 bool    `env:"DATAMART_NYU_ENABLED" envDefault:"true"`
	DatamartImportFolder               string  `env:"DATAMART_IMPORT_FOLDER" envDefault:"datamart"`
	ElasticDatasetPrefix               string  `env:"ES_DATASET_PREFIX" envDefault:"d_"`
	ElasticEndpoint                    string  `env:"ES_ENDPOINT" envDefault:"http://localhost:9200"`
	ElasticTimeout                     int     `env:"ES_TIMEOUT" envDefault:"300"`
	ESDatasetsIndex                    string  `env:"ES_DATASETS_INDEX" envDefault:"datasets"`
	ESModelsIndex                      string  `env:"ES_DATASETS_INDEX" envDefault:"models"`
	FastDataPercentage                 float64 `env:"FAST_DATA_PERCENTAGE" envDefault:"0.2"`
	FeaturizationEnabled               bool    `env:"FEATURIZATION_ENABLED" envDefault:"false"`
	FormatOutputDataRelative           string  `env:"FORMAT_OUTPUT_DATA" envDefault:"format/tables/learningData.csv"`
	FormatOutputSchemaRelative         string  `env:"FORMAT_OUTPUT_SCHEMA" envDefault:"format/datasetDoc.json"`
	GeocodingOutputDataRelative        string  `env:"GEOCODING_OUTPUT_DATA" envDefault:"geocoded/tables/learningData.csv"`
	GeocodingOutputSchemaRelative      string  `env:"GEOCODING_OUTPUT_SCHEMA" envDefault:"geocoded/datasetDoc.json"`
	GeocodingEnabled                   bool    `env:"GEOCODING_ENABLED" envDefault:"false"`
	ImportErrorThreshold               float64 `env:"IMPORT_ERROR_THRESHOLD" envDefault:"0.1"`
	IngestHardFail                     bool    `env:"INGEST_HARD_FAIL" envDefault:"false"`
	IngestOverwrite                    bool    `env:"INGEST_OVERWRITE" envDefault:"false"`
	IngestSampleRowLimit               int     `env:"INGEST_SAMPLE_ROW_LIMIT" envDefault:"25000"`
	InitialDataset                     string  `env:"INITIAL_DATASET" envDefault:""`
	MaxTrainingRows                    int     `env:"MAX_TRAINING_ROWS" envDefault:"100000"`
	MaxTestRows                        int     `env:"MAX_TEST_ROWS" envDefault:"100000"`
	MergedOutputDataPath               string  `env:"MERGED_OUTPUT_DATA_PATH" envDefault:"merged/tables/learningData.csv"`
	MergedOutputSchemaPath             string  `env:"MERGED_OUTPUT_SCHEMA_PATH" envDefault:"merged/datasetDoc.json"`
	MinTrainingRows                    int     `env:"MIN_TRAINING_ROWS" envDefault:"100"`
	MinTestRows                        int     `env:"MIN_TEST_ROWS" envDefault:"100"`
	PipelineCacheFilename              string  `env:"PIPELINE_CACHE_FILENAME" envDefault:"cache.bin"`
	PipelineQueueSize                  int     `env:"PIPELINE_QUEUE_SIZE" envDefault:"10"`
	PostgresBatchSize                  int     `env:"PG_BATCH_SIZE" envDefault:"1000"`
	PostgresDatabase                   string  `env:"PG_DATABASE" envDefault:"distil"`
	PostgresHost                       string  `env:"PG_HOST" envDefault:"localhost"`
	PostgresLogLevel                   string  `env:"PG_LOG_LEVEL" envDefault:"none"`
	PostgresPassword                   string  `env:"PG_PASSWORD" envDefault:""`
	PostgresPort                       int     `env:"PG_PORT" envDefault:"5432"`
	PostgresRandomSeed                 float64 `env:"PG_RANDOM_SEED" envDefault:"0.2"`
	PostgresUser                       string  `env:"PG_USER" envDefault:"distil"`
	RankingOutputPath                  string  `env:"RANKING_OUTPUT_PATH" envDefault:"importance.json"`
	RankingRowLimit                    int     `env:"RANKING_ROW_LIMIT" envDefault:"1000"`
	SchemaPath                         string  `env:"SCHEMA_PATH" envDefault:"datasetDoc.json"`
	SkipIngest                         bool    `env:"SKIP_INGEST" envDefault:"false"`
	SkipPreprocessing                  bool    `env:"SKIP_PREPROCESSING" envDefault:"false"`
	SolutionComputeEndpoint            string  `env:"SOLUTION_COMPUTE_ENDPOINT" envDefault:"localhost:50051"`
	SolutionComputeMockEndpoint        string  `env:"SOLUTION_COMPUTE_MOCK_ENDPOINT" envDefault:"localhost:50051"`
	SolutionComputePullTimeout         int     `env:"SOLUTION_COMPUTE_PULL_TIMEOUT" envDefault:"60"`
	SolutionComputePullMax             int     `env:"SOLUTION_COMPUTE_PULL_MAX" envDefault:"10"`
	SolutionSearchMaxTime              int     `env:"SOLUTION_SEARCH_MAX_TIME" envDefault:"10"`
	SolutionComputeTrace               bool    `env:"SOLUTION_COMPUTE_TRACE" envDefault:"false"`
	SummaryPath                        string  `env:"SUMMARY_PATH" envDefault:"summary.txt"`
	SummaryMachinePath                 string  `env:"SUMMARY_MACHINE_PATH" envDefault:"summary-machine.json"`
	SummaryEnabled                     bool    `env:"SUMMARY_ENABLED" envDefault:"true"`
	ServiceRetryCount                  int     `env:"SERVICE_RETRY_COUNT" envDefault:"10"`
	UserProblemPath                    string  `env:"USER_PROBLEM_PATH" envDefault:"outputs/problems"`
	VerboseError                       bool    `env:"VERBOSE_ERROR" envDefault:"false"`
}

// LoadConfig loads the config from the environment if necessary and returns a
// copy.
func LoadConfig() (Config, error) {
	var err error
	once.Do(func() {
		cfg = &Config{}
		err = env.Parse(cfg)
		if err != nil {
			cfg = &Config{}
		}
	})
	return *cfg, err
}
