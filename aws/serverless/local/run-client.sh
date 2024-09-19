cd ../src/client

# Initialize a new Go module (if you haven't already)
go mod init myclient

# Download any necessary Go dependencies (not needed for your current code, but good practice)
go mod tidy

# Run the Go program
go run client.go