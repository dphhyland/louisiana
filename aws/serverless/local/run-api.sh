#!/bin/bash

# Container name
CONTAINER_NAME="mock-telco-api-container"

# Step 1: Navigate to the 'api' directory
cd ../src/api

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
docker buildx build --platform linux/amd64 -t mock-telco-api .

# Step 3: Run the container locally, mapping port 8080 to the local port
echo "Running the container..."
docker run -d -p 8080:8080 --name $CONTAINER_NAME mock-telco-api

echo "Container $CONTAINER_NAME is running and accessible at http://localhost:8080"