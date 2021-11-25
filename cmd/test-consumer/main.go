package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/log.go/log"
)

const serviceName = "kafka-example-consumer"

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Event(ctx, "fatal runtime error", log.Error(err), log.FATAL)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to get config: %s", err)
	}

	// run kafka Consumer Group
	consumerGroup, err := runConsumerGroup(ctx, cfg)
	if err != nil {
		return err
	}

	// blocks until an os interrupt or a fatal error occurs
	sig := <-signals
	log.Event(ctx, "os signal received", log.Data{"signal": sig}, log.INFO)
	return closeConsumerGroup(ctx, consumerGroup, cfg.GracefulShutdownTimeout)
}

func runConsumerGroup(ctx context.Context, cfg *config.Config) (*kafka.ConsumerGroup, error) {
	log.Event(ctx, "[KAFKA-TEST] Starting ConsumerGroup (messages sent to stdout)", log.INFO, log.Data{"config": cfg})
	kafka.SetMaxMessageSize(int32(cfg.KafkaConfig.MaxBytes))

	// Create ConsumerGroup with channels and config
	kafkaOffset := kafka.OffsetOldest
	cgConfig := &kafka.ConsumerGroupConfig{
		BrokerAddrs:  cfg.KafkaConfig.Addr,
		Topic:        cfg.KafkaConfig.CategoryDimensionImportTopic,
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
	cg, err := kafka.NewConsumerGroup(ctx, cgConfig)
	if err != nil {
		return nil, err
	}

	// start consuming as soon as possible
	cg.Start()

	// go-routine to log errors from error channel
	cg.LogErrors(ctx)

	// Consumer not initialised at creation time. We need to retry to initialise it.
	if !cg.IsInitialised() {
		log.Event(ctx, "[KAFKA-TEST] Consumer could not be initialised at creation time. Waiting until we can initialise it.", log.WARN)
		waitForInitialised(ctx, cg.Channels())
	}

	// eventLoop
	consumeCount := 0
	go func() {
		for {
			select {

			case consumedMessage, ok := <-cg.Channels().Upstream:
				if !ok {
					break
				}
				// consumer will be nil if the broker could not be contacted, that's why we use the channel directly instead of consumer.Incoming()
				consumeCount++
				logData := log.Data{"consumeCount": consumeCount, "messageOffset": consumedMessage.Offset()}
				log.Event(ctx, "[KAFKA-TEST] Received message", log.INFO, logData)

				consumedData := consumedMessage.GetData()

				var e event.CategoryDimensionImport
				var s = schema.CategoryDimensionImport

				if err := s.Unmarshal(consumedMessage.GetData(), &e); err != nil {
					log.Error(fmt.Errorf("failed to unmarshal event: %s", err))
				}

				logData["event"] = e
				logData["messageString"] = string(consumedData)
				logData["messageRaw"] = consumedData
				logData["messageLen"] = len(consumedData)

				consumedMessage.CommitAndRelease()
				log.Event(ctx, "[KAFKA-TEST] committed and released message", log.INFO, log.Data{"messageOffset": consumedMessage.Offset()})
			}
		}
	}()
	return cg, nil
}

func closeConsumerGroup(ctx context.Context, cg *kafka.ConsumerGroup, gracefulShutdownTimeout time.Duration) error {
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": gracefulShutdownTimeout}, log.INFO)
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)

	// track shutown gracefully closes up
	var hasShutdownError bool

	// background graceful shutdown
	go func() {
		defer cancel()
		log.Event(ctx, "[KAFKA-TEST] Closing kafka consumerGroup", log.INFO)
		if err := cg.Close(ctx); err != nil {
			hasShutdownError = true
		}
		log.Event(ctx, "[KAFKA-TEST] Closed kafka consumerGroup", log.INFO)
	}()

	// wait for timeout or success (via cancel)
	<-ctx.Done()

	if ctx.Err() == context.DeadlineExceeded {
		log.Event(ctx, "[KAFKA-TEST] graceful shutdown timed out", log.WARN, log.Error(ctx.Err()))
		return ctx.Err()
	}

	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	return nil
}

// waitForInitialised blocks until the consumer is initialised or closed
func waitForInitialised(ctx context.Context, cgChannels *kafka.ConsumerGroupChannels) {
	select {
	case <-cgChannels.Initialised:
		log.Event(ctx, "[KAFKA-TEST] Consumer is now initialised.", log.WARN)
	case <-cgChannels.Closer:
		log.Event(ctx, "[KAFKA-TEST] Consumer is being closed.", log.WARN)
	}
}
