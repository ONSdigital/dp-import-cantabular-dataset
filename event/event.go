package event

import (
	"github.com/ONSdigital/dp-import/events"
)

// InstanceStarted provides an avro structure for a Instance Started event
type InstanceStarted events.CantabularDatasetInstanceStarted

/*
type CantabularDatasetInstanceStarted struct {
	RecipeID       string `avro:"recipe_id"`
	InstanceID     string `avro:"instance_id"`
	JobID          string `avro:"job_id"`
	CantabularType string `avro:"cantabular_type"`
}
*/

type CategoryDimensionImport struct {
	JobID       string `avro:"job_id"`
	InstanceID  string `avro:"instance_id"`
	DimensionID string `avro:"dimension_id"`
}
