package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// TODO: remove hello call config options
// Config represents service configuration for dp-import-cantabular-dataset
type Config struct {
	BindAddr                                      string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout                       time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval                           time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout                    time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	KafkaAddr                                     []string      `envconfig:"KAFKA_ADDR"                     json:"-"`
	KafkaVersion                                  string        `envconfig:"KAFKA_VERSION"`
	KafkaOffsetOldest                             bool          `envconfig:"KAFKA_OFFSET_OLDEST"`
	KafkaNumWorkers                               int           `envconfig:"KAFKA_NUM_WORKERS"`
	KafkaMaxBytes                                 int           `envconfig:"KAFKA_MAX_BYTES"`
	HelloCalledGroup                              string        `envconfig:"HELLO_CALLED_GROUP"`
	HelloCalledTopic                              string        `envconfig:"HELLO_CALLED_TOPIC"`
	CantabularDatasetCategoryDimensionImportTopic string        `envconfig:"CANTABULAR_DATASET_CATEGORY_DIMENSION_IMPORT"`
	OutputFilePath                                string        `envconfig:"OUTPUT_FILE_PATH"`
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
		HelloCalledGroup:           "dp-import-cantabular-dataset",
		HelloCalledTopic:           "hello-called",
		CantabularDatasetCategoryDimensionImportTopic: "cantabular-dataset-category-dimension-import",
		OutputFilePath: "/tmp/helloworld.txt",
	}

	return cfg, envconfig.Process("", cfg)
}
