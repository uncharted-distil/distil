package env

import (
	enjson "encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

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
	AppPort                            string  `env:"PORT" envDefault:"8080"`
	ElasticEndpoint                    string  `env:"ES_ENDPOINT" envDefault:"http://localhost:9200"`
	SolutionComputeEndpoint            string  `env:"SOLUTION_COMPUTE_ENDPOINT" envDefault:"localhost:50051"`
	SolutionComputePullTimeout         int     `env:"SOLUTION_COMPUTE_PULL_TIMEOUT" envDefault:"60"`
	SolutionComputePullMax             int     `env:"SOLUTION_COMPUTE_PULL_MAX" envDefault:"10"`
	SolutionSearchMaxTime              int     `env:"SOLUTION_SEARCH_MAX_TIME" envDefault:"10"`
	D3MInputDir                        string  `env:"D3MINPUTDIR" envDefault:"datasets"`
	SolutionComputeTrace               bool    `env:"SOLUTION_COMPUTE_TRACE" envDefault:"false"`
	D3MOutputDir                       string  `env:"D3MOUTPUTDIR" envDefault:"outputs"`
	StartupConfigFile                  string  `env:"STARTUP_CONFIG_FILE" envDefault:"search_config.json"`
	PostgresHost                       string  `env:"PG_HOST" envDefault:"localhost"`
	PostgresPort                       int     `env:"PG_PORT" envDefault:"5432"`
	PostgresUser                       string  `env:"PG_USER" envDefault:"distil"`
	PostgresPassword                   string  `env:"PG_PASSWORD" envDefault:""`
	PostgresDatabase                   string  `env:"PG_DATABASE" envDefault:"distil"`
	PostgresLogLevel                   string  `env:"PG_LOG_LEVEL" envDefault:"none"`
	TmpDataPath                        string  `env:"TEMP_STORAGE_ROOT" envDefault:"/d3m/data"`
	DataFolderPath                     string  `env:"DATA_FOLDER_PATH" envDefault:"/d3m/data"`
	ClusteringnRESTEndpoint            string  `env:"CLUSTERING_ENDPOINT" envDefault:"http://127.0.0.1:5004"`
	ClusteringFunctionName             string  `env:"CLUSTERING_FUNCTION_NAME" envDefault:"fileupload"`
	ClusteringOutputDataRelative       string  `env:"CLUSTERING_OUTPUT_DATA" envDefault:"clusters/clusters.csv"`
	ClusteringOutputSchemaRelative     string  `env:"CLUSTERING_OUTPUT_SCHEMA" envDefault:"clustersDatasetDoc.json"`
	ClusteringEnabled                  bool    `env:"CLUSTERING_ENABLED" envDefault:"false"`
	FeaturizationRESTEndpoint          string  `env:"FEATURIZATION_ENDPOINT" envDefault:"http://127.0.0.1:5002"`
	FeaturizationFunctionName          string  `env:"FEATURIZATION_FUNCTION_NAME" envDefault:"fileupload"`
	FeaturizationOutputDataRelative    string  `env:"FEATURIZATION_OUTPUT_DATA" envDefault:"features/features.csv"`
	FeaturizationOutputSchemaRelative  string  `env:"FEATURIZATION_OUTPUT_SCHEMA" envDefault:"featuresDatasetDoc.json"`
	MergedOutputDataPath               string  `env:"MERGED_OUTPUT_DATA_PATH" envDefault:"tables/merged.csv"`
	MergedOutputSchemaPath             string  `env:"MERGED_OUTPUT_SCHEMA_PATH" envDefault:"tables/mergedDataSchema.json"`
	SchemaPath                         string  `env:"SCHEMA_PATH" envDefault:"datasetDoc.json"`
	ClassificationEndpoint             string  `env:"CLASSIFICATION_ENDPOINT" envDefault:"http://127.0.0.1:5000"`
	ClassificationWait                 bool    `env:"CLASSIFICATION_WAIT" envDefault:"false"`
	ClassificationFunctionName         string  `env:"CLASSIFICATION_FUNCTION_NAME" envDefault:"fileUpload"`
	ClassificationOutputPath           string  `env:"CLASSIFICATION_OUTPUT_PATH" envDefault:"tables/classification.json"`
	ClassificationProbabilityThreshold float64 `env:"CLASSIFICATION_PROBABILITY_THRESHOLD" envDefault:"0.8"`
	ClassificationEnabled              bool    `env:"CLASSIFICATION_ENABLED" envDefault:"true"`
	RankingEndpoint                    string  `env:"RANKING_ENDPOINT" envDefault:"http://127.0.0.1:5001"`
	RankingWait                        bool    `env:"RANKING_WAIT" envDefault:"false"`
	RankingFunctionName                string  `env:"RANKING_FUNCTION_NAME" envDefault:"pca"`
	RankingOutputPath                  string  `env:"RANKING_OUTPUT_PATH" envDefault:"tables/importance.json"`
	RankingRowLimit                    int     `env:"RANKING_ROW_LIMIT" envDefault:"1000"`
	SummaryPath                        string  `env:"SUMMARY_PATH" envDefault:"summary.txt"`
	SummaryEndpoint                    string  `env:"SUMMARY_ENDPOINT" envDefault:"http://10.108.4.42:5003"`
	SummaryFunctionName                string  `env:"SUMMARY_FUNCTION_NAME" envDefault:"fileUpload"`
	SummaryMachinePath                 string  `env:"SUMMARY_MACHINE_PATH" envDefault:"summary-machine.json"`
	ElasticTimeout                     int     `env:"ES_TIMEOUT" envDefault:"300"`
	ElasticDatasetPrefix               string  `env:"ES_DATASET_PREFIX" envDefault:"d_"`
	InitialDataset                     string  `env:"INITIAL_DATASET" envDefault:""`
	ESDatasetsIndex                    string  `env:"ES_DATASETS_INDEX" envDefault:"datasets"`
	UserProblemPath                    string  `env:"USER_PROBLEM_PATH" envDefault:"/outputs/problems"`
	SkipIngest                         bool    `env:"SKIP_INGEST" envDefault:"false"`
	IngestHardFail                     bool    `env:"INGEST_HARD_FAIL" envDefault:"false"`
	ServiceRetryCount                  int     `env:"SERVICE_RETRY_COUNT" envDefault:"10"`
	VerboseError                       bool    `env:"VERBOSE_ERROR" envDefault:"false"`
	RootResourceDirectory              string  `env:"ROOT_RESOURCE_DIRECTORY" envDefault:"http://localhost:8001"`
	ResourceProxy                      string  `env:"RESOURCE_PROXY" envDefault:"d_22_hy_dataset_TRAIN,d_66_cn_dataset_TRAIN"`
	IngestPrimitive                    bool    `env:"INGEST_PRIMITIVE" envDefault:"false"`
	IsTask1                            bool    `env:"TASK1" envDefault:"false"`
	IsTask2                            bool    `env:"TASK2" envDefault:"false"`
	SkipPreprocessing                  bool    `env:"SKIP_PREPROCESSING" envDefault:"false"`
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

		cfg.IsTask1 = isTask1(cfg.D3MInputDir)
		cfg.IsTask2 = isTask2(cfg.D3MInputDir)
	}
	return *cfg, nil
}

