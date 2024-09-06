#!/bin/bash

# Define directories and files
CERTS_DIR="./certificates"
SERVER_KEY="$CERTS_DIR/server.key"
SERVER_CSR="$CERTS_DIR/server.csr"
SERVER_CERT="$CERTS_DIR/server.crt"
SERVER_PUBLIC_KEY="$CERTS_DIR/server_public_key.pem"
SERVER_OPENSSL_CNF="./openssl-server.cnf"
CA_CERT="$CERTS_DIR/ca.crt"
CA_KEY="$CERTS_DIR/ca.key"

# Ensure the certificates directory exists
mkdir -p "$CERTS_DIR"

# Delete existing server certificates and keys
rm -f "$SERVER_KEY" "$SERVER_CSR" "$SERVER_CERT" "$SERVER_PUBLIC_KEY"

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

echo "Creating server private key..."
openssl genrsa -out "$SERVER_KEY" 2048
check_file "$SERVER_KEY"

echo "Creating server CSR..."
openssl req -new -key "$SERVER_KEY" -out "$SERVER_CSR" -config "$SERVER_OPENSSL_CNF"
check_file "$SERVER_CSR"

echo "Signing server CSR with CA to create server certificate..."
openssl x509 -req -in "$SERVER_CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial -out "$SERVER_CERT" -days 365 -sha256 -extfile "$SERVER_OPENSSL_CNF" -extensions v3_req_server
check_file "$SERVER_CERT"

echo "Server certificate and key have been generated."
echo "Server Certificate: $SERVER_CERT"

# Validate the server certificate against the CA
validate_certificate "$CA_CERT" "$SERVER_CERT"

echo "Extracting public key from server certificate..."
openssl x509 -in "$SERVER_CERT" -pubkey -noout -out "$SERVER_PUBLIC_KEY"
check_file "$SERVER_PUBLIC_KEY"
echo "Public key extracted to $SERVER_PUBLIC_KEY"