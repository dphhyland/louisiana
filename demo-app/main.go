package main

import (
	// Import your client package here
	"demo-app/clientpkg"
	"flag"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// Struct for form data
type FormData struct {
	ActiveTab   string
	Telephony   string
	BankBSB     string
	BankAccNo   string
	BankAccName string
	Email       string
	Domain      string
	Website     string
	CRMName     string
	CRMEmail    string
	CRMPhone    string
	CRMBSB      string
	CRMAccNo    string
	Result      string
	TrustBadge  string
	Response    ClientResponse // New field to hold client response
}

var (
	certFile       string
	keyFile        string
	caFile         string
	participantURL string
	cacheFile      string
	clientId       string
)

// Struct for response handling
type ClientResponse struct {
	ResponseType    string
	ResponseMessage string
}

var tmpl = template.Must(template.New("index.html").Funcs(template.FuncMap{
	"ToLower": strings.ToLower, // Define ToLower function for templates
}).ParseFiles("index.html"))

func main() {

	flag.StringVar(&certFile, "cert", "certs/cert.crt", "Path to the client certificate")
	flag.StringVar(&keyFile, "key", "certs/cert.key", "Path to the client key")
	flag.StringVar(&caFile, "ca", "certs/ca.crt", "Path to the CA certificate")
	flag.StringVar(&participantURL, "participants", "https://data.sandbox.raidiam.io/participants", "URL for participant data")
	flag.StringVar(&cacheFile, "cache", "participants.json", "Path to cache file for participant data")
	flag.StringVar(&clientId, "clientId", "https://rp.sandbox.raidiam.io/openid_relying_party/b683106b-126c-4577-9041-cb869de643a4", "Client ID for authentication")

	flag.Parse()

	http.HandleFunc("/", formHandler)
	log.Println("Server started on: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	data := FormData{ActiveTab: "telephony"} // Default active tab is "telephony"

	// Log the start of request processing
	log.Println("Processing request for active tab:", r.FormValue("activeTab"))

	// Handle POST request
	if r.Method == http.MethodPost {
		data.ActiveTab = r.FormValue("activeTab")

		// Call the appropriate function based on active tab
		switch data.ActiveTab {
		case "telephony":
			data.Telephony = r.FormValue("telephony")
			log.Println("Received telephony data:", data.Telephony)
			response, err := clientpkg.CheckTelephony(clientpkg.TelephonyParams{Mobile: data.Telephony}, certFile, keyFile, caFile, participantURL, clientId, cacheFile)
			if err != nil {
				data.Response = ClientResponse{
					ResponseType:    "error",
					ResponseMessage: err.Error(),
				}
			} else {
				data.Response = ClientResponse{
					ResponseType:    "success",
					ResponseMessage: response.Response,
				}
			}

		case "bank":
			data.BankBSB = r.FormValue("bankBSB")
			data.BankAccNo = r.FormValue("bankAccNo")
			data.BankAccName = r.FormValue("bankAccName")
			log.Println("Received bank data:", data.BankBSB, data.BankAccNo, data.BankAccName)
			response, err := clientpkg.CheckBank(clientpkg.BankParams{
				BSB:           data.BankBSB,
				AccountNumber: data.BankAccNo,
				AccountName:   data.BankAccName,
			}, certFile, keyFile, caFile, participantURL, clientId, cacheFile)
			if err != nil {
				data.Response = ClientResponse{
					ResponseType:    "error",
					ResponseMessage: err.Error(),
				}
			} else {
				data.Response = ClientResponse{
					ResponseType:    "success",
					ResponseMessage: response.Response,
				}
			}

		case "email":
			data.Email = r.FormValue("email")
			log.Println("Received email data:", data.Email)
			response, err := clientpkg.CheckEmail(clientpkg.EmailParams{Email: data.Email}, certFile, keyFile, caFile, participantURL, clientId, cacheFile)
			if err != nil {
				data.Response = ClientResponse{
					ResponseType:    "error",
					ResponseMessage: err.Error(),
				}
			} else {
				data.Response = ClientResponse{
					ResponseType:    "success",
					ResponseMessage: response.Response,
				}
			}

		case "domain":
			data.Domain = r.FormValue("domain")
			log.Println("Received domain data:", data.Domain)
			response, err := clientpkg.CheckDomain(clientpkg.DomainParams{Domain: data.Domain}, certFile, keyFile, caFile, participantURL, clientId, cacheFile)
			if err != nil {
				data.Response = ClientResponse{
					ResponseType:    "error",
					ResponseMessage: err.Error(),
				}
			} else {
				data.Response = ClientResponse{
					ResponseType:    "success",
					ResponseMessage: response.Response,
				}
			}

		case "website":
			data.Website = r.FormValue("website")
			log.Println("Received website data:", data.Website)
			response, err := clientpkg.CheckWebsite(clientpkg.WebsiteParams{Website: data.Website}, certFile, keyFile, caFile, participantURL, clientId, cacheFile)
			if err != nil {
				data.Response = ClientResponse{
					ResponseType:    "error",
					ResponseMessage: err.Error(),
				}
			} else {
				data.Response = ClientResponse{
					ResponseType:    "success",
					ResponseMessage: response.Response,
				}
			}

		case "crm":
			// CRM functionality can be added here similarly
		}

		// Example logic for TrustBadge
		if data.Response.ResponseType == "success" {
			data.TrustBadge = "Valid"
			log.Println("TrustBadge set to:", data.TrustBadge)
		}
	}

	// Log the final result
	log.Println("Final result:", data.Result)

	// Render the template
	tmpl.Execute(w, data)
}
