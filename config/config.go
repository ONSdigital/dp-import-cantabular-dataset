package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// KafkaTLSProtocolFlag informs service to use TLS protocol for kafka
const KafkaTLSProtocolFlag = "TLS"

// Config represents service configuration for dp-import-cantabular-dataset
type Config struct {
	BindAddr                     string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout      time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval          time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout   time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	DatasetAPIURL                string        `envconfig:"DATASET_API_URL"`
	RecipeAPIURL                 string        `envconfig:"RECIPE_API_URL"`
	ImportAPIURL                 string        `envconfig:"IMPORT_API_URL"`
	CantabularURL                string        `envconfig:"CANTABULAR_URL"`
	CantabularHealthcheckEnabled bool          `envconfig:"CANTABULAR_HEALTHCHECK_ENABLED"`
	ServiceAuthToken             string        `envconfig:"SERVICE_AUTH_TOKEN"         json:"-"`
	ComponentTestUseLogFile      bool          `envconfig:"COMPONENT_TEST_USE_LOG_FILE"`
	KafkaConfig                  KafkaConfig
}

// KafkaConfig contains the config required to connect to Kafka
type KafkaConfig struct {
	Addr                         []string `envconfig:"KAFKA_ADDR"                            json:"-"`
	Version                      string   `envconfig:"KAFKA_VERSION"`
	OffsetOldest                 bool     `envconfig:"KAFKA_OFFSET_OLDEST"`
	NumWorkers                   int      `envconfig:"KAFKA_NUM_WORKERS"`
	MaxBytes                     int      `envconfig:"KAFKA_MAX_BYTES"`
	SecProtocol                  string   `envconfig:"KAFKA_SEC_PROTO"`
	SecCACerts                   string   `envconfig:"KAFKA_SEC_CA_CERTS"`
	SecClientKey                 string   `envconfig:"KAFKA_SEC_CLIENT_KEY"                  json:"-"`
	SecClientCert                string   `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	SecSkipVerify                bool     `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	InstanceStartedGroup         string   `envconfig:"KAFKA_DATASET_INSTANCE_STARTED_GROUP"`
	InstanceStartedTopic         string   `envconfig:"KAFKA_DATASET_INSTANCE_STARTED_TOPIC"`
	CategoryDimensionImportTopic string   `envconfig:"CANTABULAR_DATASET_CATEGORY_DIMENSION_IMPORT"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                     ":26100",
		GracefulShutdownTimeout:      5 * time.Second,
		HealthCheckInterval:          30 * time.Second,
		HealthCheckCriticalTimeout:   90 * time.Second,
		DatasetAPIURL:                "http://localhost:22000",
		RecipeAPIURL:                 "http://localhost:22300",
		ImportAPIURL:                 "http://localhost:21800",
		CantabularURL:                "http://localhost:8491",
		CantabularHealthcheckEnabled: false,
		ServiceAuthToken:             "",
		ComponentTestUseLogFile:      false,
		KafkaConfig: KafkaConfig{
			Addr:                         []string{"localhost:9092"},
			Version:                      "1.0.2",
			OffsetOldest:                 true,
			NumWorkers:                   1,
			MaxBytes:                     2000000,
			SecProtocol:                  "",
			SecCACerts:                   "",
			SecClientKey:                 "",
			SecClientCert:                "",
			SecSkipVerify:                false,
			InstanceStartedGroup:         "dp-import-cantabular-dataset",
			InstanceStartedTopic:         "cantabular-dataset-instance-started",
			CategoryDimensionImportTopic: "cantabular-dataset-category-dimension-import",
		},
	}

	return cfg, envconfig.Process("", cfg)
}
