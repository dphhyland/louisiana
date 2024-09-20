curl -X POST http://localhost:8080/stream-config \
-H "Content-Type: application/json" \
-d '{
      "events_supported": ["event1", "event2"],
      "events_endpoint": "http://example.com/endpoint"
    }'


    curl -X PUT http://localhost:8080/stream-config/stream-1726745430336780000 \
-H "Content-Type: application/json" \
-d '{
      "status": "paused",
      "reason": "Maintenance"
    }'


    curl -X POST http://localhost:8080/ssf/subjects:add \
-H "Content-Type: application/json" \
-d '{
      "stream_id": "stream-1726745430336780000",
      "subject": {
        "format": "email",
        "email": "example.user@example.com"
      },
      "verified": true
    }'

    curl -X POST http://localhost:8080/ssf/subjects:remove \
-H "Content-Type: application/json" \
-d '{
      "stream_id": "stream-<id>",
      "subject": {
        "format": "email",
        "email": "example.user@example.com"
      }
    }'

    curl -X GET http://localhost:8080/stream-config/stream-1726745430336780000

    curl -X PUT http://localhost:8080/stream-config/stream-1726745430336780000 \
-H "Content-Type: application/json" \
-d '{
      "status": "enabled",
      "reason": "Re-enabling the stream"
    }'

    curl -X PUT http://localhost:8080/stream-config/stream-1726832521638745000 \
-H "Content-Type: application/json" \
-d '{
      "status": "enabled",
      "reason": "Reactivating the stream"
    }'

    curl -X PUT http://localhost:8080/stream-config/stream-1726832521638745000 \
-H "Content-Type: application/json" \
-d '{
      "status": "invalid-status",
      "reason": "Trying an invalid status"
    }'