package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5" // Make sure you have this import
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	allowedStatuses = map[string]bool{
		"enabled":  true,
		"paused":   true,
		"disabled": true,
	}
)

type StreamUpdatedEvent struct {
	EventType string  `json:"event_type"`
	SubID     string  `json:"sub_id"`
	Status    string  `json:"status"`
	Reason    *string `json:"reason,omitempty"`
}

// Handle updating the stream status
func UpdateStreamStatus(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "stream_id")

	var updateRequest struct {
		Status string  `json:"status"`
		Reason *string `json:"reason,omitempty"`
	}

	// Parse and validate the JWT
	tokenString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Println("Error reading request body:", err)
		return
	}

	token, err := jwt.Parse(string(tokenString), func(token *jwt.Token) (interface{}, error) {
		// Ensure that the signing method is HMAC (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("your-signing-secret"), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		log.Println("Error parsing token:", err)
		return
	}

	// Extract claims from the token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		status, ok := claims["status"].(string)
		if !ok {
			http.Error(w, "Missing or invalid status", http.StatusBadRequest)
			log.Println("Missing or invalid status in JWT")
			return
		}

		updateRequest.Status = status
		if reason, ok := claims["reason"].(string); ok {
			updateRequest.Reason = &reason
		}
	} else {
		http.Error(w, "Invalid token claims", http.StatusBadRequest)
		log.Println("Invalid token claims")
		return
	}

	// Validate the status field
	if err := validateStatus(updateRequest.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Error validating status:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch the current stream configuration for logging before update
	var currentStreamConfig StreamConfig
	err = collection.FindOne(ctx, bson.M{"stream_id": streamID}).Decode(&currentStreamConfig)
	if err != nil {
		http.Error(w, "Stream not found", http.StatusNotFound)
		log.Printf("Error finding stream configuration with stream_id %s: %v", streamID, err)
		return
	}
	log.Printf("Stream configuration before update: %+v", currentStreamConfig)

	// Update the stream status in MongoDB
	filter := bson.M{"stream_id": streamID}
	update := bson.M{"$set": bson.M{"status": updateRequest.Status}}

	if updateRequest.Reason != nil {
		update["$set"].(bson.M)["reason"] = updateRequest.Reason
	} else {
		update["$unset"] = bson.M{"reason": ""}
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Failed to update stream status", http.StatusInternalServerError)
		log.Println("Error updating stream status in MongoDB:", err)
		return
	}

	// Check if the document was modified
	if result.MatchedCount == 0 {
		http.Error(w, "No document found to update", http.StatusNotFound)
		log.Printf("No document found for stream_id %s", streamID)
		return
	}

	// Fetch the updated stream configuration
	var updatedStreamConfig StreamConfig
	err = collection.FindOne(ctx, filter).Decode(&updatedStreamConfig)
	if err != nil {
		http.Error(w, "Failed to fetch updated stream configuration", http.StatusInternalServerError)
		log.Println("Error fetching updated stream configuration:", err)
		return
	}
	log.Printf("Stream configuration after update: %+v", updatedStreamConfig)

	// Send the stream-updated event
	SendStreamUpdatedEvent(updatedStreamConfig, updateRequest.Reason)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Stream status updated and event sent"))
}

// Handle retrieving the stream status
func GetStreamStatus(w http.ResponseWriter, r *http.Request) {
	// Extract stream ID
	streamID := chi.URLParam(r, "stream_id")

	// Fetch from MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	// Return the stream status as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(streamConfig)
}

func SendStreamUpdatedEvent(streamConfig StreamConfig, reason *string) {
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
	set, err := GenerateSecureEventToken(eventPayload, "your-signing-secret") // Use your secret key for signing
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

// Handle adding a subject to a stream
func AddSubjectToStream(w http.ResponseWriter, r *http.Request) {
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

// Handle removing a subject from a stream
func RemoveSubjectFromStream(w http.ResponseWriter, r *http.Request) {
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

// HTTP handler for registering a stream configuration
func RegisterStreamConfig(w http.ResponseWriter, r *http.Request) {
	var streamConfig StreamConfig
	if err := json.NewDecoder(r.Body).Decode(&streamConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(streamConfig.EventsSupported) == 0 || streamConfig.EventsEndpoint == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	streamConfig.Status = "enabled"
	streamConfig.StreamID = GenerateStreamID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

func validateStatus(status string) error {
	if _, ok := allowedStatuses[status]; !ok {
		return fmt.Errorf("invalid status: %s, allowed values are: enabled, paused, disabled", status)
	}
	return nil
}
