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
	ElasticEndpoint                    string  `env:"ES_ENDPOINT" envDefault:"http://localhost:9200"`
	UseTA2Runner                       bool    `env:"USE_TA2_RUNNER" envDefault:"false"`
	SolutionComputeEndpoint            string  `env:"SOLUTION_COMPUTE_ENDPOINT" envDefault:"localhost:50051"`
	SolutionComputeMockEndpoint        string  `env:"SOLUTION_COMPUTE_MOCK_ENDPOINT" envDefault:"localhost:50051"`
	SolutionComputePullTimeout         int     `env:"SOLUTION_COMPUTE_PULL_TIMEOUT" envDefault:"60"`
	SolutionComputePullMax             int     `env:"SOLUTION_COMPUTE_PULL_MAX" envDefault:"10"`
	SolutionSearchMaxTime              int     `env:"SOLUTION_SEARCH_MAX_TIME" envDefault:"10"`
	AugmentedSubFolder                 string  `env:"AUGMENTED_SUBFOLDER" envDefault:"augmented"`
	D3MInputDir                        string  `env:"D3MINPUTDIR" envDefault:"datasets"`
	D3MOutputDir                       string  `env:"D3MOUTPUTDIR" envDefault:"outputs"`
	DatamartURIISI                     string  `env:"DATAMART_URL_ISI" envDefault:"http://dsbox02.isi.edu:9000"`
	DatamartURINYU                     string  `env:"DATAMART_URL_NYU" envDefault:"https://auctus.vida-nyu.org"`
	DatamartISIEnabled                 bool    `env:"DATAMART_ISI_ENABLED" envDefault:"false"`
	DatamartNYUEnabled                 bool    `env:"DATAMART_NYU_ENABLED" envDefault:"true"`
	DatamartImportFolder               string  `env:"DATAMART_IMPORT_FOLDER" envDefault:"datamart"`
	SolutionComputeTrace               bool    `env:"SOLUTION_COMPUTE_TRACE" envDefault:"false"`
	PostgresHost                       string  `env:"PG_HOST" envDefault:"localhost"`
	PostgresPort                       int     `env:"PG_PORT" envDefault:"5432"`
	PostgresUser                       string  `env:"PG_USER" envDefault:"distil"`
	PostgresPassword                   string  `env:"PG_PASSWORD" envDefault:""`
	PostgresDatabase                   string  `env:"PG_DATABASE" envDefault:"distil"`
	PostgresLogLevel                   string  `env:"PG_LOG_LEVEL" envDefault:"none"`
	ClusteringOutputDataRelative       string  `env:"CLUSTERING_OUTPUT_DATA" envDefault:"clusters/tables/learningData.csv"`
	ClusteringOutputSchemaRelative     string  `env:"CLUSTERING_OUTPUT_SCHEMA" envDefault:"clusters/datasetDoc.json"`
	ClusteringEnabled                  bool    `env:"CLUSTERING_ENABLED" envDefault:"false"`
	FeaturizationOutputDataRelative    string  `env:"FEATURIZATION_OUTPUT_DATA" envDefault:"features/tables/learningData.csv"`
	FeaturizationOutputSchemaRelative  string  `env:"FEATURIZATION_OUTPUT_SCHEMA" envDefault:"features/datasetDoc.json"`
	FormatOutputDataRelative           string  `env:"FORMAT_OUTPUT_DATA" envDefault:"format/tables/learningData.csv"`
	FormatOutputSchemaRelative         string  `env:"FORMAT_OUTPUT_SCHEMA" envDefault:"format/datasetDoc.json"`
	CleanOutputDataRelative            string  `env:"CLEAN_OUTPUT_DATA" envDefault:"clean/tables/learningData.csv"`
	CleanOutputSchemaRelative          string  `env:"CLEAN_OUTPUT_SCHEMA" envDefault:"clean/datasetDoc.json"`
	GeocodingOutputDataRelative        string  `env:"GEOCODING_OUTPUT_DATA" envDefault:"geocoded/tables/learningData.csv"`
	GeocodingOutputSchemaRelative      string  `env:"GEOCODING_OUTPUT_SCHEMA" envDefault:"geocoded/datasetDoc.json"`
	GeocodingEnabled                   bool    `env:"GEOCODING_ENABLED" envDefault:"false"`
	MergedOutputDataPath               string  `env:"MERGED_OUTPUT_DATA_PATH" envDefault:"merged/tables/learningData.csv"`
	MergedOutputSchemaPath             string  `env:"MERGED_OUTPUT_SCHEMA_PATH" envDefault:"merged/datasetDoc.json"`
	SchemaPath                         string  `env:"SCHEMA_PATH" envDefault:"datasetDoc.json"`
	ClassificationOutputPath           string  `env:"CLASSIFICATION_OUTPUT_PATH" envDefault:"classification.json"`
	ClassificationProbabilityThreshold float64 `env:"CLASSIFICATION_PROBABILITY_THRESHOLD" envDefault:"0.8"`
	ClassificationEnabled              bool    `env:"CLASSIFICATION_ENABLED" envDefault:"true"`
	RankingOutputPath                  string  `env:"RANKING_OUTPUT_PATH" envDefault:"importance.json"`
	RankingRowLimit                    int     `env:"RANKING_ROW_LIMIT" envDefault:"1000"`
	SummaryPath                        string  `env:"SUMMARY_PATH" envDefault:"summary.txt"`
	SummaryMachinePath                 string  `env:"SUMMARY_MACHINE_PATH" envDefault:"summary-machine.json"`
	SummaryEnabled                     bool    `env:"SUMMARY_ENABLED" envDefault:"true"`
	ElasticTimeout                     int     `env:"ES_TIMEOUT" envDefault:"300"`
	ElasticDatasetPrefix               string  `env:"ES_DATASET_PREFIX" envDefault:"d_"`
	InitialDataset                     string  `env:"INITIAL_DATASET" envDefault:""`
	ESDatasetsIndex                    string  `env:"ES_DATASETS_INDEX" envDefault:"datasets"`
	UserProblemPath                    string  `env:"USER_PROBLEM_PATH" envDefault:"outputs/problems"`
	SkipIngest                         bool    `env:"SKIP_INGEST" envDefault:"false"`
	IngestHardFail                     bool    `env:"INGEST_HARD_FAIL" envDefault:"false"`
	ServiceRetryCount                  int     `env:"SERVICE_RETRY_COUNT" envDefault:"10"`
	VerboseError                       bool    `env:"VERBOSE_ERROR" envDefault:"false"`
	IsTask1                            bool    `env:"TASK1" envDefault:"false"`
	IsTask2                            bool    `env:"TASK2" envDefault:"false"`
	SkipPreprocessing                  bool    `env:"SKIP_PREPROCESSING" envDefault:"false"`
	MaxTrainingRows                    int     `env:"MAX_TRAINING_ROWS" envDefault:"10000"`
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
