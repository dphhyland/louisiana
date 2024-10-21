package oauthclient

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockHTTPClient is a mock HTTP client for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestNewOAuthClient(t *testing.T) {
	// Create temporary certificate files for testing
	certFile, keyFile, caFile := createTempCertFiles(t)
	defer func() {
		_ = os.Remove(certFile)
		_ = os.Remove(keyFile)
		_ = os.Remove(caFile)
	}()

	// Create a test server to mock the well-known endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token_endpoint": "https://example.com/token"})
	}))
	defer ts.Close()

	// Use the actual NewOAuthClient function
	client, err := NewOAuthClient(certFile, keyFile, caFile, ts.URL, "client-id")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://example.com/token", client.TokenURL)
	assert.Equal(t, "client-id", client.ClientID)
}

func TestCreateTLSClient(t *testing.T) {
	certFile, keyFile, caFile := createTempCertFiles(t)
	defer func() {
		_ = os.Remove(certFile)
		_ = os.Remove(keyFile)
		_ = os.Remove(caFile)
	}()

	client, err := CreateTLSClient(certFile, keyFile, caFile)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestFetchTokenEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token_endpoint": "https://example.com/token"})
	}))
	defer ts.Close()

	tokenURL, err := FetchTokenEndpoint(http.DefaultClient, ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/token", tokenURL)
}

func TestFetchAccessToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"access_token": "test-access-token"})
	}))
	defer ts.Close()

	client := &OAuthClient{
		HTTPClient: http.DefaultClient,
		TokenURL:   ts.URL,
		ClientID:   "client-id",
	}

	err := client.FetchAccessToken()
	assert.NoError(t, err)
	assert.Equal(t, "test-access-token", client.AccessToken)
}

// Helper function to create temporary certificate files for testing
func createTempCertFiles(t *testing.T) (string, string, string) {
	certFile, err := ioutil.TempFile("", "cert.pem")
	assert.NoError(t, err)
	_, err = certFile.Write([]byte(testCert))
	assert.NoError(t, err)
	assert.NoError(t, certFile.Close())

	keyFile, err := ioutil.TempFile("", "key.pem")
	assert.NoError(t, err)
	_, err = keyFile.Write([]byte(testKey))
	assert.NoError(t, err)
	assert.NoError(t, keyFile.Close())

	caFile, err := ioutil.TempFile("", "ca.pem")
	assert.NoError(t, err)
	_, err = caFile.Write([]byte(testCA))
	assert.NoError(t, err)
	assert.NoError(t, caFile.Close())

	return certFile.Name(), keyFile.Name(), caFile.Name()
}

