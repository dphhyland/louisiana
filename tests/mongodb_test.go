package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert" // Added import
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
