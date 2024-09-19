package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	certFile       string
	keyFile        string
	caFile         string
	participantURL string
	cacheFile      string
	clientId       string
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

type TelephonyParams struct {
	Mobile string `json:"mobile"`
}

type BankParams struct {
	BSB           string `json:"bsb"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
}

type EmailParams struct {
	Email string `json:"emailAddress"`
}

type DomainParams struct {
	Domain string `json:"domain"`
}

type WebsiteParams struct {
	Website string `json:"website"`
}

type InputParams struct {
	Telephony *TelephonyParams `json:"telephony,omitempty"`
	Bank      *BankParams      `json:"bank,omitempty"`
	Email     *EmailParams     `json:"emailAddress,omitempty"`
	Domain    *DomainParams    `json:"domain,omitempty"`
	Website   *WebsiteParams   `json:"website,omitempty"`
}

type Result struct {
	Response    string `json:"response"`
	APIHostname string `json:"api_hostname"`
}

func checkTelephony(params TelephonyParams) (Result, error) {
	client, err := createTLSClient(certFile, keyFile, caFile)
	if err != nil {
		return Result{}, fmt.Errorf("error creating TLS client: %v", err)
	}

	organisations, err := fetchParticipantData(participantURL, cacheFile)
	if err != nil {
		return Result{}, fmt.Errorf("error fetching participant data: %v", err)
	}

	for _, org := range organisations {
		for _, authServer := range org.AuthorisationServers {
			supportedEndpoint, apiSupported := findSupportedEndpoint(authServer, "confirmation-of-telephony")
			if !apiSupported {
				continue
			}

			tokenEndpoint, err := fetchTokenEndpoint(client, authServer.OpenIDDiscoveryDocument)
			if err != nil {
				continue
			}

			accessToken, err := getAccessToken(client, tokenEndpoint, clientId)
			if err != nil {
				continue
			}

			statusCode, responseBody, err := callTelephonyEndpoint(client, supportedEndpoint, accessToken, params)
			if err != nil {
				continue
			} else if statusCode == http.StatusOK {
				hostname, err := extractHostname(supportedEndpoint)
				if err != nil {
					return Result{}, fmt.Errorf("error extracting hostname: %v", err)
				}
				return Result{
					Response:    responseBody,
					APIHostname: hostname,
				}, nil
			}
		}
	}

	return Result{}, fmt.Errorf("no telephony threat score found")
}

func checkBank(params BankParams) (Result, error) {
	client, err := createTLSClient(certFile, keyFile, caFile)
	if err != nil {
		return Result{}, fmt.Errorf("error creating TLS client: %v", err)
	}

	organisations, err := fetchParticipantData(participantURL, cacheFile)
	if err != nil {
		return Result{}, fmt.Errorf("error fetching participant data: %v", err)
	}

	for _, org := range organisations {
		for _, authServer := range org.AuthorisationServers {
			supportedEndpoint, apiSupported := findSupportedEndpoint(authServer, "bank-account-verification")
			if !apiSupported {
				continue
			}

			tokenEndpoint, err := fetchTokenEndpoint(client, authServer.OpenIDDiscoveryDocument)
			if err != nil {
				continue
			}

			accessToken, err := getAccessToken(client, tokenEndpoint, clientId)
			if err != nil {
				continue
			}

			statusCode, responseBody, err := callBankCheckEndpoint(client, supportedEndpoint, accessToken, params)

			if err != nil {
				continue
			} else if statusCode == http.StatusOK {
				hostname, err := extractHostname(supportedEndpoint)
				if err != nil {
					return Result{}, fmt.Errorf("error extracting hostname: %v", err)
				}
				return Result{
					Response:    responseBody,
					APIHostname: hostname,
				}, nil
			}
		}
	}

	return Result{}, fmt.Errorf("no bank check result found")
}

func checkEmail(params EmailParams) (Result, error) {
	client, err := createTLSClient(certFile, keyFile, caFile)
	if err != nil {
		return Result{}, fmt.Errorf("error creating TLS client: %v", err)
	}

	organisations, err := fetchParticipantData(participantURL, cacheFile)
	if err != nil {
		return Result{}, fmt.Errorf("error fetching participant data: %v", err)
	}

	for _, org := range organisations {
		for _, authServer := range org.AuthorisationServers {
			supportedEndpoint, apiSupported := findSupportedEndpoint(authServer, "confirmation-of-email")
			if !apiSupported {
				continue
			}

			tokenEndpoint, err := fetchTokenEndpoint(client, authServer.OpenIDDiscoveryDocument)
			if err != nil {
				continue
			}

			accessToken, err := getAccessToken(client, tokenEndpoint, clientId)
			if err != nil {
				continue
			}

			statusCode, responseBody, err := callEmailCheckEndpoint(client, supportedEndpoint, accessToken, params)

			if err != nil {
				continue
			} else if statusCode == http.StatusOK {

				hostname, err := extractHostname(supportedEndpoint)
				if err != nil {
					return Result{}, fmt.Errorf("error extracting hostname: %v", err)
				}
				return Result{
					Response:    responseBody,
					APIHostname: hostname,
				}, nil
			}
		}
	}

	return Result{}, fmt.Errorf("no email check result found")
}
func checkDomain(params DomainParams) (Result, error) {
	client, err := createTLSClient(certFile, keyFile, caFile)
	if err != nil {
		return Result{}, fmt.Errorf("error creating TLS client: %v", err)
	}

	organisations, err := fetchParticipantData(participantURL, cacheFile)
	if err != nil {
		return Result{}, fmt.Errorf("error fetching participant data: %v", err)
	}

	for _, org := range organisations {
		for _, authServer := range org.AuthorisationServers {
			supportedEndpoint, apiSupported := findSupportedEndpoint(authServer, "confirmation-of-domain")
			if !apiSupported {
				continue
			}

			tokenEndpoint, err := fetchTokenEndpoint(client, authServer.OpenIDDiscoveryDocument)
			if err != nil {
				continue
			}

			accessToken, err := getAccessToken(client, tokenEndpoint, clientId)
			if err != nil {
				continue
			}

			statusCode, responseBody, err := callDomainCheckEndpoint(client, supportedEndpoint, accessToken, params)
			if err != nil {
				continue
			} else if statusCode == http.StatusOK {
				hostname, err := extractHostname(supportedEndpoint)
				if err != nil {
					return Result{}, fmt.Errorf("error extracting hostname: %v", err)
				}
				return Result{
					Response:    responseBody,
					APIHostname: hostname,
				}, nil
			}
		}
	}

	return Result{}, fmt.Errorf("no domain check result found")
}

func checkWebsite(params WebsiteParams) (Result, error) {
	client, err := createTLSClient(certFile, keyFile, caFile)
	if err != nil {
		return Result{}, fmt.Errorf("error creating TLS client: %v", err)
	}

	organisations, err := fetchParticipantData(participantURL, cacheFile)
	if err != nil {
		return Result{}, fmt.Errorf("error fetching participant data: %v", err)
	}

	for _, org := range organisations {
		for _, authServer := range org.AuthorisationServers {
			supportedEndpoint, apiSupported := findSupportedEndpoint(authServer, "confirmation-of-website")
			if !apiSupported {
				continue
			}

			tokenEndpoint, err := fetchTokenEndpoint(client, authServer.OpenIDDiscoveryDocument)
			if err != nil {
				continue
			}

			accessToken, err := getAccessToken(client, tokenEndpoint, clientId)
			if err != nil {
				continue
			}

			statusCode, responseBody, err := callWebsiteCheckEndpoint(client, supportedEndpoint, accessToken, params)
			if err != nil {
				continue
			} else if statusCode == http.StatusOK {
				hostname, err := extractHostname(supportedEndpoint)
				if err != nil {
					return Result{}, fmt.Errorf("error extracting hostname: %v", err)
				}
				return Result{
					Response:    responseBody,
					APIHostname: hostname,
				}, nil
			}
		}
	}

	return Result{}, fmt.Errorf("no website check result found")
}

func extractHostname(endpoint string) (string, error) {
	// If the URL doesn't start with a scheme, add a dummy one
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %v", err)
	}

	if u.Hostname() == "" {
		return "", fmt.Errorf("no hostname found in URL: %s", endpoint)
	}

	return u.Hostname(), nil
}
func createTLSClient(certFile, keyFile, caFile string) (*http.Client, error) {
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
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	return &http.Client{Transport: transport}, nil
}

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

func getAccessToken(client *http.Client, tokenURL string, clientId string) (string, error) {
	data := fmt.Sprintf("grant_type=client_credentials&client_id=%s", clientId)

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

func callTelephonyEndpoint(client *http.Client, endpoint string, accessToken string, params TelephonyParams) (int, string, error) {
	requestBody, err := json.Marshal(map[string]string{
		"phoneNumber": params.Mobile,
	})
	if err != nil {
		return 0, "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	return resp.StatusCode, string(body), nil
}

func callBankCheckEndpoint(client *http.Client, endpoint string, accessToken string, params BankParams) (int, string, error) {
	requestBody, err := json.Marshal(params)
	if err != nil {
		return 0, "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	return resp.StatusCode, string(body), nil
}

func callEmailCheckEndpoint(client *http.Client, endpoint string, accessToken string, params EmailParams) (int, string, error) {
	requestBody, err := json.Marshal(params)
	if err != nil {
		return 0, "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	return resp.StatusCode, string(body), nil
}

func callDomainCheckEndpoint(client *http.Client, endpoint string, accessToken string, params DomainParams) (int, string, error) {
	requestBody, err := json.Marshal(params)
	if err != nil {
		return 0, "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	return resp.StatusCode, string(body), nil
}

func callWebsiteCheckEndpoint(client *http.Client, endpoint string, accessToken string, params WebsiteParams) (int, string, error) {
	requestBody, err := json.Marshal(params)
	if err != nil {
		return 0, "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	log.Println("Endpoint: ", endpoint)

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	return resp.StatusCode, string(body), nil
}

func fetchParticipantData(url string, filename string) ([]Organisation, error) {
	var organisations []Organisation
	if _, err := os.Stat(filename); os.IsNotExist(err) {
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

		ioutil.WriteFile(filename, body, 0644)
	}

	fmt.Println("Using cached participant data...")
	fileData, err := ioutil.ReadFile(filename)
	if err != nil {
		return organisations, err
	}

	err = json.Unmarshal(fileData, &organisations)
	return organisations, err
}

func findSupportedEndpoint(authServer AuthorisationServer, requiredApiType string) (string, bool) {
	for _, apiResource := range authServer.ApiResources {
		if apiResource.ApiFamilyType == requiredApiType && len(apiResource.ApiDiscoveryEndpoints) > 0 {
			return apiResource.ApiDiscoveryEndpoints[0].ApiEndpoint, true
		}
	}
	return "", false
}

func main() {

	flag.StringVar(&certFile, "cert", "certs/cert.crt", "Path to the client certificate")
	flag.StringVar(&keyFile, "key", "certs/cert.key", "Path to the client key")
	flag.StringVar(&caFile, "ca", "certs/ca.crt", "Path to the CA certificate")
	flag.StringVar(&participantURL, "participants", "https://data.sandbox.raidiam.io/participants", "URL for participant data")
	flag.StringVar(&cacheFile, "cache", "participants.json", "Path to cache file for participant data")
	flag.StringVar(&clientId, "clientId", "https://rp.sandbox.raidiam.io/openid_relying_party/b683106b-126c-4577-9041-cb869de643a4", "Client ID for authentication")

	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Please provide input parameters in JSON format")
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Please provide input parameters in JSON format")
		os.Exit(1)
	}
	inputJSON := args[0]

	var params InputParams
	err := json.Unmarshal([]byte(inputJSON), &params)
	if err != nil {
		fmt.Printf("Error parsing input JSON: %v\n", err)
		os.Exit(1)
	}

	var result Result
	fmt.Println("Result ... ", inputJSON)

	if params.Telephony != nil {
		result, err = checkTelephony(*params.Telephony)
	} else if params.Bank != nil {
		result, err = checkBank(*params.Bank)
	} else if params.Email != nil {
		result, err = checkEmail(*params.Email)
	} else if params.Domain != nil {
		result, err = checkDomain(*params.Domain)
	} else if params.Website != nil {
		result, err = checkWebsite(*params.Website)
	} else {
		fmt.Println("Error: Please provide either telephony or bank data")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting result: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonResult))
}
