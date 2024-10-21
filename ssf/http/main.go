package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global variables
var (
	Client   *mongo.Client
	producer *kafka.Producer
	config   *Config
)

// Configuration struct
type Config struct {
	MongoURI    string
	KafkaBroker string
	JWTSecret   string
}

func main() {
	// Load configuration
	config = loadConfig()

	// Initialize Kafka Producer
	var err error
	producer, err = initKafkaProducer()
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer producer.Close()

	// Initialize MongoDB client
	Client, err = initMongoClient(config.MongoURI)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB client: %v", err)
	}
	defer func() {
		if err := Client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()

	// Initialize and start the HTTP server
	r := setupRouter(producer)
	server := &http.Server{Addr: ":8080", Handler: r}
	startServer(server)
}

// Load configuration from environment variables or defaults
func loadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found, using environment variables or defaults")
	}
	return &Config{
		MongoURI:    getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:64368"),
		JWTSecret:   getEnv("JWT_SECRET", "your-signing-secret"),
	}
}

// Retrieve environment variables with a default fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Initialize Kafka producer
func initKafkaProducer() (*kafka.Producer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:64368", "acks": "all"})
	if err != nil {
		return nil, err
	}
	return producer, nil
}

// Initialize MongoDB client
func initMongoClient(mongoURI string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return client, nil
}

// Set up the HTTP router
func setupRouter(producer KafkaProducer) *chi.Mux {
	r := chi.NewRouter()
	eventHandler := &EventProducerHandler{Producer: producer}
	r.Post("/produce-event", eventHandler.ServeHTTP)
	r.Post("/stream-config", RegisterStreamConfig)
	r.Put("/stream-config/{stream_id}", UpdateStreamStatus)
	r.Get("/stream-config/{stream_id}", GetStreamStatus)
	r.Post("/ssf/subjects:add", AddSubjectToStream)
	r.Post("/ssf/subjects:remove", RemoveSubjectFromStream)
	return r
}

// Start the HTTP server
func startServer(server *http.Server) {
	go func() {
		log.Println("Server listening on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
