package oauthclient

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// OAuthClient represents the client used to interact with OAuth2-protected endpoints
type OAuthClient struct {
	HTTPClient  *http.Client
	TokenURL    string
	ClientID    string
	ClientCert  string
	ClientKey   string
	CACert      string
	AccessToken string
}

// NewOAuthClient initializes and returns a new OAuthClient
func NewOAuthClient(certFile, keyFile, caFile, wellKnownURL, clientID string) (*OAuthClient, error) {
	tlsClient, err := CreateTLSClient(certFile, keyFile, caFile)
	if err != nil {
		return nil, fmt.Errorf("error creating TLS client: %v", err)
	}

	tokenURL, err := FetchTokenEndpoint(tlsClient, wellKnownURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching token endpoint: %v", err)
	}

	return &OAuthClient{
		HTTPClient: tlsClient,
		TokenURL:   tokenURL,
		ClientID:   clientID,
		ClientCert: certFile,
		ClientKey:  keyFile,
		CACert:     caFile,
	}, nil
}

// createTLSClient sets up a TLS-enabled HTTP client
func CreateTLSClient(certFile, keyFile, caFile string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate and key: %v", err)
	}

	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	return &http.Client{Transport: transport}, nil
}

// fetchTokenEndpoint retrieves the OAuth token endpoint from the .well-known configuration
func FetchTokenEndpoint(client *http.Client, wellKnownURL string) (string, error) {

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

func (c *OAuthClient) FetchRichAccessToken(authorizationDetails string) error {
	// Print the authorization details
	fmt.Printf("Authorization Details: %s\n", authorizationDetails)

	// Use url.Values to properly encode the form data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.ClientID)
	data.Set("authorization_details", authorizationDetails)

	// Print the encoded data string
	fmt.Printf("Data: %s\n", data.Encode())

	// Create the HTTP request
	req, err := http.NewRequest("POST", c.TokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Construct the equivalent curl command
	curlCommand := fmt.Sprintf("curl -X POST '%s' \\\n  -H 'Content-Type: application/x-www-form-urlencoded' \\\n  -d '%s'", c.TokenURL, data.Encode())
	fmt.Printf("Curl Command:\n%s\n", curlCommand)

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch token, status code: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return fmt.Errorf("failed to parse token response: %v", err)
	}

	c.AccessToken = tokenResp.AccessToken

	// Print the access token
	fmt.Printf("Access Token: %s\n", c.AccessToken)
	return nil
}

// FetchAccessToken retrieves the OAuth access token using the client credentials grant
func (c *OAuthClient) FetchAccessToken() error {
	// Use url.Values to properly encode the form data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.ClientID)

	// Create the HTTP request
	req, err := http.NewRequest("POST", c.TokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Print the equivalent curl command for debugging
	curlCommand := fmt.Sprintf("curl -X POST '%s' \\\n  -H 'Content-Type: application/x-www-form-urlencoded' \\\n  -d '%s'", c.TokenURL, data.Encode())
	fmt.Printf("Curl Command:\n%s\n", curlCommand)

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch token, status code: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return fmt.Errorf("failed to parse token response: %v", err)
	}

	c.AccessToken = tokenResp.AccessToken

	// Print the access token
	fmt.Printf("Access Token: %s\n", c.AccessToken)
	return nil
}
