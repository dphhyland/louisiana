package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
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

		// Verify that the body is a valid JWT (Secure Event Token)
		token, err := jwt.Parse(string(body), func(token *jwt.Token) (interface{}, error) {
			// Verify the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("your-signing-secret"), nil // Use your secret key
		})
		assert.NoError(t, err)
		assert.NotNil(t, token)

		// Check that the token contains the correct claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			assert.Equal(t, "https://schemas.openid.net/secevent/ssf/event-type/stream-updated", claims["event_type"])
			assert.Equal(t, "test-sub-id", claims["sub_id"])
			assert.Equal(t, "enabled", claims["status"])
			assert.Equal(t, "test-reason", claims["reason"])
		} else {
			t.Error("Invalid token claims")
		}

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

func generateTestJWT(eventPayload map[string]interface{}, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(eventPayload))
	return token.SignedString([]byte(secret))
}

func TestAddSubjectToStreamSection5(t *testing.T) {
	// Set up a test MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	assert.NoError(t, err)
	defer client.Disconnect(ctx)

	collection = client.Database("signals_db").Collection("streams")

	// Prepare the stream and insert it into the test database
	streamConfig := StreamConfig{
		StreamID:        "f67e39a0a4d34d56b3aa1bc4cff0069f",
		EventsSupported: []string{"event1", "event2"},
		EventsEndpoint:  "http://example.com/events",
		Status:          "enabled",
	}
	_, err = collection.InsertOne(ctx, streamConfig)
	assert.NoError(t, err)

	// Define a subject-added event as per SSF Section 5
	eventPayload := map[string]interface{}{
		"event_type": "https://schemas.openid.net/secevent/ssf/event-type/subject-added",
		"stream_id":  "f67e39a0a4d34d56b3aa1bc4cff0069f",
		"subject": map[string]interface{}{
			"format": "email",
			"email":  "example.user@example.com",
		},
		"verified": true,
		"iat":      time.Now().Unix(),
	}

	// Generate a JWT with the event payload
	secret := "your-signing-secret"
	tokenString, err := generateTestJWT(eventPayload, secret)
	assert.NoError(t, err)

	// Set up a request to add the subject
	req := httptest.NewRequest(http.MethodPost, "/ssf/subjects:add", bytes.NewBuffer([]byte(tokenString)))
	req.Header.Set("Content-Type", "application/jwt")

	// Record the response
	rec := httptest.NewRecorder()

	// Call the handler
	addSubjectToStream(rec, req)

	// Validate the response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check that the subject was added in MongoDB
	var updatedStreamConfig StreamConfig
	err = collection.FindOne(ctx, bson.M{"stream_id": "f67e39a0a4d34d56b3aa1bc4cff0069f"}).Decode(&updatedStreamConfig)
	assert.NoError(t, err)
	assert.Len(t, updatedStreamConfig.Subjects, 1)
	assert.Equal(t, "email", updatedStreamConfig.Subjects[0].Format)
	assert.Equal(t, "example.user@example.com", updatedStreamConfig.Subjects[0].Email)
}
func TestRemoveSubjectFromStreamSection5(t *testing.T) {
	// Set up a test MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	assert.NoError(t, err)
	defer client.Disconnect(ctx)

	collection = client.Database("signals_db").Collection("streams")

	// Prepare the stream with a subject and insert it into the test database
	streamConfig := StreamConfig{
		StreamID:        "f67e39a0a4d34d56b3aa1bc4cff0069f",
		EventsSupported: []string{"event1", "event2"},
		EventsEndpoint:  "http://example.com/events",
		Status:          "enabled",
		Subjects: []Subject{
			{
				Format: "email",
				Email:  "example.user@example.com",
			},
		},
	}
	_, err = collection.InsertOne(ctx, streamConfig)
	assert.NoError(t, err)

	// Define a subject-removed event as per SSF Section 5
	eventPayload := map[string]interface{}{
		"event_type": "https://schemas.openid.net/secevent/ssf/event-type/subject-removed",
		"stream_id":  "f67e39a0a4d34d56b3aa1bc4cff0069f",
		"subject": map[string]interface{}{
			"format": "email",
			"email":  "example.user@example.com",
		},
		"iat": time.Now().Unix(),
	}

	// Generate a JWT with the event payload
	secret := "your-signing-secret"
	tokenString, err := generateTestJWT(eventPayload, secret)
	assert.NoError(t, err)

	// Set up a request to remove the subject
	req := httptest.NewRequest(http.MethodPost, "/ssf/subjects:remove", bytes.NewBuffer([]byte(tokenString)))
	req.Header.Set("Content-Type", "application/jwt")

	// Record the response
	rec := httptest.NewRecorder()

	// Call the handler
	removeSubjectFromStream(rec, req)

	// Validate the response
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Check that the subject was removed in MongoDB
	var updatedStreamConfig StreamConfig
	err = collection.FindOne(ctx, bson.M{"stream_id": "f67e39a0a4d34d56b3aa1bc4cff0069f"}).Decode(&updatedStreamConfig)
	assert.NoError(t, err)
	assert.Len(t, updatedStreamConfig.Subjects, 0)
}

