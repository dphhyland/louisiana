package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type ApiDiscoveryEndpoint struct {
	ApiEndpoint string `json:"ApiEndpoint"`
}

type ApiResource struct {
	ApiFamilyType         string                 `json:"ApiFamilyType"`
	ApiDiscoveryEndpoints []ApiDiscoveryEndpoint `json:"ApiDiscoveryEndpoints"`
}

type AuthorisationServer struct {
	AuthorisationServerId   string        `json:"AuthorisationServerId"`
	OpenIDDiscoveryDocument string        `json:"OpenIDDiscoveryDocument"`
	OrganisationId          string        `json:"OrganisationId"`
	ApiResources            []ApiResource `json:"ApiResources"`
}

type Organisation struct {
	OrganisationId       string                `json:"OrganisationId"`
	AuthorisationServers []AuthorisationServer `json:"AuthorisationServers"`
}

// Function to load CA, client cert, and key for mutual TLS
func createTLSClient(certFile, keyFile, caFile string) (*http.Client, error) {
	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate and key: %v", err)
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create a TLS config that uses the CA cert and client cert
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	tlsConfig.BuildNameToCertificate()

	// Create an HTTP client with the configured TLS
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	return client, nil
}

// Fetch token endpoint from .well-known configuration
func fetchTokenEndpoint(client *http.Client, wellKnownURL string) (string, error) {
	req, err := http.NewRequest("GET", wellKnownURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for .well-known URL: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch .well-known configuration: %v", err)
	}
	defer resp.Body.Close()

	var config struct {
		TokenEndpoint string `json:"token_endpoint"`
	}
	err = json.NewDecoder(resp.Body).Decode(&config)
	if err != nil {
		return "", fmt.Errorf("failed to decode .well-known configuration: %v", err)
	}

	return config.TokenEndpoint, nil
}

// Fetch access token using mutual TLS (tls_client_auth)
func getAccessToken(client *http.Client, tokenURL string) (string, error) {
	// Prepare request payload for client_credentials using mutual TLS
	data := "grant_type=client_credentials"

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch token, status code: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse token response: %v", err)
	}

	return tokenResp.AccessToken, nil
}

// Call the telephony endpoint with the access token
func callTelephonyEndpoint(endpoint string, accessToken string) (int, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

// Fetch and cache participant data
func fetchParticipantData(url string, filename string) ([]Organisation, error) {
	var organisations []Organisation
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// File doesn't exist, download it
		fmt.Println("Downloading participant data...")
		resp, err := http.Get(url)
		if err != nil {
			return organisations, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return organisations, err
		}

		// Write the file locally
		ioutil.WriteFile(filename, body, 0644)
	}

	// Read the cached file
	fmt.Println("Using cached participant data...")
	fileData, err := ioutil.ReadFile(filename)
	if err != nil {
		return organisations, err
	}

	err = json.Unmarshal(fileData, &organisations)
	return organisations, err
}

func main() {
	const certFile = "certs/cert.crt" // Client certificate for mutual TLS
	const keyFile = "certs/cert.key"  // Client private key for mutual TLS
	const caFile = "certs/ca.crt"     // CA certificate to trust the server
	const participantURL = "https://data.sandbox.raidiam.io/participants"
	const cacheFile = "participants.json"
	const phoneNumber = "123456789" // Example phone number

	// Create a TLS-enabled HTTP client
	client, err := createTLSClient(certFile, keyFile, caFile)
	if err != nil {
		fmt.Println("Error creating TLS client:", err)
		return
	}

	// Fetch and cache participant data
	organisations, err := fetchParticipantData(participantURL, cacheFile)
	if err != nil {
		fmt.Println("Error fetching participant data:", err)
		return
	}

	// Loop through all organisations and authorisation servers
	for _, org := range organisations {
		for _, authServer := range org.AuthorisationServers {
			// Fetch the OpenID config and token endpoint
			tokenEndpoint, err := fetchTokenEndpoint(client, authServer.OpenIDDiscoveryDocument)
			if err != nil {
				fmt.Printf("Error fetching token endpoint for %s: %v\n", authServer.AuthorisationServerId, err)
				continue
			}

			// Get access token using mutual TLS
			accessToken, err := getAccessToken(client, tokenEndpoint)
			if err != nil {
				fmt.Printf("Error getting access token for %s: %v\n", authServer.AuthorisationServerId, err)
				continue
			}

			// Loop through the API resources and endpoints for telephony
			for _, apiResource := range authServer.ApiResources {
				if apiResource.ApiFamilyType == "confirmation-of-telephony" {
					for _, apiEndpoint := range apiResource.ApiDiscoveryEndpoints {
						fmt.Println("Found telephony API endpoint:", apiEndpoint.ApiEndpoint)

						// Call the telephony endpoint
						statusCode, err := callTelephonyEndpoint(apiEndpoint.ApiEndpoint, accessToken)
						if err != nil {
							fmt.Printf("Error calling API endpoint %s: %v\n", apiEndpoint.ApiEndpoint, err)
							continue
						}

						if statusCode == http.StatusOK {
							fmt.Printf("Telephony threat score found for phone number %s at endpoint %s\n", phoneNumber, apiEndpoint.ApiEndpoint)
							return
						} else if statusCode == http.StatusNotFound {
							fmt.Printf("Telephony threat score not found at endpoint %s, trying next...\n", apiEndpoint.ApiEndpoint)
						} else {
							fmt.Printf("Unexpected status code: %d for endpoint %s\n", statusCode, apiEndpoint.ApiEndpoint)
						}

						time.Sleep(1 * time.Second) // Pause to avoid hitting the APIs too quickly
					}
				}
			}
		}
	}

	fmt.Println("No telephony threat score found for phone number", phoneNumber)
}
