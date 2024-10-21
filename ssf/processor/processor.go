package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/joho/godotenv"
	"github.com/trustframeworks/oauthclient"
)

func main() {
	// Initialize Kafka consumer configuration
	config := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:64368",
		"group.id":          "stream-updated-event-group",
		"auto.offset.reset": "earliest",
	}

	// Create a new Kafka consumer
	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Subscribe to the topic
	topic := "stream-updated-events"
	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	// Create a channel to handle OS interrupts (e.g., Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Listening for Kafka events on topic: %s\n", topic)

	run := true
	for run {
		select {
		case sig := <-sigChan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := consumer.Poll(100)
			switch e := ev.(type) {
			case *kafka.Message:
				fmt.Printf("Received message: %s\n", string(e.Value))
				// Process the message
				handleStreamUpdatedEvent(string(e.Value))

			case kafka.Error:
				fmt.Fprintf(os.Stderr, "Error: %v\n", e)
				run = false
			}
		}
	}
}

// handleStreamUpdatedEvent processes the Kafka event and calls the receiver
func handleStreamUpdatedEvent(message string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Create a new OAuthClient
	client, err := oauthclient.NewOAuthClient(
		os.Getenv("CERT_FILE"),
		os.Getenv("KEY_FILE"),
		os.Getenv("CA_FILE"),
		os.Getenv("WELL_KNOWN_URL"),
		os.Getenv("CLIENT_ID"),
	)
	if err != nil {
		log.Fatalf("Failed to create OAuth client: %v", err)
	}

	// Fetch the access token
	err = client.FetchAccessToken()
	if err != nil {
		log.Fatalf("Failed to fetch access token: %v", err)
	}

	// Make an HTTP request to your receiver using the access token
	err = callReceiverEndpoint(client.AccessToken, "https://receiver-endpoint/api/events", message)
	if err != nil {
		log.Printf("Failed to call receiver endpoint: %v", err)
	}
}

func callReceiverEndpoint(accessToken string, endpoint string, message string) error {
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(message))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call receiver endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("receiver endpoint responded with status: %s", resp.Status)
	}

	log.Println("Successfully called the receiver endpoint")
	return nil
}
