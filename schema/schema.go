package schema

import (
	"github.com/ONSdigital/dp-import/events"
	"github.com/ONSdigital/dp-kafka/v2/avro"
)

var InstanceStarted = events.CantabularDatasetInstanceStartedSchema

// Define here until finalised and added do dp-import/events
var categoryDimensionImport = `{
  "type": "record",
  "name": "cantabular-dataset-instance-started",
  "fields": [
    {"name": "dimension_id",   "type": "string"},
    {"name": "job_id", "type": "string"},
    {"name": "instance_id", "type": "string"}
  ]
}`

var CategoryDimensionImport = &avro.Schema{
	Definition: categoryDimensionImport,
}
