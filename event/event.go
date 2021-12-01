package event

import (
	"github.com/ONSdigital/dp-import/events"
)

// InstanceStarted provides an avro structure for an Instance Started event
type InstanceStarted events.CantabularDatasetInstanceStarted

// CategoryDimensionImport provies an avro structure for a Category Dimension Import event
type CategoryDimensionImport struct {
	JobID          string `avro:"job_id"`
	InstanceID     string `avro:"instance_id"`
	DimensionID    string `avro:"dimension_id"`
	CantabularBlob string `avro:"cantabular_blob"`
}
