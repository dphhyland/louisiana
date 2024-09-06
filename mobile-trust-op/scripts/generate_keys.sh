#!/bin/bash

# Set the directory where your certificates are stored
CERT_DIR="./certificates"

# File names
CLIENT_CERT="$CERT_DIR/client.crt"
CLIENT_KEY="$CERT_DIR/client.key"
PUBLIC_KEY_PEM="$CERT_DIR/public_key.pem"
JWK_FILE="$CERT_DIR/client.jwks.json"

# Check if the certificates directory exists
if [ ! -d "$CERT_DIR" ]; then
  echo "Creating certificates directory..."
  mkdir -p "$CERT_DIR"
fi

# Check if client.crt and client.key exist
if [ ! -f "$CLIENT_CERT" ] || [ ! -f "$CLIENT_KEY" ]; then
  echo "Error: client.crt or client.key not found in $CERT_DIR."
  exit 1
fi

# Extract the public key from the client certificate
echo "Extracting public key from $CLIENT_CERT..."
openssl x509 -pubkey -noout -in "$CLIENT_CERT" > "$PUBLIC_KEY_PEM"

# Generate JWK from PEM using pem-jwk in Node.js
echo "Converting PEM to JWK..."
node -e "
  import fs from 'fs';
  import { pem2jwk } from 'pem-jwk';
  import crypto from 'crypto';

  const pem = fs.readFileSync('$PUBLIC_KEY_PEM', 'utf-8');
  const jwk = pem2jwk(pem);
  
  // Generate a kid based on the public key modulus and exponent
  const kid = crypto.createHash('sha256').update(jwk.n + jwk.e).digest('base64url');
  jwk.kid = kid;  // Add the generated kid to the JWK

  fs.writeFileSync('$JWK_FILE', JSON.stringify({ keys: [jwk] }, null, 2));
  console.log('JWK file created at $JWK_FILE with kid:', kid);
"

echo "Public key PEM and JWK generation completed."