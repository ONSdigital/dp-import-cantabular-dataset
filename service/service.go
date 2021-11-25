package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/handler"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"

	"github.com/gorilla/mux"
)

// Service contains all the configs, server and clients to run the event handler service
type Service struct {
	Cfg              *config.Config
	Server           HTTPServer
	HealthCheck      HealthChecker
	Consumer         kafka.IConsumerGroup
	Producer         kafka.IProducer
	Processor        Processor
	cantabularClient CantabularClient
	datasetAPIClient DatasetAPIClient
	recipeAPIClient  RecipeAPIClient
	importAPIClient  ImportAPIClient
}

// GetKafkaConsumer returns a Kafka consumer with the provided config
var GetKafkaConsumer = func(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
	kafkaOffset := kafka.OffsetNewest
	if cfg.KafkaConfig.OffsetOldest {
		kafkaOffset = kafka.OffsetOldest
	}
	cgConfig := &kafka.ConsumerGroupConfig{
		BrokerAddrs:  cfg.KafkaConfig.Addr,
		Topic:        cfg.KafkaConfig.InstanceStartedTopic,
		GroupName:    cfg.KafkaConfig.InstanceStartedGroup,
		KafkaVersion: &cfg.KafkaConfig.Version,
		Offset:       &kafkaOffset,
	}
	if cfg.KafkaConfig.SecProtocol == config.KafkaTLSProtocolFlag {
		cgConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaConfig.SecCACerts,
			cfg.KafkaConfig.SecClientCert,
			cfg.KafkaConfig.SecClientKey,
			cfg.KafkaConfig.SecSkipVerify,
		)
	}
	return kafka.NewConsumerGroup(ctx, cgConfig)
}

// GetKafkaProducer returns a kafka producer with the provided config
var GetKafkaProducer = func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:     cfg.KafkaConfig.Addr,
		Topic:           cfg.KafkaConfig.CategoryDimensionImportTopic,
		KafkaVersion:    &cfg.KafkaConfig.Version,
		MaxMessageBytes: &cfg.KafkaConfig.MaxBytes,
	}
	if cfg.KafkaConfig.SecProtocol == config.KafkaTLSProtocolFlag {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaConfig.SecCACerts,
			cfg.KafkaConfig.SecClientCert,
			cfg.KafkaConfig.SecClientKey,
			cfg.KafkaConfig.SecSkipVerify,
		)
	}
	return kafka.NewProducer(ctx, pConfig)
}

