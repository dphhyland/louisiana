#!/bin/bash

# Define directories and files
CERT_DIR="./certificates"
CA_KEY="$CERT_DIR/ca.key"
CA_CERT="$CERT_DIR/ca.crt"
CONFIG_FILE="ca.cnf"  # CA-specific config file

# Create certificates directory if it doesn't exist
mkdir -p "$CERT_DIR"

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

echo "Creating CA private key..."
openssl genrsa -out "$CA_KEY" 4096

echo "Creating CA certificate..."
openssl req -x509 -new -nodes -key "$CA_KEY" -sha256 -days 3650 -out "$CA_CERT" -config "$CONFIG_FILE" -extensions v3_ca

echo "Validating CA certificate..."
openssl x509 -in "$CA_CERT" -noout -text

# Validate the CA certificate and private key
openssl x509 -noout -modulus -in "$CA_CERT" | openssl md5 > ca_cert.md5
openssl rsa -noout -modulus -in "$CA_KEY" | openssl md5 > ca_key.md5

if cmp -s ca_cert.md5 ca_key.md5; then
  echo "Test Passed: CA certificate matches the private key."
else
  echo "Test Failed: CA certificate does not match the private key."
  rm -f ca_cert.md5 ca_key.md5
  exit 1
fi

rm -f ca_cert.md5 ca_key.md5

# Check if the CA certificate has the correct basicConstraints for a CA
if openssl x509 -in "$CA_CERT" -noout -text | grep -q "CA:TRUE"; then
  echo "Test Passed: CA certificate has the correct basicConstraints: CA:TRUE."
else
  echo "Test Failed: CA certificate does not have the correct basicConstraints."
  exit 1
fi

echo "CA certificate and key have been generated and validated."
echo "CA Certificate: $CA_CERT"
echo "CA Private Key: $CA_KEY"