package main

import (
	"context"
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
