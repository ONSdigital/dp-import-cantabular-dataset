package event

import (
	"context"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out mock/handler.go -pkg mock . Handler

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(context.Context, *config.Config, *InstanceStarted) error
}

// Consume converts messages to event instances, and pass the event to the provided handler.
func Consume(ctx context.Context, consumer kafka.IConsumerGroup, h Handler, cfg *config.Config) {
	// consume loop, to be executed by each worker
	var consume = func(workerID int) {
		for {
			select {
			case msg, ok := <-consumer.Channels().Upstream:
				if !ok {
					log.Info(ctx, "upstream channel closed - closing event consumer loop", log.Data{"worker_id": workerID})
					return
				}

				processMessage(context.Background(), msg, h, cfg)

				msg.Release()
			case <-consumer.Channels().Closer:
				log.Info(ctx, "closer channel closed - closing event consumer loop ", log.Data{"worker_id": workerID})
				return
			}
		}
	}

	// workers to consume messages in parallel
	for w := 1; w <= cfg.KafkaNumWorkers; w++ {
		go consume(w)
	}
}

// processMessage unmarshals the provided kafka message into an event and calls the handler.
// After the message is handled, it is committed, by default even on error to prevent reconsumption
// of dead messages.
func processMessage(ctx context.Context, msg kafka.Message, h Handler, cfg *config.Config) {
	var e InstanceStarted

	if err := schema.InstanceStartedEvent.Unmarshal(msg.GetData(), &e); err != nil {
		log.Error(ctx, "failed to unmarshal event", err)
		msg.Commit()
		return
	}

	log.Info(ctx, "event received", log.Data{"event": e})

	if err := h.Handle(ctx, cfg, &e); err != nil {
		log.Error(ctx, "failed to handle event", err)
		msg.Commit()
		return
	}

	log.Info(ctx, "event processed - committing message", log.Data{"event": e})

	msg.Commit()
}
