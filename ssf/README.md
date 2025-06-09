# SSF (Signals Stream Framework)

A robust implementation of the OpenID Shared Signals Framework specification for secure event sharing between cooperating systems.

## Overview

The SSF component provides a standardized way to share security signals and events between cooperating systems, with a focus on real-time event processing and secure communication. It implements the OpenID Shared Signals Framework specification and enables applications such as Risk Incident Sharing and Coordination (RISC) and Continuous Access Evaluation Profile (CAEP).

## Features

- Stream configuration management
- Subject registration and management
- Secure event transmission
- Multiple authentication methods (JWT, mTLS, OAuth 2.0)
- Event processing and forwarding
- Standards-compliant implementation

## Architecture

```
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│               │     │               │     │               │
│  SSF API      │────▶│  Kafka        │────▶│  SSF          │
│  Service      │     │  Message Bus  │     │  Processor    │
│               │     │               │     │               │
└───────┬───────┘     └───────────────┘     └───────┬───────┘
        │                                           │
        ▼                                           ▼
┌───────────────┐                         ┌───────────────────┐
│               │                         │                   │
│  MongoDB      │                         │  Receiver         │
│  Database     │                         │  Endpoints        │
│               │                         │                   │
└───────────────┘                         └───────────────────┘
```

### Components

1. **SSF API Service**: RESTful API for stream configuration management
2. **Event Processor**: Consumes events from Kafka and forwards to receivers
3. **MongoDB**: Stores stream configurations and subject registrations
4. **Kafka**: Message broker for reliable event delivery

## Installation

### Prerequisites
- Go 1.16+
- MongoDB
- Kafka
- TLS certificates for mTLS authentication

### Setup

1. Clone the repository:
   ```
   git clone https://github.com/yourorg/louisiana.git
   cd louisiana/ssf
   ```

2. Create a `.env` file with your configuration:
   ```
   MONGODB_URI=mongodb://localhost:27017
   CERT_FILE=certs/cert.crt
   KEY_FILE=certs/cert.key
   CA_FILE=certs/ca.crt
   WELL_KNOWN_URL=https://your-auth-server/.well-known/openid-configuration
   CLIENT_ID=your-client-id
   ```

3. Build the components:
   ```
   # Build the API service
   go build -o ssf ./
   
   # Build the processor
   cd processor && go build -o processor ./
   ```

## Usage

### Starting the Services

1. Start the API service:
   ```
   ./ssf
   ```

2. Start the processor:
   ```
   cd processor && ./processor
   ```

### API Endpoints

#### Stream Management

- **Register Stream Configuration**
  - `POST /stream-config`
  - Body: 
    ```json
    {
      "events_supported": ["event1", "event2"],
      "events_endpoint": "https://example.com/events"
    }
    ```

- **Update Stream Status**
  - `PUT /stream-config/{stream_id}`
  - Body:
    ```json
    {
      "status": "enabled|paused|disabled",
      "reason": "Optional reason"
    }
    ```

- **Get Stream Status**
  - `GET /stream-config/{stream_id}`

#### Subject Management

- **Add Subject to Stream**
  - `POST /ssf/subjects:add`
  - Authentication: JWT with subject and stream_id claims

- **Remove Subject from Stream**
  - `POST /ssf/subjects:remove`
  - Authentication: JWT with subject and stream_id claims

#### Event Production

- **Produce Event**
  - `POST /produce-event`
  - Content-Type: application/json

## Data Model

### Stream Configuration
```json
{
  "stream_id": "f67e39a0a4d34d56b3aa1bc4cff0069f",
  "events_supported": ["event1", "event2"],
  "events_endpoint": "https://example.com/events",
  "status": "enabled",
  "subjects": [
    {
      "format": "email",
      "email": "user@example.com"
    }
  ],
  "reason": "Optional reason for status"
}
```

### Subject
```json
{
  "format": "email|phone_number",
  "email": "user@example.com",
  "phone_number": "+1234567890"
}
```

### Stream Updated Event
```json
{
  "event_type": "https://schemas.openid.net/secevent/ssf/event-type/stream-updated",
  "sub_id": "f67e39a0a4d34d56b3aa1bc4cff0069f",
  "status": "enabled|paused|disabled",
  "reason": "Optional explanation"
}
```

## Security

The SSF implementation includes multiple layers of security:

1. **JWT Authentication**: For subject management operations
2. **mTLS Authentication**: For secure service-to-service communication
3. **OAuth 2.0**: For secure access to receiver endpoints
4. **SET (Security Event Tokens)**: JWT-based format for secure event transmission

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| MONGODB_URI | MongoDB connection string | mongodb://localhost:27017 |
| CERT_FILE | Client certificate for mTLS | certs/cert.crt |
| KEY_FILE | Private key for mTLS | certs/cert.key |
| CA_FILE | CA certificate | certs/ca.crt |
| WELL_KNOWN_URL | OpenID configuration URL | - |
| CLIENT_ID | OAuth client ID | - |

## Testing

Run the test script to verify MongoDB interaction:

```
python test.py
```

This script demonstrates:
- Connection setup
- Stream configuration creation
- Status updates
- Data retrieval verification

## Standards Compliance

This implementation adheres to the OpenID Shared Signals Framework specification:

- Security Event Token (SET) format compliance
- Subject identifier handling
- Push-based token delivery
- Transmitter configuration discovery

## Future Enhancements

1. Enhanced validation for subject formats
2. Improved error handling with detailed response codes
3. Support for additional subject identifier formats
4. Implementation of polling-based SET token delivery
5. Multi-tenant support with isolated stream configurations

## License

[Specify your license here]

## Contributing

[Instructions for contributors]
