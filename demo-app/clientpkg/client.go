package clientpkg

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type ThreatMetric struct {
	FraudReported bool `json:"fraudReported"`
	DaysActive    int  `json:"daysActive"`
}

type EmailThreatMetric struct {
	FraudReported       bool `json:"fraudReported"`
	DaysActive          int  `json:"daysActive"`
	SpamFlagged         bool `json:"spamFlagged"`
	DaysSinceLastUpdate int  `json:"daysSinceLastUpdate"`
}

type EmailThreatScoreResponse struct {
	Score   int               `json:"score"`
	Metrics EmailThreatMetric `json:"metrics"`
}

type DomainThreatMetric struct {
	DaysActive          int  `json:"daysActive"`
	DaysSinceLastUpdate int  `json:"daysSinceLastUpdate"`
	SpamFlagged         bool `json:"spamFlagged"`
	FraudReported       bool `json:"fraudReported"`
}

type DomainThreatScoreResponse struct {
	Score   int                `json:"score"`
	Metrics DomainThreatMetric `json:"metrics"`
}

type TelcoThreatScoreResponse struct {
	Score   int          `json:"score"`
	Metrics ThreatMetric `json:"metrics"`
}

type TelcoThreatMetric struct {
	DaysActive    int  `json:"daysActive"`
	SimSwap       bool `json:"simSwap"`
	FraudReported bool `json:"fraudReported"`
}

type WebsiteThreatMetric struct {
	DaysActive      int  `json:"daysActive"`
	SpamFlagged     bool `json:"spamFlagged"`
	PhishingFlagged bool `json:"phishingFlagged"`
}

type WebsiteThreatScoreResponse struct {
	Score   int                 `json:"score"`
	Metrics WebsiteThreatMetric `json:"metrics"`
}

type BankCheckResponse struct {
	Match         string `json:"match"`                   // "yes", "no", or "maybe"
	ActualAccount string `json:"actualAccount,omitempty"` // Only populated if "maybe"
}

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

// --------------------- Reusable Client Functions ---------------------

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
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	return &http.Client{Transport: transport}, nil
}

