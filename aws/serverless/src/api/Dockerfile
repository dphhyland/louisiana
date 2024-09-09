FROM golang:1.22
WORKDIR /app
# Download Go modules

COPY go.mod ./
# Copy the entrypoint script into the container
# COPY entrypoint.sh ./
# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./
COPY config.json ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /api_go
# Make the entrypoint script executable
# RUN chmod +x ./entrypoint.sh
EXPOSE 8080

# ENTRYPOINT ["./entrypoint.sh"]
# Run
CMD ["/api_go"]