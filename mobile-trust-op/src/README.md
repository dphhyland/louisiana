


Certainly! Here's the updated set of certificate generation scripts that incorporate the improvements and suggestions I mentioned earlier. I've combined some common tasks into reusable functions to reduce repetition and enhance readability.

### 1. **CA Certificate Generation Script (`generate_ca_cert.sh`)**

```bash
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
```

### 2. **Client Certificate Generation Script (`generate_client_cert.sh`)**

```bash
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
```

### 3. **Server Certificate Generation Script (`generate_server_cert.sh`)**

```bash
#!/bin/bash

# Define directories and files
CERTS_DIR="./certificates"
SERVER_KEY="$CERTS_DIR/server.key"
SERVER_CSR="$CERTS_DIR/server.csr"
SERVER_CERT="$CERTS_DIR/server.crt"
SERVER_OPENSSL_CNF="./openssl-server.cnf"
CA_CERT="$CERTS_DIR/ca.crt"
CA_KEY="$CERTS_DIR/ca.key"

# Ensure the certificates directory exists
mkdir -p "$CERTS_DIR"

# Delete existing server certificates and keys
rm -f "$SERVER_KEY" "$SERVER_CSR" "$SERVER_CERT"

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
```

### 4. **Clean-Up Script (`clean_up_certs.sh`)**

```bash
#!/bin/bash

# Define directories
CERT_DIR="./certificates"

echo "Cleaning up certificates..."
rm -f "$CERT_DIR"/*.key "$CERT_DIR"/*.crt "$CERT_DIR"/*.csr "$CERT_DIR"/*.srl
echo "Cleanup complete."
```

### **Usage Notes**

1. **Execution Permissions**: Ensure all scripts have executable permissions:
   ```bash
   chmod +x generate_ca_cert.sh generate_client_cert.sh generate_server_cert.sh clean_up_certs.sh
   ```

2. **Run the Scripts**: Execute each script to generate the CA, client, and server certificates:
   ```bash
   ./generate_ca_cert.sh
   ./generate_client_cert.sh
   ./generate_server_cert.sh
   ```

3. **Configuration Files**: Ensure `openssl-client.cnf`, `openssl-server.cnf`, and `ca.cnf` are correctly configured and available in the same directory as your scripts.

These updated scripts should help streamline the certificate creation process, enforce consistency, and ensure robust validation of your generated certificates.