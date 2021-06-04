package event

// InstanceStarted provides an avro structure for a Instance Started event
type InstanceStarted struct {
	RecipientName string   `avro:"recipient_name"`
	DatablobName   string  `avro:"datablob_name"`
	CollectionID  string   `avro:"collection_id"`
	//Variables     []string //`avro:"variables"`
}
