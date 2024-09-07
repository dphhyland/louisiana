#!/bin/bash

# Define directories and files
CERTS_DIR="./certificates"
CLIENT_KEY="$CERTS_DIR/client.key"
CLIENT_CSR="$CERTS_DIR/client.csr"
CLIENT_CERT="$CERTS_DIR/client.crt"
CLIENT_OPENSSL_CNF="./openssl-client.cnf"
CA_CERT="$CERTS_DIR/ca.crt"
CA_KEY="$CERTS_DIR/ca.key"

# Ensure the certificates directory exists
mkdir -p "$CERTS_DIR"

# Delete existing client certificates and keys
rm -f "$CLIENT_KEY" "$CLIENT_CSR" "$CLIENT_CERT"

# Function to check if a file exists and is non-empty
check_file() {
  if [ ! -f "$1" ]; then
    echo "Error: $1 was not created."
    exit 1
  elif [ ! -s "$1" ]; then
    echo "Error: $1 is empty."
    exit 1
  fi
}

# Function to validate a certificate
validate_certificate() {
  openssl verify -CAfile "$1" "$2" > /dev/null 2>&1
  if [ $? -eq 0 ]; then
    echo "Certificate $2 is valid and correctly signed by CA."
  else
    echo "Certificate validation failed for $2."
    exit 1
  fi
}

echo "Creating client private key..."
openssl genrsa -out "$CLIENT_KEY" 2048
check_file "$CLIENT_KEY"

echo "Creating client CSR..."
openssl req -new -key "$CLIENT_KEY" -out "$CLIENT_CSR" -config "$CLIENT_OPENSSL_CNF"
check_file "$CLIENT_CSR"

echo "Signing client CSR with CA to create client certificate..."
openssl x509 -req -in "$CLIENT_CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial -out "$CLIENT_CERT" -days 365 -sha256 -extfile "$CLIENT_OPENSSL_CNF" -extensions v3_req_client
check_file "$CLIENT_CERT"

echo "Client certificate and key have been generated."
echo "Client Certificate: $CLIENT_CERT"

# Validate the client certificate against the CA
validate_certificate "$CA_CERT" "$CLIENT_CERT"