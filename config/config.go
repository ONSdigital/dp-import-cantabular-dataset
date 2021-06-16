package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-import-cantabular-dataset
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	KafkaAddr                  []string      `envconfig:"KAFKA_ADDR"                     json:"-"`
	KafkaVersion               string        `envconfig:"KAFKA_VERSION"`
	KafkaOffsetOldest          bool          `envconfig:"KAFKA_OFFSET_OLDEST"`
	KafkaNumWorkers            int           `envconfig:"KAFKA_NUM_WORKERS"`
	InstanceStartedGroup       string        `envconfig:"KAFKA_DATASET_INSTANCE_STARTED_GROUP"`
	InstanceStartedTopic       string        `envconfig:"KAFKA_DATASET_INSTANCE_STARTED_TOPIC"`
	CantabularDatasetCategoryDimensionImportTopic string        `envconfig:"CANTABULAR_DATASET_CATEGORY_DIMENSION_IMPORT"`
	OutputFilePath             string        `envconfig:"OUTPUT_FILE_PATH"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	CodelistAPIURL             string        `envconfig:"CODELIST_API_URL"`
	RecipeAPIURL               string       ` envconfig:"RECIPE_API_URL"`
	CantabularURL              string        `envconfig:"CANTABULAR_URL"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"         json:"-"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   "localhost:26100",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		KafkaAddr:                  []string{"localhost:9092"},
		KafkaVersion:               "1.0.2",
		KafkaOffsetOldest:          true,
		KafkaNumWorkers:            1,
		KafkaMaxBytes:              2000000,
		InstanceStartedGroup:       "dp-import-cantabular-dataset",
		InstanceStartedTopic:       "cantabular-dataset-instance-started",
		OutputFilePath:             "/tmp/helloworld.txt",
		DatasetAPIURL:              "http://localhost:22000",
		CodelistAPIURL:             "http://localhost:22400",
		RecipeAPIURL:               "http://localhost:22300",
		CantabularURL:              "http://localhost:8491",
		ServiceAuthToken:           "",
		CantabularDatasetCategoryDimensionImportTopic: "cantabular-dataset-category-dimension-import",
	}

	return cfg, envconfig.Process("", cfg)
}