func setupTestDB(t *testing.T) *mongo.Collection {
	// Establish a test MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel) // Clean up the context after the test

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ensure the client is connected
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	t.Cleanup(func() {
		// Disconnect the client at the end of the test
		err := client.Disconnect(ctx)
		if err != nil {
			t.Fatalf("Failed to disconnect MongoDB client: %v", err)
		}
	})

	collection := client.Database("signals_db").Collection("streams")

	// Ensure cleanup of the collection after each test
	t.Cleanup(func() {
		_, err := collection.DeleteMany(ctx, bson.M{})
		if err != nil {
			t.Fatalf("Failed to clean up collection: %v", err)
		}
	})

	return collection
}

func TestStreamUpdatedEventSection5(t *testing.T) {
	collection := setupTestDB(t) // Set up the test database correctly

	// Insert a sample stream configuration into the test database
	streamConfig := StreamConfig{
		StreamID:        "f67e39a0a4d34d56b3aa1bc4cff0069f",
		EventsSupported: []string{"event1", "event2"},
		EventsEndpoint:  "http://example.com/events",
		Status:          "enabled",
	}
	_, err := collection.InsertOne(context.Background(), streamConfig)
	assert.NoError(t, err, "Failed to insert stream configuration into MongoDB")

	// Define a stream-updated event according to SSF Section 5
	eventPayload := map[string]interface{}{
		"event_type": "https://schemas.openid.net/secevent/ssf/event-type/stream-updated",
		"sub_id":     "f67e39a0a4d34d56b3aa1bc4cff0069f",
		"status":     "paused",
		"reason":     "Maintenance",
		"iat":        time.Now().Unix(),
	}

	// Generate a JWT containing the event payload
	secret := "your-signing-secret"
	tokenString, err := generateTestJWT(eventPayload, secret)
	assert.NoError(t, err, "Failed to generate JWT for event payload")

	// Create a request to update the stream status
	req := httptest.NewRequest(http.MethodPut, "/stream-config/f67e39a0a4d34d56b3aa1bc4cff0069f", bytes.NewBuffer([]byte(tokenString)))
	req.Header.Set("Content-Type", "application/jwt")

	// Record the response
	rec := httptest.NewRecorder()

	// Call the updateStreamStatus handler
	updateStreamStatus(rec, req)

	// Validate the response code
	assert.Equal(t, http.StatusOK, rec.Code, "Expected HTTP 200 OK but got a different response")

	// Verify that the stream status was correctly updated in MongoDB
	var updatedStreamConfig StreamConfig
	err = collection.FindOne(context.Background(), bson.M{"stream_id": "f67e39a0a4d34d56b3aa1bc4cff0069f"}).Decode(&updatedStreamConfig)
	assert.NoError(t, err, "Failed to retrieve updated stream configuration from MongoDB")

	// Check that the status and reason were updated correctly
	assert.Equal(t, "paused", updatedStreamConfig.Status, "Stream status should be 'paused'")
	if updatedStreamConfig.Reason != nil {
		assert.Equal(t, "Maintenance", *updatedStreamConfig.Reason, "Stream reason should be 'Maintenance'")
	} else {
		t.Error("Expected Reason to be 'Maintenance' but got nil")
	}
}

// Helper function to create a string pointer
func newString(s string) *string {
	return &s
}
