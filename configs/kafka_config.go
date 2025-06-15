package configs

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func KafkaProducer() *kafka.Producer{
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	log.Println("Kafka producer connected successfully")
	return producer
}


func KafkaConsumer() *kafka.Consumer{
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "delevery-tracker",
		"auto.offset.reset": "earliest",
		"debug":             "all",
	})
	
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	log.Println("Kafka consumer connected successfully")
	return consumer
}