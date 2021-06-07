package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the event handler service
type Service struct {
	cfg         *config.Config
	server      HTTPServer
	healthCheck HealthChecker
	consumer    kafka.IConsumerGroup
}

// GetKafkaConsumer returns a Kafka consumer with the provided config
var GetKafkaConsumer = func(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
	cgChannels := kafka.CreateConsumerGroupChannels(1)
	kafkaOffset := kafka.OffsetNewest
	if cfg.KafkaOffsetOldest {
		kafkaOffset = kafka.OffsetOldest
	}
	return kafka.NewConsumerGroup(
		ctx,
		cfg.KafkaAddr,
		cfg.HelloCalledTopic,
		cfg.HelloCalledGroup,
		cgChannels,
		&kafka.ConsumerGroupConfig{
			KafkaVersion: &cfg.KafkaVersion,
			Offset:       &kafkaOffset,
		},
	)
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

// New creates a new empty service
func New() *Service {
	return &Service{}
}

// Init initialises all the service dependencies, including healthcheck with checkers, api and middleware
func (svc *Service) Init(ctx context.Context, cfg *config.Config, buildTime, gitCommit, version string) (err error) {

	svc.cfg = cfg

	// Get Kafka consumer
	svc.consumer, err = GetKafkaConsumer(ctx, cfg)
	if err != nil {
		log.Event(ctx, "failed to initialise kafka consumer", log.FATAL, log.Error(err))
		return err
	}

	// Get HealthCheck
	svc.healthCheck, err = GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "could not instantiate healthcheck", log.FATAL, log.Error(err))
		return err
	}
	if err := registerCheckers(ctx, svc.healthCheck, svc.consumer); err != nil {
		return errors.Wrap(err, "unable to register checkers")
	}

	// Get HTTP Server with collectionID checkHeader
	r := mux.NewRouter()
	r.StrictSlash(true).Path("/health").HandlerFunc(svc.healthCheck.Handler)
	svc.server = GetHTTPServer(cfg.BindAddr, r)

	return nil
}

// Start starts an initialised service
func (svc *Service) Start(ctx context.Context, svcErrors chan error) {
	log.Event(ctx, "starting service...", log.INFO)

	// Start kafka error logging
	svc.consumer.Channels().LogErrors(ctx, "error received from kafka consumer, topic: "+svc.cfg.HelloCalledTopic)

	// Event Handler for Kafka Consumer
	event.Consume(ctx, svc.consumer, &event.HelloCalledHandler{}, svc.cfg)

	svc.healthCheck.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := svc.server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.cfg.GracefulShutdownTimeout
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.healthCheck != nil {
			svc.healthCheck.Stop()
			log.Event(ctx, "stopped health checker", log.INFO)
		}

		// If kafka consumer exists, stop listening to it.
		// This will automatically stop the event consumer loops and no more messages will be processed.
		// The kafka consumer will be closed after the service shuts down.
		if svc.consumer != nil {
			if err := svc.consumer.StopListeningToConsumer(ctx); err != nil {
				log.Event(ctx, "error stopping kafka consumer listener", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
			log.Event(ctx, "stopped kafka consumer listener", log.INFO)
		}

		// stop any incoming requests before closing any outbound connections
		if svc.server != nil {
			if err := svc.server.Shutdown(ctx); err != nil {
				log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
				hasShutdownError = true
			}
			log.Event(ctx, "stopped http server", log.INFO)
		}

		// If kafka consumer exists, close it.
		if svc.consumer != nil {
			if err := svc.consumer.Close(ctx); err != nil {
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
		err := errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	return nil
}

// registerCheckers adds the checkers for the service clients to the health check object.
func registerCheckers(ctx context.Context,
	hc HealthChecker,
	consumer kafka.IConsumerGroup) (err error) {

	hasErrors := false

	if err := hc.AddCheck("Kafka consumer", consumer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for Kafka", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
