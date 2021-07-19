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
	kafkaProducer, err := kafka.NewProducer(ctx, []string{"localhost:9092"}, config.InstanceStartedTopic, pChannels, &kafka.ProducerConfig{
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

		s := schema.InstanceStarted

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

	fmt.Println("Please type the recipe id")
	fmt.Printf("$ ")
	scanner.Scan()
	rID := scanner.Text()

	fmt.Println("Please type the instance id")
	fmt.Printf("$ ")
	scanner.Scan()
	iID := scanner.Text()

	fmt.Println("Please type the job id")
	fmt.Printf("$ ")
	scanner.Scan()
	jID := scanner.Text()

	fmt.Println("Please type the Cntabular Type id")
	fmt.Printf("$ ")
	scanner.Scan()
	cType := scanner.Text()

	return &event.InstanceStarted{
		RecipeID:       rID,
		InstanceID:     iID,
		JobID:          jID,
		CantabularType: cType,
	}
}
