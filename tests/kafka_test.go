package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type MockKafkaProducer struct {
	ProducedMessages []*kafka.Message
}

func (m *MockKafkaProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	m.ProducedMessages = append(m.ProducedMessages, msg)
	if deliveryChan != nil {
		deliveryChan <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: msg.TopicPartition.Topic, Partition: msg.TopicPartition.Partition, Offset: kafka.Offset(1)},
			Value:          msg.Value,
		}
	}
	return nil
}

func TestProduceEventHandler(t *testing.T) {
	mockProducer := &MockKafkaProducer{}

	r := chi.NewRouter()
	r.Post("/produce-event", func(w http.ResponseWriter, r *http.Request) {
		ProduceEventHandler(w, r, mockProducer) // Ensure the handler function is exported
	})

	server := httptest.NewServer(r)
	defer server.Close()

	eventPayload := map[string]interface{}{
		"event_type": "https://schemas.openid.net/secevent/ssf/event-type/stream-updated",
		"sub_id":     "f67e39a0a4d34d56b3aa1bc4cff0069f",
		"status":     "paused",
		"reason":     "Maintenance",
		"iat":        1694974123,
	}

	eventPayloadBytes, err := json.Marshal(eventPayload)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", server.URL+"/produce-event", bytes.NewBuffer(eventPayloadBytes))
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, mockProducer.ProducedMessages, 1)
	assert.Equal(t, string(eventPayloadBytes), string(mockProducer.ProducedMessages[0].Value))
}
