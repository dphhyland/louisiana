module github.com/trustframeworks/ssf/processor

go 1.22.5

replace github.com/trustframeworks/oauthclient => /users/davidhyland/source/louisiana/oauthclient

require (
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/trustframeworks/oauthclient v0.0.0
)

require github.com/joho/godotenv v1.5.1 // indirect
