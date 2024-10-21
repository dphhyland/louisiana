package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os/exec"
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

// Struct for response handling
type ClientResponse struct {
	ResponseType    string
	ResponseMessage string
}

var tmpl = template.Must(template.New("index.html").Funcs(template.FuncMap{
	"ToLower": strings.ToLower, // Define ToLower function for templates
}).ParseFiles("index.html"))

func main() {
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

		var jsonData []byte
		var err error

		// Convert form data to JSON based on active tab with the correct structure
		switch data.ActiveTab {
		case "telephony":
			data.Telephony = r.FormValue("telephony")
			log.Println("Received telephony data:", data.Telephony)
			jsonData, err = json.Marshal(map[string]interface{}{
				"telephony": map[string]string{
					"mobile": data.Telephony,
				},
			})
		case "bank":
			data.BankBSB = r.FormValue("bankBSB")
			data.BankAccNo = r.FormValue("bankAccNo")
			data.BankAccName = r.FormValue("bankAccName")
			log.Println("Received bank data:", data.BankBSB, data.BankAccNo, data.BankAccName)
			jsonData, err = json.Marshal(map[string]interface{}{
				"bank": map[string]string{
					"bsb":           data.BankBSB,
					"accountNumber": data.BankAccNo,
					"accountName":   data.BankAccName,
				},
			})
		case "email":
			data.Email = r.FormValue("email")
			log.Println("Received email data:", data.Email)
			jsonData, err = json.Marshal(map[string]interface{}{
				"emailAddress": map[string]string{
					"emailAddress": data.Email,
				},
			})
		case "domain":
			data.Domain = r.FormValue("domain")
			log.Println("Received domain data:", data.Domain)
			jsonData, err = json.Marshal(map[string]interface{}{
				"domain": map[string]string{
					"domain": data.Domain,
				},
			})
		case "website":
			data.Website = r.FormValue("website")
			log.Println("Received website data:", data.Website)
			jsonData, err = json.Marshal(map[string]interface{}{
				"website": map[string]string{
					"website": data.Website,
				},
			})
		case "crm":
			data.CRMName = r.FormValue("crmName")
			data.CRMEmail = r.FormValue("crmEmail")
			data.CRMPhone = r.FormValue("crmPhone")
			data.CRMBSB = r.FormValue("crmBSB")
			data.CRMAccNo = r.FormValue("crmAccNo")
			log.Println("Received CRM data:", data.CRMName, data.CRMEmail, data.CRMPhone, data.CRMBSB, data.CRMAccNo)
			jsonData, err = json.Marshal(map[string]interface{}{
				"CRM": map[string]string{
					"name":          data.CRMName,
					"email":         data.CRMEmail,
					"phone":         data.CRMPhone,
					"bsb":           data.CRMBSB,
					"accountNumber": data.CRMAccNo,
				},
			})
		}

		if err != nil {
			log.Println("Error generating JSON input:", err)
			data.Response = ClientResponse{
				ResponseType:    "error",
				ResponseMessage: "Error generating JSON input: " + err.Error(),
			}
		} else {
			log.Println("Generated JSON input:", string(jsonData))
			clientResponse := processClientBinary(jsonData)
			data.Response = clientResponse
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

// processClientBinary executes client-binary and classifies the response
func processClientBinary(jsonData []byte) ClientResponse {
	output, err := executeClientBinary(jsonData)

	var response ClientResponse

	if err != nil {
		response.ResponseType = "error"
		response.ResponseMessage = "Error executing client_binary: " + err.Error()
		return response
	}

	// Classify the output as success, warning, info, or error
	if strings.Contains(output, "success") {
		response.ResponseType = "success"
		response.ResponseMessage = output
	} else if strings.Contains(output, "info") {
		response.ResponseType = "info"
		response.ResponseMessage = output
	} else if strings.Contains(output, "warning") {
		response.ResponseType = "warning"
		response.ResponseMessage = output
	} else {
		response.ResponseType = "error"
		response.ResponseMessage = "Unknown error: " + output
	}

	return response
}

// executeClientBinary runs the client-binary with the given JSON input as a command-line argument
func executeClientBinary(jsonData []byte) (string, error) {
	// Log the JSON input being passed to the client binary
	log.Println("Executing client-binary with JSON input as an argument:", string(jsonData))

	// Pass the JSON data directly as a command-line argument to the client-binary
	cmd := exec.Command("./client-binary", string(jsonData))

	// Capture both stdout and stderr
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	// Log the stderr output for debugging
	if stderr.Len() > 0 {
		log.Println("Error output from client-binary (stderr):", stderr.String())
	}

	// Log the stdout output
	if stdout.Len() > 0 {
		log.Println("Standard output from client-binary (stdout):", stdout.String())
	}

	// Check if there was an error running the command
	if err != nil {
		log.Println("Error during client-binary execution:", err)
		return "", err
	}

	// Return the trimmed stdout output
	return strings.TrimSpace(stdout.String()), nil
}
