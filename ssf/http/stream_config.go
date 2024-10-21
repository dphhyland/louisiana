package main

// StreamConfig represents a stream configuration document in MongoDB
type StreamConfig struct {
	StreamID        string    `json:"stream_id" bson:"stream_id"`
	EventsSupported []string  `json:"events_supported" bson:"events_supported"`
	EventsEndpoint  string    `json:"events_endpoint" bson:"events_endpoint"`
	Status          string    `json:"status" bson:"status"`
	Subjects        []Subject `json:"subjects,omitempty" bson:"subjects,omitempty"`
	Reason          *string   `json:"reason,omitempty" bson:"reason,omitempty"`
}
