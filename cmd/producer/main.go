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

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/log"
	"github.com/ONSdigital/go-ns/avro"
)

const serviceName = "dp-import-cantabular-dataset"

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	// Get Config
	config, err := config.Get()
	if err != nil {
		log.Event(ctx, "error getting config", log.FATAL, log.Error(err))
		os.Exit(1)
	}

	// Create Kafka Producer
	pChannels := kafka.CreateProducerChannels()
	kafkaProducer, err := kafka.NewProducer(ctx, config.KafkaAddr, config.InstanceStartedTopic, pChannels, &kafka.ProducerConfig{
		KafkaVersion: &config.KafkaVersion,
	})
	if err != nil {
		log.Event(ctx, "fatal error trying to create kafka producer", log.FATAL, log.Error(err), log.Data{"topic": config.InstanceStartedTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.Channels().LogErrors(ctx, "kafka producer")

	time.Sleep(500 * time.Millisecond)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		e := scanEvent(scanner)
		log.Event(ctx, "sending instance-started event", log.INFO, log.Data{"instanceStartedEvent": e})

		s := avro.Schema{
			Definition: schema.InstanceStarted,
		}
		bytes, err := s.Marshal(e)
		if err != nil {
			log.Event(ctx, "instance-started event error", log.FATAL, log.Error(err))
			os.Exit(1)
		}

		// Send bytes to Output channel, after calling Initialise just in case it is not initialised.
		kafkaProducer.Initialise(ctx)
		kafkaProducer.Channels().Output <- bytes
	}
}

// scanEvent creates a InstanceStarted event according to the user input
func scanEvent(scanner *bufio.Scanner) *event.InstanceStarted {
	fmt.Println("--- [Send Kafka InstanceStarted] ---")

	fmt.Println("Please type the recipient name")
	fmt.Printf("$ ")
	scanner.Scan()
	name := scanner.Text()

	fmt.Println("Please type the datablob name")
	fmt.Printf("$ ")
	scanner.Scan()
	dbName := scanner.Text()

	fmt.Println("Please type the collection id")
	fmt.Printf("$ ")
	scanner.Scan()
	cID := scanner.Text()

	return &event.InstanceStarted{
		RecipientName: name,
		DatablobName: dbName,
		CollectionID: cID,
	}
}
