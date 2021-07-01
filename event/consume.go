package event

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dp-import-cantabular-dataset/schema"
	importapi "github.com/ONSdigital/dp-import-api/models"
	"github.com/ONSdigital/dp-api-clients-go/dataset"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
)

// Consume converts messages to event instances, and pass the event to the provided handler.
func (p *Processor) Consume(ctx context.Context, cg kafka.IConsumerGroup, h Handler) {
	// consume loop, to be executed by each worker
	var consume = func(workerID int) {
		for {
			select {
			case msg, ok := <-cg.Channels().Upstream:
				if !ok {
					log.Info(ctx, "upstream channel closed - closing event consumer loop", log.Data{"worker_id": workerID})
					return
				}

				ctx := context.Background()

				if errs := p.processMessage(ctx, msg, h); len(errs) != 0{
					var errdata []map[string]interface{}

					for _, err := range errs {
						errdata = append(errdata, map[string]interface{}{
							"error":    err.Error(),
							"log_data": unwrapLogData(err),
							"status_code": statusCode(err),
						})
					}

					log.Event(ctx, "failed to process message", log.ERROR, log.Data{
						"errors": errdata,
					})
				}

				msg.Release()
			case <-cg.Channels().Closer:
				log.Info(ctx, "closer channel closed - closing event consumer loop ", log.Data{"worker_id": workerID})
				return
			}
		}
	}

	// workers to consume messages in parallel
	for w := 1; w <= p.cfg.KafkaNumWorkers; w++ {
		go consume(w)
	}
}

// processMessage unmarshals the provided kafka message into an event and calls the handler.
// After the message is handled, it is committed, by default even on error to prevent reconsumption
// of dead messages.
func (p *Processor) processMessage(ctx context.Context, msg kafka.Message, h Handler) []error {
	defer msg.Commit()

	var e InstanceStarted
	s := schema.InstanceStarted

	if err := s.Unmarshal(msg.GetData(), &e); err != nil {
		return []error{
			&Error{
				err: fmt.Errorf("failed to unmarshal event: %w", err),
				logData: map[string]interface{}{
					"msg_data": msg.GetData(),
				},
			},
		}
	}

	log.Info(ctx, "event received", log.Data{"event": e})

	var errs []error

	if err := h.Handle(ctx, &e); err != nil {
		errs = append(errs, fmt.Errorf("failed to handle event: %w", err))

		if err := p.importAPI.UpdateImportJobState(ctx, e.JobID, p.cfg.ServiceAuthToken, importapi.FailedState); err != nil{
			errs = append(errs, &Error{
				err:     fmt.Errorf("failed to update job state: %w", err),
				logData: log.Data{
					"job_id": e.JobID,
				},
			})
		}

		if !instanceCompleted(err){
			if err := p.datasetAPI.PutInstanceState(ctx, p.cfg.ServiceAuthToken, e.InstanceID, dataset.StateFailed); err != nil{
				errs = append(errs, &Error{
					err:     fmt.Errorf("failed to update instance state: %w", err),
					logData: log.Data{
						"instance_id": e.InstanceID,
						"job_id":      e.JobID,
					},
				})
			}
		}

		return errs
	}

	log.Info(ctx, "event successfully processed - committing message", log.Data{"event": e})
	return nil
}
