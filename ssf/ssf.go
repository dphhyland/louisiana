package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
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
	EventType string  `json:"event_type"`
	SubID     string  `json:"sub_id"`
	Status    string  `json:"status"`
	Reason    *string `json:"reason,omitempty"`
}

var (
	client     *mongo.Client
	collection *mongo.Collection
	producer   *kafka.Producer
)

var (
	allowedStatuses = map[string]bool{
		"enabled":  true,
		"paused":   true,
		"disabled": true,
	}
)

type EventProducerHandler struct {
	Producer KafkaProducer
}

func (h *EventProducerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	produceEventHandler(w, r, h.Producer)
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

// ValidateStatus checks if the provided status is in the list of allowed statuses
func ValidateStatus(status string) error {
	if _, ok := allowedStatuses[status]; !ok {
		return fmt.Errorf("invalid status: %s, allowed values are: enabled, paused, disabled", status)
	}
	return nil
}

func main() {
	var err error

	// Kafka producer initialization
	producerConfig := &kafka.ConfigMap{"bootstrap.servers": "localhost:9092"}
	producer, err = kafka.NewProducer(producerConfig)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Get MongoDB URI from environment variable
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// Initialize MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	// Check if MongoDB connection is active
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	defer func() {
		log.Println("Disconnecting MongoDB client...")
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error during MongoDB client disconnection: %v", err)
		}
	}()

	// Get the collection for storing stream configurations
	collection = client.Database("signals_db").Collection("streams")

	// Set up router
	r := chi.NewRouter()
	eventHandler := &EventProducerHandler{Producer: producer}
	r.Post("/produce-event", eventHandler.ServeHTTP)
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

// produceEventHandler handles incoming HTTP requests and sends events to the Kafka topic
// Refactored produceEventHandler to accept a Kafka producer
func produceEventHandler(w http.ResponseWriter, r *http.Request, producer KafkaProducer) {
	// Validate the request content type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	// Read and parse the request body
	var eventPayload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&eventPayload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Marshal the event payload back into JSON for Kafka
	messageBytes, err := json.Marshal(eventPayload)
	if err != nil {
		http.Error(w, "Failed to marshal event payload", http.StatusInternalServerError)
		return
	}

	// Produce the message to Kafka
	topic := "stream-updated-events"
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          messageBytes,
	}

	if err := producer.Produce(msg, nil); err != nil {
		http.Error(w, "Failed to produce Kafka message", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Event produced successfully"))
}

// Define an interface for Kafka producer to allow for easy mocking/testing
type KafkaProducer interface {
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
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
	// Parse the JWT from the request body
	claims, err := parseJWT(r, "your-signing-secret") // Use your actual signing secret
	if err != nil {
		http.Error(w, "Invalid JWT: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Extract the stream_id and subject from the claims
	streamID, ok := claims["stream_id"].(string)
	if !ok {
		http.Error(w, "Missing or invalid stream_id in JWT", http.StatusBadRequest)
		return
	}

	subjectMap, ok := claims["subject"].(map[string]interface{})
	if !ok {
		http.Error(w, "Missing or invalid subject in JWT", http.StatusBadRequest)
		return
	}

	subject := Subject{
		Format: subjectMap["format"].(string),
		Email:  subjectMap["email"].(string), // Assuming it's an email subject
	}

	// Update MongoDB to add the subject to the stream
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"stream_id": streamID}
	update := bson.M{"$push": bson.M{"subjects": subject}}

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
	// Parse the JWT from the request body
	claims, err := parseJWT(r, "your-signing-secret") // Use your actual signing secret
	if err != nil {
		http.Error(w, "Invalid JWT: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Extract the stream_id and subject from the claims
	streamID, ok := claims["stream_id"].(string)
	if !ok {
		http.Error(w, "Missing or invalid stream_id in JWT", http.StatusBadRequest)
		return
	}

	subjectMap, ok := claims["subject"].(map[string]interface{})
	if !ok {
		http.Error(w, "Missing or invalid subject in JWT", http.StatusBadRequest)
		return
	}

	subject := Subject{
		Format: subjectMap["format"].(string),
		Email:  subjectMap["email"].(string), // Assuming it's an email subject
	}

	// Remove the subject from the stream in MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"stream_id": streamID}
	update := bson.M{"$pull": bson.M{"subjects": subject}}

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
	// Create the payload for the event
	eventPayload := map[string]interface{}{
		"event_type": "https://schemas.openid.net/secevent/ssf/event-type/stream-updated",
		"sub_id":     streamConfig.StreamID, // 'sub_id' as defined in SSF, Section 7.1.2
		"status":     streamConfig.Status,   // 'status' as per SSF, Section 7.1.2
		"iat":        time.Now().Unix(),     // 'iat' is used for token issued at time
	}

	// If reason is provided, include it
	if reason != nil {
		eventPayload["reason"] = *reason
	}

	// Generate the SET (JWT) by signing the event payload
	set, err := generateSecureEventToken(eventPayload, "your-signing-secret") // Use your secret key for signing
	if err != nil {
		log.Printf("Error generating SET: %v", err)
		return
	}

	// Send the SET to the event endpoint
	req, err := http.NewRequest("POST", streamConfig.EventsEndpoint, bytes.NewBuffer([]byte(set)))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/jwt")

	client := &http.Client{}
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

func generateSecureEventToken(eventPayload map[string]interface{}, secret string) (string, error) {
	// Create a new token using the HS256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(eventPayload))

	// Sign the token using the secret
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func updateStreamStatus(w http.ResponseWriter, r *http.Request) {

	if client == nil {
		log.Println("MongoDB client is nil")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if the MongoDB client is connected
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Printf("MongoDB client is disconnected: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	streamID := chi.URLParam(r, "stream_id")
	log.Printf("Updating stream status for streamID: %s", streamID)

	var updateRequest struct {
		Status string  `json:"status"`
		Reason *string `json:"reason,omitempty"`
	}

	// Parse and validate the JWT
	tokenString, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	token, err := jwt.Parse(string(tokenString), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("your-signing-secret"), nil
	})

	if err != nil || !token.Valid {
		log.Printf("Error parsing token: %v", err)
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	// Extract claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Println("Invalid token claims")
		http.Error(w, "Invalid token claims", http.StatusBadRequest)
		return
	}

	status, ok := claims["status"].(string)
	if !ok || status == "" {
		log.Println("Missing or invalid status in JWT")
		http.Error(w, "Missing or invalid status", http.StatusBadRequest)
		return
	}

	updateRequest.Status = status
	if reason, ok := claims["reason"].(string); ok && reason != "" {
		updateRequest.Reason = &reason
	}

	log.Printf("Parsed update request: status=%s, reason=%v", updateRequest.Status, updateRequest.Reason)

	// Validate the status field
	if err := ValidateStatus(updateRequest.Status); err != nil {
		log.Printf("Invalid status: %v", err)
		http.Error(w, fmt.Sprintf("Invalid status: %v", err.Error()), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch the current stream configuration for logging before update
	var currentStreamConfig StreamConfig
	err = collection.FindOne(ctx, bson.M{"stream_id": streamID}).Decode(&currentStreamConfig)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Printf("Error finding stream configuration with stream_id %s: %v", streamID, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		log.Printf("Stream with ID %s not found, will be created", streamID)
	} else {
		log.Printf("Stream configuration before update: %+v", currentStreamConfig)
	}

	// Update the stream status in MongoDB
	filter := bson.M{"stream_id": streamID}
	update := bson.M{"$set": bson.M{"status": updateRequest.Status}}

	if updateRequest.Reason != nil {
		update["$set"].(bson.M)["reason"] = updateRequest.Reason
	} else {
		update["$unset"] = bson.M{"reason": ""}
	}

	opts := options.Update().SetUpsert(true)
	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Printf("Error updating stream status for stream_id %s: %v", streamID, err)
		http.Error(w, "Failed to update stream status in MongoDB", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 && result.UpsertedCount == 0 {
		log.Printf("Failed to update or insert stream with ID %s", streamID)
		http.Error(w, fmt.Sprintf("Failed to update or insert stream with ID %s", streamID), http.StatusInternalServerError)
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
	log.Printf("Stream configuration after update: %+v", updatedStreamConfig)

	// Send the stream-updated event
	sendStreamUpdatedEvent(updatedStreamConfig, updateRequest.Reason)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Stream status updated and event sent"))
}

func parseJWT(r *http.Request, secret string) (jwt.MapClaims, error) {
	// Read the request body to get the JWT token
	tokenString, err := io.ReadAll(r.Body) // `io.ReadAll` if you're using Go 1.16+
	if err != nil {
		return nil, err
	}

	// Parse the JWT token
	token, err := jwt.Parse(string(tokenString), func(token *jwt.Token) (interface{}, error) {
		// Ensure that the signing method is HMAC (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract and return the claims (payload)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

func generateStreamID() string {
	return fmt.Sprintf("stream-%d", time.Now().UnixNano())
}
