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

    # Replace http://localhost:8080 with the actual server address and endpoint
curl -X POST http://localhost:8080/ssf/subjects:add \
-H "Content-Type: application/jwt" \
--data-binary "$(jwt encode -S 'your-signing-secret' --alg HS256 <<EOF
{
  "event_type": "https://schemas.openid.net/secevent/ssf/event-type/subject-added",
  "stream_id": "stream-1726832521638745000",
  "subject": {
    "format": "email",
    "email": "example.user@example.com"
  },
  "verified": true,
  "iat": $(date +%s)
}
EOF
)"

curl -X POST http://localhost:8080/ssf/subjects:add \
-H "Content-Type: application/jwt" \
--data-binary "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJldmVudF90eXBlIjoiaHR0cHM6Ly9zY2hlbWFzLm9wZW5pZC5uZXQvc2VjZXZlbnQvc3NmL2V2ZW50LXR5cGUvc3ViamVjdC1hZGRlZCIsInN0cmVhbV9pZCI6InN0cmVhbS0xNzI2ODMyNTIxNjM4NzQ1MDAwIiwic3ViamVjdCI6eyJmb3JtYXQiOiJlbWFpbCIsImVtYWlsIjoiZXhhbXBsZS51c2VyQGV4YW1wbGUuY29tIn0sInZlcmlmaWVkIjp0cnVlLCJpYXQiOjE1MTYyMzkwMjJ9.PGihFmz-jUZxO_KtUN-Ni-4OOfNBAHPQEgZID2MaD28"