func overideFromStartupFile(cfg *Config) error {
	// Override env/default value with the command line value if set.
	// startup config file is assumed to be in the input directory.
	startConfigFile := path.Join(cfg.D3MInputDir, cfg.StartupConfigFile)

	log.Infof("Loading overrides from config file (%s)", startConfigFile)

	// read startup config JSON file
	startupConfig, err := ioutil.ReadFile(startConfigFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Failed to read startup config file (%s): %v", startConfigFile, err)
		}
		log.Infof("No config file found at (%s)", startConfigFile)
		return nil
	}
	// parse out the entries
	var startupData map[string]interface{}
	err = enjson.Unmarshal(startupConfig, &startupData)
	if err != nil {
		return fmt.Errorf("Failed to parse startup config file (%s): %v", cfg.StartupConfigFile, err)
	}

	// override / add values

	result, ok := json.String(startupData, executablesRoot)
	if ok {
		cfg.D3MOutputDir = result
	}

	result, ok = json.String(startupData, tempStorageRoot)
	if ok {
		cfg.TmpDataPath = result
	} else {
		cfg.TmpDataPath = cfg.D3MOutputDir
	}

	result, ok = json.String(startupData, userProblemsRoot)
	if ok {
		cfg.UserProblemPath = result
	}

	result, ok = json.String(startupData, trainingDataRoot)
	if ok {
		cfg.DataFolderPath = result
		cfg.InitialDataset = result
	} else {
		dataPath := path.Join(cfg.D3MInputDir, "TRAIN", "dataset_TRAIN")
		cfg.DataFolderPath = dataPath
		cfg.InitialDataset = dataPath
	}

	return nil
}

func isTask1(inputPath string) bool {
	// if Task1 (problem discovery), dataset folder will exist, problem folder
	// will not.
	_, datasetErr := os.Stat(path.Join(inputPath, "TRAIN", "dataset_TRAIN"))
	_, problemErr := os.Stat(path.Join(inputPath, "TRAIN", "problem_TRAIN"))
	return datasetErr == nil && problemErr != nil
}

func isTask2(inputPath string) bool {
	// if Task2 dataset folder AND problem folder will exist
	_, datasetErr := os.Stat(path.Join(inputPath, "TRAIN", "dataset_TRAIN"))
	_, problemErr := os.Stat(path.Join(inputPath, "TRAIN", "problem_TRAIN"))
	return datasetErr == nil && problemErr == nil
}
