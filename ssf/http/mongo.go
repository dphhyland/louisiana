package main

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDB collection variable
var collection *mongo.Collection

// Retrieve MongoDB collection
func getMongoCollection(client *mongo.Client) *mongo.Collection {
	if collection == nil {
		collection = client.Database("signals_db").Collection("streams")
	}
	return collection
}
