package event

// InstanceStarted provides an avro structure for a Instance Started event
type InstanceStarted struct {
	RecipientName string `avro:"recipient_name"`
}
