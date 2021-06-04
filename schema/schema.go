package schema

// TODO: remove or replace hello called structure and model with app specific
var InstanceStarted = `{
  "type": "record",
  "name": "cantabular-dataset-instance-started",
  "fields": [
    {"name": "recipient_name", "type": "string", "default": ""},
    {"name": "collection_id",  "type": "string", "default": ""},
    {"name": "datablob_name",  "type": "string", "default": ""}
  ]
}`