// Test certificates for testing purposes
const (
	testCert = `-----BEGIN CERTIFICATE-----
MIIGFTCCBP2gAwIBAgIUXIHaeHUKGPnM5rysYEuZAx48O+MwDQYJKoZIhvcNAQEL
BQAwYDELMAkGA1UEBhMCR0IxHTAbBgNVBAoTFFJhaWRpYW0gc2VydmljZXMgbHRk
MRAwDgYDVQQLEwdyYWlkaWFtMSAwHgYDVQQDExdyYWlkaWFtIElzc3VpbmcgQ0Eg
LSBHMjAeFw0yNDA5MDgwNTA0MDBaFw0yNTEwMDgwNTA0MDBaMH0xCzAJBgNVBAYT
AlVLMRAwDgYDVQQKEwdSYWlkaWFtMS0wKwYDVQQLEyQ3MmMxNjVkYS01YzdlLTQ0
MjktOTU4Zi1jNmIyZGJkNWQ0ZmIxLTArBgNVBAMTJGI2ODMxMDZiLTEyNmMtNDU3
Ny05MDQxLWNiODY5ZGU2NDNhNDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBALxtADbjpzcQr08KDsOeS6IJ8YHqwHq3WTeBgDiPT6ufgPr6r5xyrGfZsuqw
RlwfQF0h2olWSrf5Y9nqIhY8we2QiTcFRcJnA3OkMyyxbt6+iOJzea5xuRrblkbj
ZyLtyp4oo0+Br0vKE/XDWCZX9eCEJkBxci67UWSji0BmQs8vRk/j1ROPP9sKoqyF
Eqv3UUge6ie0HjZMcGWnlC/5E/oMXe1NGrVczyT8txktFdM40yX2Nu0+8zfBO6EL
ReXG3ADM53KZoLaF+n4Y0nE7rV/mofHwyxBUV4GDwCaepOHQArKaObbg2+0t4Vm9
X5bNBlklP0ucEnhhxj3GSC+DJtUCAwEAAaOCAqgwggKkMA4GA1UdDwEB/wQEAwID
qDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0TAQH/BAIwADAd
BgNVHQ4EFgQUdfDjub5eu95i9bQ4Bezi95DZqfQwHwYDVR0jBBgwFoAUivWvC3L3
QbSrnlF239IM22W7wQYwQQYIKwYBBQUHAQEENTAzMDEGCCsGAQUFBzABhiVodHRw
Oi8vb2NzcC5wa2ktZzIuc2FuZGJveC5yYWlkaWFtLmlvMEAGA1UdHwQ5MDcwNaAz
oDGGL2h0dHA6Ly9jcmwucGtpLWcyLnNhbmRib3gucmFpZGlhbS5pby9pc3N1ZXIu
Y3JsMIIBngYDVR0gBIIBlTCCAZEwggGNBgsrBgEEAYO6L24BAjCCAXwwggE2Bggr
BgEFBQcCAjCCASgMggEkVGhpcyBDZXJ0aWZpY2F0ZSBpcyBzb2xlbHkgZm9yIHVz
ZSB3aXRoIFJhaWRpYW0gU2VydmljZXMgTGltaXRlZCBhbmQgb3RoZXIgcGFydGlj
aXBhdGluZyBvcmdhbmlzYXRpb25zIHVzaW5nIFJhaWRpYW0gU2VydmljZXMgTGlt
aXRlZHMgVHJ1c3QgRnJhbWV3b3JrIFNlcnZpY2VzLiBJdHMgcmVjZWlwdCwgcG9z
c2Vzc2lvbiBvciB1c2UgY29uc3RpdHV0ZXMgYWNjZXB0YW5jZSBvZiB0aGUgUmFp
ZGlhbSBTZXJ2aWNlcyBMdGQgQ2VydGljaWNhdGUgUG9saWN5IGFuZCByZWxhdGVk
IGRvY3VtZW50cyB0aGVyZWluLjBABggrBgEFBQcCARY0aHR0cDovL3JlcG9zaXRv
cnkucGtpLWcyLnNhbmRib3gucmFpZGlhbS5pby9wb2xpY2llczANBgkqhkiG9w0B
AQsFAAOCAQEAdE+HI2cpG09TXcYnYwqIOXuIQ79L1ByHPHttsAMGnIB1hXIE3JGe
xu6SVJ06wQ2U/WGYX5wFVumYNrk0iy2gVhQBH1BTuoqXOLEsTOrroh2WUbCiA/r/
5lRnzgc17GHtRRGOFroYMF0A7XRmzF7eT11C3o2k4KckxmddeNk3FmUKhXu3Ua6e
CXwmiMx53l0amhQJEhPCJgl0nRNNbVWyNh5GUCa+il5brSXw8uUNYDAqERc+lwEB
9G5IvONZoqu8YKUrvVW6eNp/ZncxNDMe0I1kj7UgD06+ymJ3x1tfpGOof9Z3k5lN
GvJn6DS12+2o3HH+GxPZ0dlXh6bq962TEQ==
-----END CERTIFICATE-----`

	testKey = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC8bQA246c3EK9P
Cg7DnkuiCfGB6sB6t1k3gYA4j0+rn4D6+q+ccqxn2bLqsEZcH0BdIdqJVkq3+WPZ
6iIWPMHtkIk3BUXCZwNzpDMssW7evojic3mucbka25ZG42ci7cqeKKNPga9LyhP1
w1gmV/XghCZAcXIuu1Fko4tAZkLPL0ZP49UTjz/bCqKshRKr91FIHuontB42THBl
p5Qv+RP6DF3tTRq1XM8k/LcZLRXTONMl9jbtPvM3wTuhC0XlxtwAzOdymaC2hfp+
GNJxO61f5qHx8MsQVFeBg8AmnqTh0AKymjm24NvtLeFZvV+WzQZZJT9LnBJ4YcY9
xkgvgybVAgMBAAECggEALgMCB+VMim5JGOhwVYV8m5uI/XwZN345S4wUhvs77cFb
6i24H2CDSDLJdXTJIarB+VwQdPP89/Lu1qJNY5e/lWbzktt3PvMPzTfeBT5ov/zQ
CVhAcQA1PuE7T6EtKMLfdQKgaoRVDZFRkXR7vJVDJemvO5JYWyADzqr/EKFIFDZf
Z30xWQcsGxG79c+kor4GSIk5ro3GjK4yD//6gUE+CfbdCpYN+fXzla0FfKfcnVbd
4RluIBy9zfFDbU5hv/zwYDBomvjOD5G21Y9XkSCYyuQ1C3O0rGq8jTskqPlP/TK0
vC1I0z4dtAGKrVaMPPx05jumX5bPmVSpRUaNq92GuwKBgQD6iulm8egiDW6acNJa
PbX+pmiGcJmAs9rwquEik88vO2pZYtaWVEguNtKiCero5oDj9lN8BVTyfvEgiTir
6hszGc1R/wVySsDyIL/oKwfzyfTHrFmppFn6ZWrGESYWSUaL6GcW3ZMkKaIy5fha
I8eNW4JxB2g26htPIrEs7XRjWwKBgQDAh7XIqP2p0RE65LUe6gJwrBHvGrLohp1R
RcrHXQ++GAPP7+tSkuksZM8JG32AjyWy5VO0K7ONVAafvjB30xu/2vLOThGHhgP4
8peD8fIj/JQKy2/jngpcV6pYyg5NBJ08nMkOKCzesHtcsi8yZBPY6SbSucIGBjJl
Yfc73kKljwKBgDLq3kpsqVeaUTsT6LwsRHtvSFjiM2AjrUAyCjUjwvx/X7qwypmc
oVq7C42g5FvW1KT/n7HZx4zM3aZWHO1bU5HMEZ0zbeDvbk0G+NlvPVt/VL7ruQEw
BJVN3ShJverTk4HFhoXwHAJCb9NWR2XSVbDVwynDbpuScmBf1ZAi5f93AoGBAJcX
KWKitAbbikEEkNsE3AteDejts+9lDPpl4f/YmW0d3YgGiU9Q+WocZpmIGFKWhAhD
jg+7p/nGMjiUgebXJlTG13ttqrYHRwDMKHmkmtkA85ERG+qt8QWMyqNJVjW85ERX
6jSQ9L2CFB2nvAA4p5a3Sf9fRdOCc3Q6kFJMV1MjAoGAEef3pWO6u8wMSnYNuLVM
t/6fNwaSRU/sAK79ip9F7c8/etmwP06+OF6nlaELZroWCZ9vobJHVj81edrynv0r
fEiFQDCJ43XXjRzjCWZTnBgss/PYF6qFokWuPq8giJRDUn3Etbv32BeVPAEDXPwD
lyZVWQLBnVjUxfAK4aE7+vc=
-----END PRIVATE KEY-----
`

	testCA = `-----BEGIN CERTIFICATE-----
MIIG3DCCBMSgAwIBAgIUDXaBl+/Y6BHKdTliXwXWBuAm1l8wDQYJKoZIhvcNAQEN
BQAwXTELMAkGA1UEBhMCR0IxHTAbBgNVBAoTFFJhaWRpYW0gc2VydmljZXMgbHRk
MRAwDgYDVQQLEwdyYWlkaWFtMR0wGwYDVQQDExRyYWlkaWFtIFJvb3QgQ0EgLSBH
MjAeFw0yMjExMjQxMTU5MDBaFw0zMjExMjExMTU5MDBaMGAxCzAJBgNVBAYTAkdC
MR0wGwYDVQQKExRSYWlkaWFtIHNlcnZpY2VzIGx0ZDEQMA4GA1UECxMHcmFpZGlh
bTEgMB4GA1UEAxMXcmFpZGlhbSBJc3N1aW5nIENBIC0gRzIwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQDAyxYEbSRP+SgUUUHTMqfZmDxa0kH3VFWksspY
XxMAkFGiJKoaQzZPusxXCYt/R6Q0hknLNbzmUPs908NiJkFeUXU1rOlEHSrrJ7/O
E7jLeOWnnx3jN8rZSJODRiQYRmd0/uTrMGVG/PvN7errXI0C3q0R1XLCF1OWbca+
MfPfIDnJ85YAK7/iF2yDD5QRenm3Z09aV2thV5jJImCwSlXWt1kfZpHCiu90PEvb
kTdud1XF6lREzjyACYJyY+aJYKkb5i62iDfw7MRxU3V/slYvKI06fcMZuc5Etl14
7emKQn+Bwc82FmB/ss2YBA5DvhTSlNW+qQWGJ/jHARMemRVjAgMBAAGjggKPMIIC
izAOBgNVHQ8BAf8EBAMCAQYwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQU
ivWvC3L3QbSrnlF239IM22W7wQYwHwYDVR0jBBgwFoAUmk0SWh5eUTfUfs0AzC7Q
PRc4sTwwQQYIKwYBBQUHAQEENTAzMDEGCCsGAQUFBzABhiVodHRwOi8vb2NzcC5w
a2ktZzIuc2FuZGJveC5yYWlkaWFtLmlvMEAGA1UdHwQ5MDcwNaAzoDGGL2h0dHA6
Ly9jcmwucGtpLWcyLnNhbmRib3gucmFpZGlhbS5pby9pc3N1ZXIuY3JsMIIBngYD
VR0gBIIBlTCCAZEwggGNBgsrBgEEAYO6L24BAjCCAXwwggE2BggrBgEFBQcCAjCC
ASgMggEkVGhpcyBDZXJ0aWZpY2F0ZSBpcyBzb2xlbHkgZm9yIHVzZSB3aXRoIFJh
aWRpYW0gU2VydmljZXMgTGltaXRlZCBhbmQgb3RoZXIgcGFydGljaXBhdGluZyBv
cmdhbmlzYXRpb25zIHVzaW5nIFJhaWRpYW0gU2VydmljZXMgTGltaXRlZHMgVHJ1
c3QgRnJhbWV3b3JrIFNlcnZpY2VzLiBJdHMgcmVjZWlwdCwgcG9zc2Vzc2lvbiBv
ciB1c2UgY29uc3RpdHV0ZXMgYWNjZXB0YW5jZSBvZiB0aGUgUmFpZGlhbSBTZXJ2
aWNlcyBMdGQgQ2VydGljaWNhdGUgUG9saWN5IGFuZCByZWxhdGVkIGRvY3VtZW50
cyB0aGVyZWluLjBABggrBgEFBQcCARY0aHR0cDovL3JlcG9zaXRvcnkucGtpLWcy
LnNhbmRib3gucmFpZGlhbS5pby9wb2xpY2llczANBgkqhkiG9w0BAQ0FAAOCAgEA
ZzaYHVxPGLgek9+TJ+GECWluDtCxQDeWKY2cksM4GaK8tnp03S0U+hn3MALmmVW0
CoLes/sbBtEdOD+tPTGvVXFvPT4b2ZiSeuMN/vYqOwzk6bjwk6xH9evU5qTVB544
WmhFkyn5qa3/dvRm4fggxArCnneYnRhZ7mzmt0QonsXLBnx/pO+3znyRlt7+r2i9
w17Aj+7obqqJVEn6zSLFDQyM5gk1GqUV0C+BI4QaodB48gCHQhfhIQ4YAr+Em5+1
mxfNCnGYIbok2Lg/QJOCucjPtyJdPUw431xgVr6J/X+fwvVvYxSMXl9qPYlI0lsG
1Mb19U/nj9LMP3m8GKnPESBl+Y0aC/ouDDtsUJQq2U40w+lXo1hlK9Cp4PrjpnFP
0JZfkiMgd51kaYarJa5of0Sf2ygtYwwNmfouBG90VSfd8fzL0x7DleYTzOwgokXk
/H7FnYQ6uGUCAA/IazkjiKnGXe5j2eJkXfrfx8g9IxitkpcdX93FOuOftk/maJ8S
9DezpZGTu/8zsyXcjMYzsuyC3TejVolYauiXQlW0vtAaRHspburhR//VrgeAVm9q
s/OpZnoz5jpn64x+G8XaHbjkWtRFYA9GWH7mfGvwbF7Cjqi6DnjeUJxtwZFroStx
PfdbP+2Q7SPDEYAfjlwlWhT1rITXQyMi4Q5B4mvZSSE=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIFijCCA3KgAwIBAgIUGC7beBn7UGSmBVEHUlf186Gsn0gwDQYJKoZIhvcNAQEN
BQAwXTELMAkGA1UEBhMCR0IxHTAbBgNVBAoTFFJhaWRpYW0gc2VydmljZXMgbHRk
MRAwDgYDVQQLEwdyYWlkaWFtMR0wGwYDVQQDExRyYWlkaWFtIFJvb3QgQ0EgLSBH
MjAeFw0yMjExMjQxMTU5MDBaFw0zNzExMjAxMTU5MDBaMF0xCzAJBgNVBAYTAkdC
MR0wGwYDVQQKExRSYWlkaWFtIHNlcnZpY2VzIGx0ZDEQMA4GA1UECxMHcmFpZGlh
bTEdMBsGA1UEAxMUcmFpZGlhbSBSb290IENBIC0gRzIwggIiMA0GCSqGSIb3DQEB
AQUAA4ICDwAwggIKAoICAQCwTd+j6zg8rVdeibxasHKdzqdhx9mygJAuE3JdPpNu
0PQwzbkZbzxxpU23wMGje9XbIsQIkh71ONRwI3bVCMgoZiRAUkT20lPH7fFoutDa
5GbHlsy/KWB6qpbOoK3w2EXjLaru5XjdSrN4pKQ80DK9b3CBY1AgGM85MhMRi0DO
s36X04bFGAOFQShw2/ijnOICnjP9I/cyiuNt4VtiO4NmSJ8ZpZ7Yvkm9oyQ3ODSo
9EQBoxeCeDPzK5OlfikEWjxqCmCsPebBxg7s2+GPXDdPjQuk3avpgCGBa4pa5nHU
hMFDN9dH7mn60t9N0Zwh01IM040/7Uflrx162MZmCl7Aw7sKJZyIx0d0JzqUmucV
YFAId3lF6h7FEaqUYNzWR73BunfpV+HURuiwfAf/ieZxb7N631xO4T13MIdfgiRM
7g6y+UGAsvdN7ALRW9sn8Crc1WIvIARnFViot7EY6LiCaDGieQrntCFVqY8gWODn
R8O6x7dAkGz+F++q2XALI3D3kO/Sql16mWjOSf4jKlrSO4v9z3aooqjXL+YugyHe
Nur8X/VEWp7oOv2TUlVenQm28gbyrO5ObFMm8hOl+6H0kNeXLdn5yYp9W5FAjBjM
LNh1xEWArzHIVXM1639izJ3ZQBj6syc42HCw0TIMyzVQXxrM+pAm5xAcJkIVbOwY
kQIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNV
HQ4EFgQUmk0SWh5eUTfUfs0AzC7QPRc4sTwwDQYJKoZIhvcNAQENBQADggIBAE/G
ElmDGRfl49Lamd97obrDTUsZS3jUjLu/fh1TFfbuzgkw/kLGNQ+Tn+4piDN37ZZR
tyObiiXbjN//7Fk0vtAeXcX/BQxsNKfyrzmmlMJ6G58EIQivF2IaSJzlyrWVDuDA
WFkK8xr1DdS4LepTsV5+MrhnvKqwI8otP3IVEsgJE0Oc0/Uco72/0O+KGu940NfD
+Wh284qdpEj9r5n2RpSU1sVoR9GNO+WQ1XEOb8p5L5mJTEvhLBA0BwOgg+cDU420
PF5LA7Cdo4RaNrhMNEthGED1+AGyBpVLImzzw1wuZQmnS0VCbKkYbEGhi+cVxnaU
IQxdgb1NX9C1cnjxtt41Wo8z7htLfb0ZYl9NCyj2U3hxpuu6WG718ftCX+0vVXWY
3YohxXnEWtlJ9tGSS4f8vnjm9z0fMA4IGpCHQCud2yU+bawXjofUaQ/hyJscuWK/
ghFh0OnDBCpq5cfQ1nmLTrCLyDbZycywJ6nDiFY7TB7gPZgBtqwhq0LxVMjOHSMk
qpXPNIlUNoLv4JpxgQX7N8vamKov/YgryqvG4DCzBWcX0Uvm/gEnk4Wlkf/etYPP
42PUUKEk41f8rbG4WNcJhFPF3HOdjdEd1MBKlIHnYbMIaJjnEBc7yPrJLTZR5SCu
orVECHTRjePlG5WNKggG+1z4QdBfOV7LrVIl6eyI
-----END CERTIFICATE-----
`
)
