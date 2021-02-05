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
	BatchSubFolder                     string  `env:"BATCH_SUBFOLDER" envDefault:"batch"`
	ClassificationOutputPath           string  `env:"CLASSIFICATION_OUTPUT_PATH" envDefault:"classification.json"`
	ClassificationProbabilityThreshold float64 `env:"CLASSIFICATION_PROBABILITY_THRESHOLD" envDefault:"0.8"`
	ClassificationEnabled              bool    `env:"CLASSIFICATION_ENABLED" envDefault:"true"`
	ClusteringEnabled                  bool    `env:"CLUSTERING_ENABLED" envDefault:"true"`
	ClusteringKMeans                   bool    `env:"CLUSTERING_KMEANS" envDefault:"true"`
	D3MInputDir                        string  `env:"D3MINPUTDIR" envDefault:"datasets"`
	D3MOutputDir                       string  `env:"D3MOUTPUTDIR" envDefault:"outputs"`
	DatamartURIISI                     string  `env:"DATAMART_ISI_URL" envDefault:"https://dsbox02.isi.edu:9000"`
	DatamartURINYU                     string  `env:"DATAMART_NYU_URL" envDefault:"https://auctus.vida-nyu.org"`
	DatamartISIEnabled                 bool    `env:"DATAMART_ISI_ENABLED" envDefault:"false"`
	DatamartNYUEnabled                 bool    `env:"DATAMART_NYU_ENABLED" envDefault:"true"`
	DatamartImportFolder               string  `env:"DATAMART_IMPORT_FOLDER" envDefault:"datamart"`
	DatasetBatchSize                   int     `env:"DATASET_BATCH_SIZE" envDefault:"10000"`
	ElasticDatasetPrefix               string  `env:"ES_DATASET_PREFIX" envDefault:"d_"`
	ElasticEndpoint                    string  `env:"ES_ENDPOINT" envDefault:"http://localhost:9200"`
	ElasticTimeout                     int     `env:"ES_TIMEOUT" envDefault:"300"`
	ESDatasetsIndex                    string  `env:"ES_DATASETS_INDEX" envDefault:"datasets"`
	ESModelsIndex                      string  `env:"ES_DATASETS_INDEX" envDefault:"models"`
	FastDataPercentage                 float64 `env:"FAST_DATA_PERCENTAGE" envDefault:"0.2"`
	FeaturizationEnabled               bool    `env:"FEATURIZATION_ENABLED" envDefault:"false"`
	GeocodingEnabled                   bool    `env:"GEOCODING_ENABLED" envDefault:"false"`
	HelpURL                            string  `env:"HELP_URL" envDefault:"https://d3m.uncharted.software/"`
	ImportErrorThreshold               float64 `env:"IMPORT_ERROR_THRESHOLD" envDefault:"0.1"`
	IngestHardFail                     bool    `env:"INGEST_HARD_FAIL" envDefault:"false"`
	IngestOverwrite                    bool    `env:"INGEST_OVERWRITE" envDefault:"false"`
	IngestSampleRowLimit               int     `env:"INGEST_SAMPLE_ROW_LIMIT" envDefault:"25000"`
	InitialDataset                     string  `env:"INITIAL_DATASET" envDefault:""`
	MaxTrainingRows                    int     `env:"MAX_TRAINING_ROWS" envDefault:"100000"`
	MaxTestRows                        int     `env:"MAX_TEST_ROWS" envDefault:"100000"`
	MinTrainingRows                    int     `env:"MIN_TRAINING_ROWS" envDefault:"100"`
	MinTestRows                        int     `env:"MIN_TEST_ROWS" envDefault:"100"`
	PipelineCacheEnabled               bool    `env:"PIPELINE_CACHE_ENABLED" envDefault:"true"`
	PipelineCacheFilename              string  `env:"PIPELINE_CACHE_FILENAME" envDefault:"cache.bin"`
	PipelineQueueSize                  int     `env:"PIPELINE_QUEUE_SIZE" envDefault:"10"`
	PoolFeatures                       bool    `env:"POOL_FEATURES" envDefault:"true"`
	PostgresBatchSize                  int     `env:"PG_BATCH_SIZE" envDefault:"1000"`
	PostgresDatabase                   string  `env:"PG_DATABASE" envDefault:"distil"`
	PostgresHost                       string  `env:"PG_HOST" envDefault:"localhost"`
	PostgresLogLevel                   string  `env:"PG_LOG_LEVEL" envDefault:"none"`
	PostgresPassword                   string  `env:"PG_PASSWORD" envDefault:""`
	PostgresPort                       int     `env:"PG_PORT" envDefault:"5432"`
	PostgresRandomSeed                 float64 `env:"PG_RANDOM_SEED" envDefault:"0.2"`
	PostgresUser                       string  `env:"PG_USER" envDefault:"distil"`
	PublicSubFolder                    string  `env:"PUBLIC_SUBFOLDER" envDefault:"public"`
	RankingOutputPath                  string  `env:"RANKING_OUTPUT_PATH" envDefault:"importance.json"`
	RankingRowLimit                    int     `env:"RANKING_ROW_LIMIT" envDefault:"1000"`
	RemoteSensingGPUBatchSize          int     `env:"REMOTE_SENSING_GPU_BATCH_SIZE" envDefault:"32"`
	RemoteSensingNumJobs               int     `env:"REMOTE_SENSING_NUM_JOBS" envDefault:"-1"` // -1 sets num jobs = num cpus
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
	TrainTestSplit                     float64 `env:"TRAIN_TEST_SPLIT" envDefault:"0.9"`
	TrainTestSplitTimeSeries           float64 `env:"TRAIN_TEST_SPLIT_TIMESERIES" envDefault:"0.9"`
	UserProblemPath                    string  `env:"USER_PROBLEM_PATH" envDefault:"outputs/problems"`
	VerboseError                       bool    `env:"VERBOSE_ERROR" envDefault:"false"`
	ShouldScaleImages				   bool    `env:"SHOULD_SCALE_IMAGES" envDefault:"false"` // enables and disables image scaling
}

// LoadConfig loads the config from the environment if necessary and returns a copy.
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
