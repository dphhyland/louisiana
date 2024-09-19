#!/bin/bash

# Name of the container to check
CONTAINER_NAME="mock-telco-mtls"

# Step 1: Navigate to the 'mtls' directory
cd ../src/mtls

# Check if the container is running
if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
    echo "Container $CONTAINER_NAME is running. Stopping and removing it now..."
    
    # Stop the running container
    docker stop $CONTAINER_NAME
    
    # Remove the container
    docker rm $CONTAINER_NAME
    
    echo "Container $CONTAINER_NAME has been stopped and removed."
else
    echo "Container $CONTAINER_NAME is not running."
fi

# Step 2: Build the Docker image
echo "Building Docker image..."
docker buildx build --platform linux/amd64 -t mock-telco-mtls .

# Step 3: Run the container locally, mapping port 443 to 8443 for local testing
echo "Running the container..."
docker run -d -p 7443:443 --name $CONTAINER_NAME mock-telco-mtls

echo "Container $CONTAINER_NAME is running and accessible at https://localhost:7443"