// GetHTTPServer returns an http server
var GetHTTPServer = func(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// GetHealthCheck returns a healthcheck
var GetHealthCheck = func(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

var GetCantabularClient = func(cfg *config.Config) CantabularClient {
	return cantabular.NewClient(
		cantabular.Config{
			Host: cfg.CantabularURL,
		},
		dphttp.NewClient(),
		nil,
	)
}

var GetRecipeAPIClient = func(cfg *config.Config) RecipeAPIClient {
	return recipe.NewClient(cfg.RecipeAPIURL)
}

var GetDatasetAPIClient = func(cfg *config.Config) DatasetAPIClient {
	return dataset.NewAPIClient(cfg.DatasetAPIURL)
}

var GetImportAPIClient = func(cfg *config.Config) ImportAPIClient {
	return importapi.New(cfg.ImportAPIURL)
}

var GetProcessor = func(cfg *config.Config, i ImportAPIClient, d DatasetAPIClient) Processor {
	return event.NewProcessor(*cfg, i, d)
}

// New creates a new empty service
func New() *Service {
	return &Service{}
}

// Init initialises all the service dependencies, including healthcheck with checkers, api and middleware
func (svc *Service) Init(ctx context.Context, cfg *config.Config, buildTime, gitCommit, version string) error {
	var err error

	if cfg == nil {
		return errors.New("nil config passed to service init")
	}

	svc.Cfg = cfg

	// Get Kafka consumer
	if svc.Consumer, err = GetKafkaConsumer(ctx, cfg); err != nil {
		return fmt.Errorf("failed to initialise kafka consumer: %w", err)
	}

	// Get Kafka producer
	svc.Producer, err = GetKafkaProducer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialise kafka producer: %w", err)
	}

	// Get API clients
	svc.cantabularClient = GetCantabularClient(cfg)
	svc.recipeAPIClient = GetRecipeAPIClient(cfg)
	svc.datasetAPIClient = GetDatasetAPIClient(cfg)
	svc.importAPIClient = GetImportAPIClient(cfg)

	// Get processor
	svc.Processor = GetProcessor(cfg, svc.importAPIClient, svc.datasetAPIClient)

	// Get HealthCheck
	if svc.HealthCheck, err = GetHealthCheck(cfg, buildTime, gitCommit, version); err != nil {
		return fmt.Errorf("could not instantiate healthcheck: %w", err)
	}

	if err := svc.registerCheckers(); err != nil {
		return fmt.Errorf("unable to register checkers: %w", err)
	}

	// Get HTTP Server with collectionID checkHeader
	r := mux.NewRouter()
	r.StrictSlash(true).Path("/health").HandlerFunc(svc.HealthCheck.Handler)
	svc.Server = GetHTTPServer(cfg.BindAddr, r)

	return nil
}

// Start starts an initialised service
func (svc *Service) Start(ctx context.Context, svcErrors chan error) {
	log.Event(ctx, "starting service...", log.INFO)

	// Start kafka error logging
	svc.Consumer.LogErrors(ctx)
	svc.Producer.LogErrors(ctx)

	// Start consuming Kafka messages with the Event Handler
	svc.Processor.Consume(
		ctx,
		svc.Consumer,
		handler.NewInstanceStarted(
			*svc.Cfg,
			svc.cantabularClient,
			svc.recipeAPIClient,
			svc.datasetAPIClient,
			svc.Producer,
		),
	)

	// Start the kafka consumer (TODO this will need to be subscribed the the healthcheck)
	svc.Consumer.Start()

	// Start health checker
	svc.HealthCheck.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := svc.Server.ListenAndServe(); err != nil {
			svcErrors <- fmt.Errorf("failure in http listen and serve: %w", err)
		}
	}()
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Cfg.GracefulShutdownTimeout
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.HealthCheck != nil {
			svc.HealthCheck.Stop()
			log.Event(ctx, "stopped health checker", log.INFO)
		}

		// If kafka consumer exists, stop listening to it.
		// This will automatically stop the event consumer loops and no more messages will be processed.
		// The kafka consumer will be closed after the service shuts down.
		if svc.Consumer != nil {
			svc.Consumer.StopAndWait()
			log.Event(ctx, "stopped kafka consumer listener", log.INFO)
		}

		// stop any incoming requests before closing any outbound connections
		if svc.Server != nil {
			if err := svc.Server.Shutdown(ctx); err != nil {
				log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
				hasShutdownError = true
			}
			log.Event(ctx, "stopped http server", log.INFO)
		}

		// If kafka producer exists, close it.
		if svc.Producer != nil {
			if err := svc.Producer.Close(ctx); err != nil {
				log.Event(ctx, "error closing kafka producer", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
			log.Event(ctx, "closed kafka producer", log.INFO)
		}

		// If kafka consumer exists, close it.
		if svc.Consumer != nil {
			if err := svc.Consumer.Close(ctx); err != nil {
				log.Event(ctx, "error closing kafka consumer", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
			log.Event(ctx, "closed kafka consumer", log.INFO)
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Event(ctx, "shutdown timed out", log.ERROR, log.Error(ctx.Err()))
		return ctx.Err()
	}

	// other error
	if hasShutdownError {
		err := fmt.Errorf("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	return nil
}

// registerCheckers adds the checkers for the service clients to the health check object.
func (svc *Service) registerCheckers() error {
	hc := svc.HealthCheck

	if err := hc.AddCheck("Kafka consumer", svc.Consumer.Checker); err != nil {
		return fmt.Errorf("error adding check for Kafka consumer: %w", err)
	}

	if err := hc.AddCheck("Kafka producer", svc.Producer.Checker); err != nil {
		return fmt.Errorf("error adding check for Kafka producer: %w", err)
	}

	if err := hc.AddCheck("Recipe API client", svc.recipeAPIClient.Checker); err != nil {
		return fmt.Errorf("error adding check for Recipe API Client: %w", err)
	}

	// TODO - when Cantabular server is deployed to Production, remove this placeholder and the flag,
	// and always use the real Checker instead: svc.cantabularClient.Checker
	cantabularChecker := svc.cantabularClient.Checker
	if !svc.Cfg.CantabularHealthcheckEnabled {
		cantabularChecker = func(ctx context.Context, state *healthcheck.CheckState) error {
			state.Update(healthcheck.StatusOK, "Cantabular healthcheck placeholder", http.StatusOK)
			return nil
		}
	}
	if err := svc.HealthCheck.AddCheck("Cantabular client", cantabularChecker); err != nil {
		return fmt.Errorf("error adding check for Cantabular client: %w", err)
	}

	if err := hc.AddCheck("Dataset API client", svc.datasetAPIClient.Checker); err != nil {
		return fmt.Errorf("error adding check for Dataset API Client: %w", err)
	}

	return nil
}
