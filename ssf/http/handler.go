package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Handler struct for event production
type EventProducerHandler struct {
	Producer KafkaProducer
}

// ServeHTTP method for handling incoming requests and producing Kafka events
func (h *EventProducerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ProduceEventHandler(w, r, h.Producer)
}

// Produce an event to Kafka
func ProduceEventHandler(w http.ResponseWriter, r *http.Request, producer KafkaProducer) {
	if err := validateContentType(r, "application/json"); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	eventPayload, err := parseJSONBody(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	messageBytes, err := json.Marshal(eventPayload)
	if err != nil {
		http.Error(w, "Failed to marshal event payload", http.StatusInternalServerError)
		return
	}

	if err := produceKafkaMessage(producer, "stream-updated-events", messageBytes); err != nil {
		http.Error(w, "Failed to produce Kafka message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Event produced successfully"))
}

// Utility function to validate content type
func validateContentType(r *http.Request, expectedType string) error {
	if r.Header.Get("Content-Type") != expectedType {
		return fmt.Errorf("Content-Type must be %s", expectedType)
	}
	return nil
}

// Utility function to parse JSON body
func parseJSONBody(body io.Reader) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, fmt.Errorf("Invalid JSON payload")
	}
	return result, nil
}

// Produce a Kafka message
func produceKafkaMessage(producer KafkaProducer, topic string, message []byte) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}
	return producer.Produce(msg, nil)
}
