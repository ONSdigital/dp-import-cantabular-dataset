package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/log.go/v2/log"
)

const serviceName = "dp-import-cantabular-dataset"

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(ctx, "fatal runtime error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Get Config
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to get config: %s", err)
	}

	log.Info(ctx, "Starting Kafka Producer (messages read from stdin)", log.Data{"config": cfg})

	// Create Kafka Producer
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:     cfg.KafkaConfig.Addr,
		Topic:           cfg.KafkaConfig.InstanceStartedTopic,
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
	kafkaProducer, err := kafka.NewProducer(ctx, pConfig)
	if err != nil {
		return fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// kafka error logging go-routines
	kafkaProducer.LogErrors(ctx)

	// Wait for producer to be initialised plus 500ms
	<-kafkaProducer.Channels().Initialised
	time.Sleep(500 * time.Millisecond)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		e := scanEvent(scanner)
		log.Info(ctx, "sending instance-started event", log.Data{"instanceStartedEvent": e})

		bytes, err := schema.InstanceStarted.Marshal(e)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		// Send bytes to output channel
		kafkaProducer.Channels().Output <- bytes
	}
}

// scanEvent creates a InstanceStarted event according to the user input
func scanEvent(scanner *bufio.Scanner) *event.InstanceStarted {
	fmt.Println("--- [Send Kafka InstanceStarted] ---")

	e := &event.InstanceStarted{}

	fmt.Println("Please type the recipe id")
	fmt.Printf("$ ")
	scanner.Scan()
	e.RecipeID = scanner.Text()

	fmt.Println("Please type the instance id")
	fmt.Printf("$ ")
	scanner.Scan()
	e.InstanceID = scanner.Text()

	fmt.Println("Please type the job id")
	fmt.Printf("$ ")
	scanner.Scan()
	e.JobID = scanner.Text()

	fmt.Println("Please type the Cantabular Type id")
	fmt.Printf("$ ")
	scanner.Scan()
	e.CantabularType = scanner.Text()

	return e
}
