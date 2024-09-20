package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestStreamConfig(t *testing.T) {
	streamConfig := StreamConfig{
		StreamID:        "test-stream",
		EventsSupported: []string{"event1", "event2"},
		EventsEndpoint:  "http://example.com/events",
		Status:          "enabled",
	}

	assert.Equal(t, "test-stream", streamConfig.StreamID)
	assert.Equal(t, []string{"event1", "event2"}, streamConfig.EventsSupported)
	assert.Equal(t, "http://example.com/events", streamConfig.EventsEndpoint)
	assert.Equal(t, "enabled", streamConfig.Status)
}

func TestStreamUpdatedEvent(t *testing.T) {
	streamUpdatedEvent := StreamUpdatedEvent{
		EventType: "https://schemas.openid.net/secevent/ssf/event-type/stream-updated",
		SubID:     "test-sub-id",
		Status:    "enabled",
		Reason:    "test-reason",
	}

	assert.Equal(t, "https://schemas.openid.net/secevent/ssf/event-type/stream-updated", streamUpdatedEvent.EventType)
	assert.Equal(t, "test-sub-id", streamUpdatedEvent.SubID)
	assert.Equal(t, "enabled", streamUpdatedEvent.Status)
	assert.Equal(t, "test-reason", streamUpdatedEvent.Reason)
}

func TestMongoDBConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	assert.NoError(t, err)
	assert.NotNil(t, client)

	err = client.Ping(ctx, nil)
	assert.NoError(t, err)

	err = client.Disconnect(ctx)
	assert.NoError(t, err)
}

func TestSendStreamUpdatedEvent(t *testing.T) {
	// Create a test server to mock the stream endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var event StreamUpdatedEvent
		err = json.Unmarshal(body, &event)
		assert.NoError(t, err)

		assert.Equal(t, "https://schemas.openid.net/secevent/ssf/event-type/stream-updated", event.EventType)
		assert.Equal(t, "test-sub-id", event.SubID)
		assert.Equal(t, "enabled", event.Status)
		assert.Equal(t, "test-reason", event.Reason)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	streamConfig := StreamConfig{
		StreamID:       "test-sub-id",
		Status:         "enabled",
		EventsEndpoint: ts.URL,
	}

	sendStreamUpdatedEvent(streamConfig, "test-reason")
}

func TestGenerateStreamID(t *testing.T) {
	streamID := generateStreamID()
	assert.Regexp(t, `^stream-\d+$`, streamID)
}

func TestAddSubjectToStream(t *testing.T) {
	// Create a test server to mock the add subject endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var request struct {
			StreamID string  `json:"stream_id"`
			Subject  Subject `json:"subject"`
			Verified *bool   `json:"verified,omitempty"`
		}
		err = json.Unmarshal(body, &request)
		assert.NoError(t, err)

		assert.Equal(t, "test-stream", request.StreamID)
		assert.Equal(t, "email", request.Subject.Format)
		assert.Equal(t, "example.user@example.com", request.Subject.Email)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := &http.Client{}
	subject := Subject{
		Format: "email",
		Email:  "example.user@example.com",
	}

	payload := struct {
		StreamID string  `json:"stream_id"`
		Subject  Subject `json:"subject"`
		Verified *bool   `json:"verified,omitempty"`
	}{
		StreamID: "test-stream",
		Subject:  subject,
		Verified: nil,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", ts.URL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveSubjectFromStream(t *testing.T) {
	// Create a test server to mock the remove subject endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var request struct {
			StreamID string  `json:"stream_id"`
			Subject  Subject `json:"subject"`
		}
		err = json.Unmarshal(body, &request)
		assert.NoError(t, err)

		assert.Equal(t, "test-stream", request.StreamID)
		assert.Equal(t, "email", request.Subject.Format)
		assert.Equal(t, "example.user@example.com", request.Subject.Email)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := &http.Client{}
	subject := Subject{
		Format: "email",
		Email:  "example.user@example.com",
	}

	payload := struct {
		StreamID string  `json:"stream_id"`
		Subject  Subject `json:"subject"`
	}{
		StreamID: "test-stream",
		Subject:  subject,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", ts.URL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
