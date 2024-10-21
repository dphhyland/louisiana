package main

import "github.com/confluentinc/confluent-kafka-go/kafka"

// KafkaProducer is an interface for producing messages to Kafka
type KafkaProducer interface {
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
}
