package env

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/caarlos0/env"
)

var (
	cfg *Config
)

// Config represents the application configuration state loaded from env vars.
type Config struct {
	AppPort string `env:"PORT" envDefault:"8080"`

	ElasticEndpoint string `env:"ES_ENDPOINT" envDefault:"http://localhost:9200"`

	RedisEndpoint string `env:"REDIS_ENDPOINT" envDefault:"localhost:6379"`
	RedisExpiry   int    `env:"REDIS_EXPIRY" envDefault:"-1"`

	PipelineComputeEndpoint string `env:"PIPELINE_COMPUTE_ENDPOINT " envDefault:"localhost:50051"`
	PipelineDataDir         string `env:"PIPELINE_DATA_DIR" envDefault:"datasets"`
	PipelineComputeTrace    bool   `env:"PIPELINE_COMPUTE_TRACE" envDefault:"false"`
	ExportPath              string `env:"EXPORT_PATH"`

	StartupConfigFile string `env:"CONFIG_JSON_PATH" envDefault:"startup.json"`

	PostgresHost         string `env:"PG_HOST" envDefault:"localhost"`
	PostgresPort         string `env:"PG_PORT" envDefault:"5432"`
	PostgresUser         string `env:"PG_USER" envDefault:"distil"`
	PostgresPassword     string `env:"PG_PASSWORD" envDefault:""`
	PostgresDatabase     string `env:"PG_DATABASE" envDefault:"distil"`
	PostgresRetryCount   int    `env:"PG_RETRY_COUNT" envDefault:"100"`
	PostgresRetryTimeout int    `env:"PG_RETRY_TIMEOUT" envDefault:"4000"`
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
		err = overideFromStartupFile()
		if err != nil {
			return Config{}, err
		}
	}
	return *cfg, nil
}

func overideFromStartupFile() error {
	// read startup config
	startupConfig, err := ioutil.ReadFile(cfg.StartupConfigFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Failed to read startup config file (%s): %v", cfg.StartupConfigFile, err)
		}
		return nil
	}
	var startupData map[string]interface{}
	err = json.Unmarshal(startupConfig, &startupData)
	if err != nil {
		return fmt.Errorf("Failed to parse startup config file (%s): %v", cfg.StartupConfigFile, err)
	}
	// TODO: implement overrides here
	return nil
}
