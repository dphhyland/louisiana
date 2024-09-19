package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type Config struct {
	ThreatMetrics map[string]ThreatMetric `json:"threat_metrics"`
}

type ThreatMetric struct {
	DaysActive    int  `json:"daysActive"`
	SimSwap       bool `json:"simSwap"`
	FraudReported bool `json:"fraudReported"`
}

type PhoneNumberRequest struct {
	PhoneNumber string `json:"phoneNumber"`
}

type BankAccountRequest struct {
	BSB           string `json:"bsb"`
	AccountNumber string `json:"accountNumber"`
}

type ThreatScoreResponse struct {
	Score   int          `json:"score"`
	Metrics ThreatMetric `json:"metrics"`
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

var config Config

func main() {
	// Load config
	loadConfig()

	// Setup HTTP server
	apiPathPrefix := getEnv("API_PATH", "/au/v1.0/confirmation-of-telephony/threat-score")
	// apiBankPathPrefix := getEnv("API_BANK_PATH", "/au/v1.0/confirmation-of-bank-account/threat-score")

	http.HandleFunc(apiPathPrefix, handleThreatScore)
	http.HandleFunc(apiBankPathPrefix, handleBankScore)
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil)) // TLS proxy will handle HTTPS
}

func loadConfig() {

	// Get config path from environment variable, default to "config.json" in current directory
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	// Open the config file
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configPath, err)
	}
	defer file.Close()

	// Decode the JSON configuration into the global config variable
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Failed to decode config from %s: %v", configPath, err)
	}

	log.Println("Config successfully loaded from", configPath)
}

func handleThreatScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req PhoneNumberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Check if the phone number exists in the configuration
	metrics, found := config.ThreatMetrics[req.PhoneNumber]
	if !found {
		http.Error(w, "Phone number not found", http.StatusNotFound)
		return
	}

	score := calculateThreatScore(metrics)

	res := ThreatScoreResponse{
		Score:   score,
		Metrics: metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func calculateThreatScore(metrics ThreatMetric) int {
	score := 0

	// Add points if SIM swap occurred
	if metrics.SimSwap {
		score += 50
	}

	// Add points if fraudulent activity was reported
	if metrics.FraudReported {
		score += 30
	}

	// Adjust score based on how new the account is (lower daysActive gets higher risk)
	if metrics.DaysActive < 365 {
		score += 20 // Higher score for accounts less than 1 year old
	}

	return score
}

func handleBankScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req BankAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Check if the phone number exists in the configuration
	metrics, found := config.ThreatMetrics[req.BSB+"-"+req.AccountNumber]
	if !found {
		http.Error(w, "BSB and Account number not found", http.StatusNotFound)
		return
	}

	score := calculateBankScore(metrics)

	res := ThreatScoreResponse{
		Score:   score,
		Metrics: metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func calculateBankScore(metrics ThreatMetric) int {
	score := 0

	return score
}
