package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	testClient *mongo.Client
	testCtx    context.Context
)

func TestMain(m *testing.M) {
	var cancel context.CancelFunc
	testCtx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var err error
	testClient, err = mongo.Connect(testCtx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fmt.Printf("Failed to connect to MongoDB: %v", err)
		os.Exit(1)
	}

	// Set the global client
	client = testClient // Make sure client is a global variable in your main package

	code := m.Run()

	if err = testClient.Disconnect(testCtx); err != nil {
		fmt.Printf("Failed to disconnect MongoDB client: %v", err)
	}
	os.Exit(code)
}
