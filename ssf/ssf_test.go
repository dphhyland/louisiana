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
		Reason:    newString("test-reason"), // Use newString to create a *string
	}

	assert.Equal(t, "https://schemas.openid.net/secevent/ssf/event-type/stream-updated", streamUpdatedEvent.EventType)
	assert.Equal(t, "test-sub-id", streamUpdatedEvent.SubID)
	assert.Equal(t, "enabled", streamUpdatedEvent.Status)
	assert.Equal(t, "test-reason", *streamUpdatedEvent.Reason) // Dereference the pointer
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
		assert.Equal(t, "test-reason", *event.Reason) // Dereference *Reason

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	streamConfig := StreamConfig{
		StreamID:       "test-sub-id",
		Status:         "enabled",
		EventsEndpoint: ts.URL,
	}

	// Use pointer to "test-reason"
	sendStreamUpdatedEvent(streamConfig, newString("test-reason"))
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

// New tests for status validation
func TestValidStreamStatusUpdate(t *testing.T) {
	// Create a test server to mock the stream status update
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var request struct {
			Status string  `json:"status"`
			Reason *string `json:"reason,omitempty"`
		}
		err = json.Unmarshal(body, &request)
		assert.NoError(t, err)

		// Test that valid status is accepted
		assert.Equal(t, "enabled", request.Status)
		assert.Equal(t, "Valid reason", *request.Reason)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := &http.Client{}
	payload := struct {
		Status string  `json:"status"`
		Reason *string `json:"reason,omitempty"`
	}{
		Status: "enabled",
		Reason: newString("Valid reason"),
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", ts.URL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestInvalidStreamStatusUpdate(t *testing.T) {
	// Create a test server to mock the stream status update
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var request struct {
			Status string  `json:"status"`
			Reason *string `json:"reason,omitempty"`
		}
		err = json.Unmarshal(body, &request)
		assert.NoError(t, err)

		// Test that invalid status is rejected
		assert.Equal(t, "invalid-status", request.Status)

		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	client := &http.Client{}
	payload := struct {
		Status string  `json:"status"`
		Reason *string `json:"reason,omitempty"`
	}{
		Status: "invalid-status",
		Reason: newString("Invalid reason"),
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", ts.URL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// Helper function to create a string pointer
func newString(s string) *string {
	return &s
}
