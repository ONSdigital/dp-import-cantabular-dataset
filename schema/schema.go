package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

// TODO: remove or replace hello called structure and model with app specific
var instanceStartedEvent = `{
  "type": "record",
  "name": "cantabular-dataset-instance-started",
  "fields": [
    {"name": "recipient_name", "type": "string", "default": ""}
  ]
}`

// InstanceStartedEvent is the Avro schema for Instance Started messages.
var InstanceStartedEvent = &avro.Schema{
	Definition: instanceStartedEvent,
}
