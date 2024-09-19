# Stage 1: Build the Go application
FROM golang:1.23-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules files and download dependencies
COPY go.mod ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /mtls .

# Stage 2: Create a minimal image with the compiled Go binary
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /mtls .

# Expose the application's port
EXPOSE 8080

# Set the entrypoint to the Go application binary
ENTRYPOINT ["./mtls"]
