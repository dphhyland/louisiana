#!/bin/bash

# Define directories
CERT_DIR="./certificates"

echo "Cleaning up certificates..."
rm -f "$CERT_DIR"/*.key "$CERT_DIR"/*.crt "$CERT_DIR"/*.csr "$CERT_DIR"/*.srl
echo "Cleanup complete."