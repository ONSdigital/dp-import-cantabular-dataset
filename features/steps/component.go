package steps

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/service"
	"github.com/ONSdigital/log.go/v2/log"

	cmpntest "github.com/ONSdigital/dp-component-test"
	kafka "github.com/ONSdigital/dp-kafka/v2"

	"github.com/maxcnunes/httpfake"
)

var (
	BuildTime string = "1625046891"
	GitCommit string = "7434fe334d9f51b7239f978094ea29d10ac33b16"
	Version   string = ""
)

type Component struct {
	cmpntest.ErrorFeature
	producer      kafka.IProducer
	consumer      kafka.IConsumerGroup
	errorChan     chan error
	svc           *service.Service
	cfg           *config.Config
	RecipeAPI     *httpfake.HTTPFake
	DatasetAPI    *httpfake.HTTPFake
	ImportAPI     *httpfake.HTTPFake
	CantabularSrv *httpfake.HTTPFake
	wg            *sync.WaitGroup
	signals       chan os.Signal
}

func NewComponent() (*Component, error) {
	c := &Component{
		errorChan:     make(chan error),
		DatasetAPI:    httpfake.New(),
		RecipeAPI:     httpfake.New(),
		ImportAPI:     httpfake.New(),
		CantabularSrv: httpfake.New(),
		wg:            &sync.WaitGroup{},
	}

	return c, nil
}

func (c *Component) initService(ctx context.Context) error {
	// Read config
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	cfg.DatasetAPIURL = c.DatasetAPI.ResolveURL("")
	cfg.RecipeAPIURL = c.RecipeAPI.ResolveURL("")
	cfg.CantabularURL = c.CantabularSrv.ResolveURL("")
	cfg.ImportAPIURL = c.ImportAPI.ResolveURL("")

	// producer for triggering test events
	if c.producer, err = kafka.NewProducer(
		ctx,
		cfg.KafkaAddr,
		cfg.InstanceStartedTopic,
		kafka.CreateProducerChannels(),
		&kafka.ProducerConfig{
			KafkaVersion:    &cfg.KafkaVersion,
			MaxMessageBytes: &cfg.KafkaMaxBytes,
		},
	); err != nil {
		return fmt.Errorf("error creating kafka producer: %w", err)
	}
	c.producer.Channels().LogErrors(ctx, "producer")

	// wait for producer to be ready
	<-c.producer.Channels().Ready
	log.Info(ctx, "producer ready")

	// consumer for receiving events
	cgChannels := kafka.CreateConsumerGroupChannels(1)
	kafkaOffset := kafka.OffsetNewest
	if cfg.KafkaOffsetOldest {
		kafkaOffset = kafka.OffsetOldest
	}

	if c.consumer, err = kafka.NewConsumerGroup(
		ctx,
		cfg.KafkaAddr,
		cfg.CategoryDimensionImportTopic,
		"category-dimension-import-group",
		cgChannels,
		&kafka.ConsumerGroupConfig{
			KafkaVersion: &cfg.KafkaVersion,
			Offset:       &kafkaOffset,
		},
	); err != nil {
		return fmt.Errorf("error creating kafka consumer: %w", err)
	}

	// wait for consumer to be ready
	<-c.consumer.Channels().Ready
	log.Info(ctx, "producer ready")

	// Create service and initialise it
	c.svc = service.New()
	if err = c.svc.Init(ctx, cfg, BuildTime, GitCommit, Version); err != nil {
		return fmt.Errorf("unexpected service Init error in NewComponent: %w", err)
	}

	c.cfg = cfg

	// register interrupt signals
	c.signals = make(chan os.Signal, 1)
	signal.Notify(c.signals, os.Interrupt, syscall.SIGTERM)

	return nil
}

func (c *Component) startService(ctx context.Context) {
	defer c.wg.Done()
	c.svc.Start(context.Background(), c.errorChan)

	// blocks until an os interrupt or a fatal error occurs
	select {
	case err := <-c.errorChan:
		err = fmt.Errorf("service error received: %w", err)
		c.svc.Close(ctx)
		panic(fmt.Errorf("unexpected error received from errorChan: %w", err))
	case sig := <-c.signals:
		log.Info(ctx, "os signal received", log.Data{"signal": sig})
	}
	if err := c.svc.Close(ctx); err != nil {
		panic(fmt.Errorf("unexpected error during service graceful shutdown: %w", err))
	}
}

// drainTopic drains the topic of any residual messages between scenarios.
// Prevents future tests failing if previous tests fail unexpectedly and
// leave messages in the queue.
func (c *Component) drainTopic(ctx context.Context) error {
	var msgs []interface{}

	defer func() {
		log.Info(ctx, "drained topic", log.Data{
			"len":      len(msgs),
			"messages": msgs,
		})
	}()

	for {
		select {
		case <-time.After(time.Second * 1):
			return nil
		case msg, ok := <-c.consumer.Channels().Upstream:
			if !ok {
				return errors.New("upstream channel closed")
			}

			msgs = append(msgs, msg)
			msg.Commit()
			msg.Release()
		}
	}
}

// Close runs after each scenario
func (c *Component) Close() {
	ctx := context.Background()

	if err := c.drainTopic(ctx); err != nil {
		log.Error(ctx, "error draining topic", err)
	}

	// close producer
	if err := c.producer.Close(ctx); err != nil {
		log.Error(ctx, "error closing kafka producer", err)
	}

	// close consumer
	if err := c.consumer.Close(ctx); err != nil {
		log.Error(ctx, "error closing kafka consumer", err)
	}

	// kill application
	c.signals <- os.Interrupt

	// wait for graceful shutdown to finish (or timeout)
	c.wg.Wait()
}

// Reset runs before each scenario
func (c *Component) Reset() error {
	ctx := context.Background()

	if err := c.initService(ctx); err != nil {
		return fmt.Errorf("failed to initialise service: %w", err)
	}

	c.DatasetAPI.Reset()
	c.RecipeAPI.Reset()
	c.ImportAPI.Reset()
	c.CantabularSrv.Reset()

	// run application in separate goroutine
	c.wg.Add(1)
	go c.startService(ctx)

	return nil
}
