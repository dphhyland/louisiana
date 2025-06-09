package main

import (
	"client/clientpkg"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	certFile       string
	keyFile        string
	caFile         string
	participantURL string
	cacheFile      string
	clientId       string
)

func main() {
	// Parse command-line flags
	flag.StringVar(&certFile, "cert", "certs/cert.crt", "Path to the client certificate")
	flag.StringVar(&keyFile, "key", "certs/cert.key", "Path to the client key")
	flag.StringVar(&caFile, "ca", "certs/ca.crt", "Path to the CA certificate")
	flag.StringVar(&participantURL, "participants", "https://data.sandbox.raidiam.io/participants", "URL for participant data")
	flag.StringVar(&cacheFile, "cache", "participants.json", "Path to cache file for participant data")
	flag.StringVar(&clientId, "clientId", "https://rp.sandbox.raidiam.io/openid_relying_party/b683106b-126c-4577-9041-cb869de643a4", "Client ID for authentication")

	flag.Parse()

	// Ensure at least one argument (JSON input) is provided
	if len(flag.Args()) < 1 {
		fmt.Println("Please provide input parameters in JSON format")
		os.Exit(1)
	}

	inputJSON := flag.Args()[0]

	// Parse input parameters
	var params clientpkg.InputParams
	err := json.Unmarshal([]byte(inputJSON), &params)
	if err != nil {
		log.Fatalf("Error parsing input JSON: %v", err)
	}

	// Handle telephony, bank, email, domain, and website checks
	var result clientpkg.Result
	if params.Telephony != nil {
		result, err = clientpkg.CheckTelephony(*params.Telephony, certFile, keyFile, caFile, participantURL, clientId, cacheFile)
	} else if params.Bank != nil {
		result, err = clientpkg.CheckBank(*params.Bank, certFile, keyFile, caFile, participantURL, clientId, cacheFile)
	} else if params.Email != nil {
		result, err = clientpkg.CheckEmail(*params.Email, certFile, keyFile, caFile, participantURL, clientId, cacheFile)
	} else if params.Domain != nil {
		result, err = clientpkg.CheckDomain(*params.Domain, certFile, keyFile, caFile, participantURL, clientId, cacheFile)
	} else if params.Website != nil {
		result, err = clientpkg.CheckWebsite(*params.Website, certFile, keyFile, caFile, participantURL, clientId, cacheFile)
	} else {
		fmt.Println("Error: Please provide telephony, bank, email, domain, or website data")
		os.Exit(1)
	}

	// Handle errors
	if err != nil {
		log.Fatalf("Error checking: %v", err)
	}

	// Print the result
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Error formatting result: %v", err)
	}
	fmt.Println(string(jsonResult))
}
