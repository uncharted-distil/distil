package env

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/caarlos0/env"
)

const (
	tempStorageRoot = "temp_storage_root"
	executablesRoot = "executables_root"
)

var (
	cfg *Config
)

// Config represents the application configuration state loaded from env vars.
type Config struct {
	AppPort                    string `env:"PORT" envDefault:"8080"`
	ElasticEndpoint            string `env:"ES_ENDPOINT" envDefault:"http://localhost:9200"`
	RedisEndpoint              string `env:"REDIS_ENDPOINT" envDefault:"localhost:6379"`
	RedisExpiry                int    `env:"REDIS_EXPIRY" envDefault:"-1"`
	PipelineComputeEndpoint    string `env:"PIPELINE_COMPUTE_ENDPOINT" envDefault:"localhost:50051"`
	PipelineDataDir            string `env:"PIPELINE_DATA_DIR" envDefault:"datasets"`
	PipelineComputeTrace       bool   `env:"PIPELINE_COMPUTE_TRACE" envDefault:"false"`
	ExportPath                 string `env:"EXPORT_PATH"`
	StartupConfigFile          string `env:"JSON_CONFIG_PATH" envDefault:"/d3m/config"`
	PostgresHost               string `env:"PG_HOST" envDefault:"localhost"`
	PostgresPort               string `env:"PG_PORT" envDefault:"5432"`
	PostgresUser               string `env:"PG_USER" envDefault:"distil"`
	PostgresPassword           string `env:"PG_PASSWORD" envDefault:""`
	PostgresDatabase           string `env:"PG_DATABASE" envDefault:"distil"`
	PostgresRetryCount         int    `env:"PG_RETRY_COUNT" envDefault:"100"`
	PostgresRetryTimeout       int    `env:"PG_RETRY_TIMEOUT" envDefault:"4000"`
	PostgresLogLevel           string `env:"PG_LOG_LEVEL" envDefault:"none"`
	PrimitiveEndPoint          string `env:"PRIMITIVE_END_POINT" envDefault:"http://localhost:5000"`
	DataFolderPath             string `env:"DATA_FOLDER_PATH" envDefault:"/d3m/data"`
	DataFilePath               string `env:"DATA_FILE_PATH" envDefault:"/tables/learningData.csv"`
	DatasetFolderSuffix        string `env:"DATASET_FOLDER_SUFFIX" envDefault:"_dataset"`
	MergedOutputDataPath       string `env:"MERGED_OUTPUT_DATA_PATH" envDefault:"tables/merged.csv"`
	MergedOutputSchemaPath     string `env:"MERGED_OUTPUT_SCHEMA_PATH" envDefault:"tables/mergedDataSchema.json"`
	SchemaPath                 string `env:"SCHEMA_PATH" envDefault:"datasetDoc.json"`
	ClassificationEndpoint     string `env:"CLASSIFICATION_ENDPOINT" envDefault:"http://localhost:5000"`
	ClassificationFunctionName string `env:"CLASSIFICATION_FUNCTION_NAME" envDefault:"fileUpload"`
	ClassificationOutputPath   string `env:"CLASSIFICATION_OUTPUT_PATH" envDefault:"tables/classification.json"`
	RankingEndpoint            string `env:"RANKING_ENDPOINT" envDefault:"http://localhost:5001"`
	RankingFunctionName        string `env:"RANKING_FUNCTION_NAME" envDefault:"pca"`
	RankingOutputPath          string `env:"RANKING_OUTPUT_PATH" envDefault:"tables/importance.json"`
	RankingRowLimit            int    `env:"RANKING_ROW_LIMIT" envDefault:"1000"`
	SummaryPath                string `env:"SUMMARY_PATH" envDefault:"summary.txt"`
	ElasticTimeout             int    `env:"ES_TIMEOUT" envDefault:"300"`
	ElasticDatasetPrefix       string `env:"ES_DATASET_PREFIX" envDefault:"d_"`
	InitialDataset             string `env:"INITIAL_DATASET" envDefault:""`
	ESDatasetsIndex            string `env:"ES_DATASETS_INDEX" envDefault:"datasets"`
}

// LoadConfig loads the config from the environment if necessary and returns a
// copy.
func LoadConfig() (Config, error) {
	if cfg == nil {
		cfg = &Config{}
		err := env.Parse(cfg)
		if err != nil {
			return Config{}, err
		}
		// load any overrides from startup file
		err = overideFromStartupFile(cfg)
		if err != nil {
			return Config{}, err
		}
	}
	return *cfg, nil
}

func overideFromStartupFile(cfg *Config) error {
	// Override env/default value with the command line value if set.
	if len(os.Args) > 1 {
		cfg.StartupConfigFile = os.Args[1]
	}

	// read startup config JSON file
	startupConfig, err := ioutil.ReadFile(cfg.StartupConfigFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Failed to read startup config file (%s): %v", cfg.StartupConfigFile, err)
		}
		return nil
	}
	// parse out the entries
	var startupData map[string]interface{}
	err = json.Unmarshal(startupConfig, &startupData)
	if err != nil {
		return fmt.Errorf("Failed to parse startup config file (%s): %v", cfg.StartupConfigFile, err)
	}

	// override / add values
	cfg.PipelineDataDir = startupData[tempStorageRoot].(string)
	cfg.ExportPath = startupData[executablesRoot].(string)

	return nil
}
