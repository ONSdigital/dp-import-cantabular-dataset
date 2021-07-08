package steps

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"sync"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/service"
	"github.com/ONSdigital/log.go/v2/log"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	cmpntest "github.com/ONSdigital/dp-component-test"

	"github.com/maxcnunes/httpfake"
)

var (
	BuildTime string = "1625046891"
	GitCommit string = "7434fe334d9f51b7239f978094ea29d10ac33b16"
	Version   string = ""
)

type Component struct {
	cmpntest.ErrorFeature
	producer kafka.IProducer
	consumer kafka.IConsumerGroup
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

	// Read config
	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	cfg.DatasetAPIURL = c.DatasetAPI.ResolveURL("") 
	cfg.RecipeAPIURL  = c.RecipeAPI.ResolveURL("")
	cfg.CantabularURL = c.CantabularSrv.ResolveURL("")
	cfg.ImportAPIURL  = c.ImportAPI.ResolveURL("")

	ctx := context.Background()

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
		return nil, fmt.Errorf("error creating kafka producer: %w", err)
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
		return nil, fmt.Errorf("error creating kafka consumer: %w", err)
	}

	// Create service and initialise it
	c.svc = service.New()
	if err = c.svc.Init(ctx, cfg, BuildTime, GitCommit, Version); err != nil {
		return nil, fmt.Errorf("unexpected service Init error in NewComponent: %w", err)
	}

	c.cfg = cfg

	// register interrupt signals
	c.signals = make(chan os.Signal, 1)
	signal.Notify(c.signals, os.Interrupt, syscall.SIGTERM)

	return c, nil
}

func (c *Component) Close() {
	ctx := context.TODO()
	// close producer
	if err := c.producer.Close(ctx); err != nil {
		log.Error(ctx, "error closing a kafka producer", err)
	}

	// kill application
	c.signals <- os.Interrupt

	// wait for graceful shutdown to finish (or timeout)
	c.wg.Wait()
}

func (c *Component) Reset() {
	c.DatasetAPI.Reset()
	c.RecipeAPI.Reset()
	c.ImportAPI.Reset()
	c.CantabularSrv.Reset()

	// run application in separate goroutine
	c.wg.Add(1)
	go c.startService()
}
