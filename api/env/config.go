package env

import (
	enjson "encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/caarlos0/env"
	"github.com/unchartedsoftware/distil/api/util/json"
	"github.com/unchartedsoftware/plog"
)

const (
	tempStorageRoot  = "temp_storage_root"
	executablesRoot  = "executables_root"
	userProblemsRoot = "user_problems_root"
	trainingDataRoot = "training_data_root"
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
	TmpDataPath                string `env:"TEMP_STORAGE_ROOT" envDefault:"/d3m/data"`
	DataFolderPath             string `env:"DATA_FOLDER_PATH" envDefault:"/d3m/data"`
	DataFilePath               string `env:"DATA_FILE_PATH" envDefault:"/tables/learningData.csv"`
	DatasetFolderSuffix        string `env:"DATASET_FOLDER_SUFFIX" envDefault:"_dataset"`
	MergedOutputDataPath       string `env:"MERGED_OUTPUT_DATA_PATH" envDefault:"tables/merged.csv"`
	MergedOutputSchemaPath     string `env:"MERGED_OUTPUT_SCHEMA_PATH" envDefault:"tables/mergedDataSchema.json"`
	SchemaPath                 string `env:"SCHEMA_PATH" envDefault:"datasetDoc.json"`
	ClassificationEndpoint     string `env:"CLASSIFICATION_ENDPOINT" envDefault:"http://localhost:5000"`
	ClassificationWait         bool   `env:"CLASSIFICATION_WAIT" envDefault:"false"`
	ClassificationFunctionName string `env:"CLASSIFICATION_FUNCTION_NAME" envDefault:"fileUpload"`
	ClassificationOutputPath   string `env:"CLASSIFICATION_OUTPUT_PATH" envDefault:"tables/classification.json"`
	RankingEndpoint            string `env:"RANKING_ENDPOINT" envDefault:"http://localhost:5001"`
	RankingWait                bool   `env:"RANKING_WAIT" envDefault:"false"`
	RankingFunctionName        string `env:"RANKING_FUNCTION_NAME" envDefault:"pca"`
	RankingOutputPath          string `env:"RANKING_OUTPUT_PATH" envDefault:"tables/importance.json"`
	RankingRowLimit            int    `env:"RANKING_ROW_LIMIT" envDefault:"1000"`
	SummaryPath                string `env:"SUMMARY_PATH" envDefault:"summary.txt"`
	SummaryEndpoint            string `env:"SUMMARY_ENDPOINT" envDefault:"http://10.108.4.42:5001"`
	SummaryFunctionName        string `env:"SUMMARY_FUNCTION_NAME" envDefault:"fileUpload"`
	SummaryMachinePath         string `env:"SUMMARY_MACHINE_PATH" envDefault:"summary-machine.json"`
	ElasticTimeout             int    `env:"ES_TIMEOUT" envDefault:"300"`
	ElasticDatasetPrefix       string `env:"ES_DATASET_PREFIX" envDefault:"d_"`
	InitialDataset             string `env:"INITIAL_DATASET" envDefault:""`
	ESDatasetsIndex            string `env:"ES_DATASETS_INDEX" envDefault:"datasets"`
	UserProblemPath            string `env:"USER_PROBLEM_PATH" envDefault:"datasets"`
	SkipIngest                 bool   `env:"SKIP_INGEST" envDefault:"false"`
	ServiceRetryCount          int    `env:"SERVICE_RETRY_COUNT" envDefault:"10"`
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

	log.Infof("Loading overrides from config file (%s)", cfg.StartupConfigFile)

	// read startup config JSON file
	startupConfig, err := ioutil.ReadFile(cfg.StartupConfigFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Failed to read startup config file (%s): %v", cfg.StartupConfigFile, err)
		}
		log.Infof("No config file found at (%s)", cfg.StartupConfigFile)
		return nil
	}
	// parse out the entries
	var startupData map[string]interface{}
	err = enjson.Unmarshal(startupConfig, &startupData)
	if err != nil {
		return fmt.Errorf("Failed to parse startup config file (%s): %v", cfg.StartupConfigFile, err)
	}

	// override / add values

	result, ok := json.String(startupData, tempStorageRoot)
	if ok {
		cfg.PipelineDataDir = result
		cfg.TmpDataPath = result
	}

	result, ok = json.String(startupData, executablesRoot)
	if ok {
		cfg.ExportPath = result
	}

	result, ok = json.String(startupData, userProblemsRoot)
	if ok {
		cfg.UserProblemPath = result
	}

	result, ok = json.String(startupData, trainingDataRoot)
	if ok {
		cfg.DataFolderPath = result
		cfg.InitialDataset = result
	}
	return nil
}
