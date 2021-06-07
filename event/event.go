package event

import (
	"github.com/ONSdigital/dp-import/events"
)

// InstanceStarted provides an avro structure for a Instance Started event
type InstanceStarted events.CantabularDatasetInstanceStarted

/*
type CantabularDatasetInstanceStarted struct {
	InstanceID     string `avro:"instance_id"`
	JobID          string `avro:"job_id"`
	CantabularType string `avro:"cantabular_type"`
}
*/