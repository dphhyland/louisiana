package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Subject structure representing a subject in an event stream
type Subject struct {
	Format      string `json:"format" bson:"format"`
	Email       string `json:"email,omitempty" bson:"email,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty" bson:"phone_number,omitempty"`
}

type StreamConfig struct {
	StreamID        string    `json:"stream_id" bson:"stream_id"`
	EventsSupported []string  `json:"events_supported" bson:"events_supported"`
	EventsEndpoint  string    `json:"events_endpoint" bson:"events_endpoint"`
	Status          string    `json:"status" bson:"status"`
	Subjects        []Subject `json:"subjects,omitempty" bson:"subjects,omitempty"`
	Reason          *string   `json:"reason,omitempty" bson:"reason,omitempty"`
}

type StreamUpdatedEvent struct {
	EventType string `json:"event_type"`
	SubID     string `json:"sub_id"`
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
}

var (
	client     *mongo.Client
	collection *mongo.Collection
)

var (
	allowedStatuses = map[string]bool{
		"enabled":  true,
		"paused":   true,
		"disabled": true,
	}
)

// ValidateStatus checks if the provided status is in the list of allowed statuses
func ValidateStatus(status string) error {
	if _, ok := allowedStatuses[status]; !ok {
		return fmt.Errorf("invalid status: %s, allowed values are: enabled, paused, disabled", status)
	}
	return nil
}

func main() {
	// Get MongoDB URI from environment variable
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// Initialize MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Get the collection for storing stream configurations
	collection = client.Database("signals_db").Collection("streams")

	// Set up router
	r := chi.NewRouter()
	r.Post("/stream-config", registerStreamConfig)
	r.Put("/stream-config/{stream_id}", updateStreamStatus)
	r.Get("/stream-config/{stream_id}", getStreamStatus)
	r.Post("/ssf/subjects:add", addSubjectToStream)         // Add subject
	r.Post("/ssf/subjects:remove", removeSubjectFromStream) // Remove subject

	// Set up HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Println("Server listening on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Graceful shutdown
	waitForShutdown(server)
}

func waitForShutdown(server *http.Server) {
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func registerStreamConfig(w http.ResponseWriter, r *http.Request) {
	var streamConfig StreamConfig
	if err := json.NewDecoder(r.Body).Decode(&streamConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if len(streamConfig.EventsSupported) == 0 || streamConfig.EventsEndpoint == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Set initial status to "enabled"
	streamConfig.Status = "enabled"
	streamConfig.StreamID = generateStreamID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Store the stream configuration in MongoDB
	_, err := collection.InsertOne(ctx, streamConfig)
	if err != nil {
		http.Error(w, "Failed to register stream configuration", http.StatusInternalServerError)
		log.Printf("Error registering stream configuration: %v", err)
		return
	}

	log.Printf("Stream configuration registered with StreamID: %s", streamConfig.StreamID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(streamConfig)
}

func updateStreamStatus(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "stream_id")
	var updateRequest struct {
		Status string  `json:"status"`
		Reason *string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate the status field
	if err := ValidateStatus(updateRequest.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Update the stream status in MongoDB
	filter := bson.M{"stream_id": streamID}
	update := bson.M{"$set": bson.M{"status": updateRequest.Status}}

	if updateRequest.Reason != nil {
		update["$set"].(bson.M)["reason"] = updateRequest.Reason
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Error updating stream status: %v", err)
		http.Error(w, "Failed to update stream status", http.StatusInternalServerError)
		return
	}

	// Fetch the updated stream configuration
	var updatedStreamConfig StreamConfig
	err = collection.FindOne(ctx, filter).Decode(&updatedStreamConfig)
	if err != nil {
		log.Printf("Error fetching updated stream configuration: %v", err)
		http.Error(w, "Failed to fetch updated stream configuration", http.StatusInternalServerError)
		return
	}

	// Send the stream-updated event
	sendStreamUpdatedEvent(updatedStreamConfig, updateRequest.Reason)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Stream status updated and event sent"))
}

func getStreamStatus(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "stream_id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the stream configuration by stream_id
	var streamConfig StreamConfig
	err := collection.FindOne(ctx, bson.M{"stream_id": streamID}).Decode(&streamConfig)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Stream configuration not found", http.StatusNotFound)
			return
		}
		log.Printf("Error fetching stream configuration: %v", err)
		http.Error(w, "Failed to fetch stream configuration", http.StatusInternalServerError)
		return
	}

	// Return the stream status
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(streamConfig)
}

// addSubjectToStream handles adding a subject to a stream as per SSF 7.1.3.1
func addSubjectToStream(w http.ResponseWriter, r *http.Request) {
	var request struct {
		StreamID string  `json:"stream_id"`
		Subject  Subject `json:"subject"`
		Verified *bool   `json:"verified,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the stream configuration by stream_id
	filter := bson.M{"stream_id": request.StreamID}
	update := bson.M{"$push": bson.M{"subjects": request.Subject}}

	result := collection.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			http.Error(w, "Stream not found", http.StatusNotFound)
			return
		}
		log.Printf("Error adding subject to stream: %v", result.Err())
		http.Error(w, "Failed to add subject to stream", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// removeSubjectFromStream handles removing a subject from a stream as per SSF 7.1.3.2
func removeSubjectFromStream(w http.ResponseWriter, r *http.Request) {
	var request struct {
		StreamID string  `json:"stream_id"`
		Subject  Subject `json:"subject"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Remove the subject from the stream
	filter := bson.M{"stream_id": request.StreamID}
	update := bson.M{"$pull": bson.M{"subjects": request.Subject}}

	result := collection.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			http.Error(w, "Stream not found", http.StatusNotFound)
			return
		}
		log.Printf("Error removing subject from stream: %v", result.Err())
		http.Error(w, "Failed to remove subject from stream", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func sendStreamUpdatedEvent(streamConfig StreamConfig, reason *string) {
	event := StreamUpdatedEvent{
		EventType: "https://schemas.openid.net/secevent/ssf/event-type/stream-updated",
		SubID:     streamConfig.StreamID,
		Status:    streamConfig.Status,
		Reason:    *reason,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error encoding stream-updated event: %v", err)
		return
	}

	// Send the event to the stream endpoint
	req, err := http.NewRequest("POST", streamConfig.EventsEndpoint, bytes.NewBuffer(eventData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending stream-updated event to endpoint %s: %v", streamConfig.EventsEndpoint, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Stream endpoint responded with status: %s", resp.Status)
	}
}

func generateStreamID() string {
	return fmt.Sprintf("stream-%d", time.Now().UnixNano())
}