func FetchParticipantData(url string, filename string) ([]Organisation, error) {
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

func FindSupportedEndpoint(authServer AuthorisationServer, requiredApiType string) (string, bool) {
	for _, apiResource := range authServer.ApiResources {
		if apiResource.ApiFamilyType == requiredApiType && len(apiResource.ApiDiscoveryEndpoints) > 0 {
			return apiResource.ApiDiscoveryEndpoints[0].ApiEndpoint, true
		}
	}
	return "", false
}

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

func GetAccessToken(client *http.Client, tokenURL string, clientId string) (string, error) {
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

func ExtractHostname(endpoint string) (string, error) {
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

// --------------------- Check Functions ---------------------

func CheckTelephony(params TelephonyParams, certFile, keyFile, caFile, participantURL, clientId, cacheFile string) (Result, error) {
	client, err := CreateTLSClient(certFile, keyFile, caFile)
	if err != nil {
		return Result{}, fmt.Errorf("error creating TLS client: %v", err)
	}

	organisations, err := FetchParticipantData(participantURL, cacheFile)
	if err != nil {
		return Result{}, fmt.Errorf("error fetching participant data: %v", err)
	}

	for _, org := range organisations {
		for _, authServer := range org.AuthorisationServers {
			supportedEndpoint, apiSupported := FindSupportedEndpoint(authServer, "confirmation-of-telephony")
			if !apiSupported {
				continue
			}

			tokenEndpoint, err := FetchTokenEndpoint(client, authServer.OpenIDDiscoveryDocument)
			if err != nil {
				continue
			}

			accessToken, err := GetAccessToken(client, tokenEndpoint, clientId)
			if err != nil {
				continue
			}

			statusCode, responseBody, err := CallTelephonyEndpoint(client, supportedEndpoint, accessToken, params)
			if err != nil {
				continue
			} else if statusCode == http.StatusOK {
				hostname, err := ExtractHostname(supportedEndpoint)
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

func CheckBank(params BankParams, certFile, keyFile, caFile, participantURL, clientId, cacheFile string) (Result, error) {
	client, err := CreateTLSClient(certFile, keyFile, caFile)
	if err != nil {
		return Result{}, fmt.Errorf("error creating TLS client: %v", err)
	}

	organisations, err := FetchParticipantData(participantURL, cacheFile)
	if err != nil {
		return Result{}, fmt.Errorf("error fetching participant data: %v", err)
	}

	for _, org := range organisations {
		for _, authServer := range org.AuthorisationServers {
			supportedEndpoint, apiSupported := FindSupportedEndpoint(authServer, "bank-account-verification")
			if !apiSupported {
				continue
			}

			tokenEndpoint, err := FetchTokenEndpoint(client, authServer.OpenIDDiscoveryDocument)
			if err != nil {
				continue
			}

			accessToken, err := GetAccessToken(client, tokenEndpoint, clientId)
			if err != nil {
				continue
			}

			statusCode, responseBody, err := CallBankCheckEndpoint(client, supportedEndpoint, accessToken, params)
			if err != nil {
				continue
			} else if statusCode == http.StatusOK {
				hostname, err := ExtractHostname(supportedEndpoint)
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

func CheckEmail(params EmailParams, certFile, keyFile, caFile, participantURL, clientId, cacheFile string) (Result, error) {
	client, err := CreateTLSClient(certFile, keyFile, caFile)
	if err != nil {
		return Result{}, fmt.Errorf("error creating TLS client: %v", err)
	}

	organisations, err := FetchParticipantData(participantURL, cacheFile)
	if err != nil {
		return Result{}, fmt.Errorf("error fetching participant data: %v", err)
	}

	for _, org := range organisations {
		for _, authServer := range org.AuthorisationServers {
			supportedEndpoint, apiSupported := FindSupportedEndpoint(authServer, "confirmation-of-email")
			if !apiSupported {
				continue
			}

			tokenEndpoint, err := FetchTokenEndpoint(client, authServer.OpenIDDiscoveryDocument)
			if err != nil {
				continue
			}

			accessToken, err := GetAccessToken(client, tokenEndpoint, clientId)
			if err != nil {
				continue
			}

			statusCode, responseBody, err := CallEmailCheckEndpoint(client, supportedEndpoint, accessToken, params)
			if err != nil {
				continue
			} else if statusCode == http.StatusOK {
				hostname, err := ExtractHostname(supportedEndpoint)
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

func CheckDomain(params DomainParams, certFile, keyFile, caFile, participantURL, clientId, cacheFile string) (Result, error) {
	client, err := CreateTLSClient(certFile, keyFile, caFile)
	if err != nil {
		return Result{}, fmt.Errorf("error creating TLS client: %v", err)
	}

	organisations, err := FetchParticipantData(participantURL, cacheFile)
	if err != nil {
		return Result{}, fmt.Errorf("error fetching participant data: %v", err)
	}

	for _, org := range organisations {
		for _, authServer := range org.AuthorisationServers {
			supportedEndpoint, apiSupported := FindSupportedEndpoint(authServer, "confirmation-of-domain")
			if !apiSupported {
				continue
			}

			tokenEndpoint, err := FetchTokenEndpoint(client, authServer.OpenIDDiscoveryDocument)
			if err != nil {
				continue
			}

			accessToken, err := GetAccessToken(client, tokenEndpoint, clientId)
			if err != nil {
				continue
			}

			statusCode, responseBody, err := CallDomainCheckEndpoint(client, supportedEndpoint, accessToken, params)
			if err != nil {
				continue
			} else if statusCode == http.StatusOK {
				hostname, err := ExtractHostname(supportedEndpoint)
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

func CheckWebsite(params WebsiteParams, certFile, keyFile, caFile, participantURL, clientId, cacheFile string) (Result, error) {
	client, err := CreateTLSClient(certFile, keyFile, caFile)
	if err != nil {
		return Result{}, fmt.Errorf("error creating TLS client: %v", err)
	}

	organisations, err := FetchParticipantData(participantURL, cacheFile)
	if err != nil {
		return Result{}, fmt.Errorf("error fetching participant data: %v", err)
	}

	for _, org := range organisations {
		for _, authServer := range org.AuthorisationServers {
			supportedEndpoint, apiSupported := FindSupportedEndpoint(authServer, "confirmation-of-website")
			if !apiSupported {
				continue
			}

			tokenEndpoint, err := FetchTokenEndpoint(client, authServer.OpenIDDiscoveryDocument)
			if err != nil {
				continue
			}

			accessToken, err := GetAccessToken(client, tokenEndpoint, clientId)
			if err != nil {
				continue
			}

			statusCode, responseBody, err := CallWebsiteCheckEndpoint(client, supportedEndpoint, accessToken, params)
			if err != nil {
				continue
			} else if statusCode == http.StatusOK {
				hostname, err := ExtractHostname(supportedEndpoint)
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

// --------------------- Call Functions ---------------------

func CallTelephonyEndpoint(client *http.Client, endpoint string, accessToken string, params TelephonyParams) (int, string, error) {
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

func CallBankCheckEndpoint(client *http.Client, endpoint string, accessToken string, params BankParams) (int, string, error) {
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

func CallEmailCheckEndpoint(client *http.Client, endpoint string, accessToken string, params EmailParams) (int, string, error) {
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

func CallDomainCheckEndpoint(client *http.Client, endpoint string, accessToken string, params DomainParams) (int, string, error) {
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

func CallWebsiteCheckEndpoint(client *http.Client, endpoint string, accessToken string, params WebsiteParams) (int, string, error) {
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
