package configs

import (
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func KafkaProducer() *kafka.Producer {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"acks":             "all",
		"retries":          3,
		"batch.size":       16384,
		"linger.ms":        1,
		"queue.buffering.max.kbytes":   2 * 1024 * 1024, 
	})
	
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	log.Println("✅ Kafka producer connected successfully")
	return producer
}

func KafkaConsumer() *kafka.Consumer {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "delivery-tracker"
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
		"enable.auto.commit": false,
	})
	
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	topics := []string{"shipment", "location-update", "notification"}
	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %v", err)
	}

	log.Println("✅ Kafka consumer connected successfully")
	return consumer
}

func CloseKafkaProducer(producer *kafka.Producer) {
	if producer != nil {
		producer.Flush(15 * 1000)
		producer.Close()
	}
}

func CloseKafkaConsumer(consumer *kafka.Consumer) {
	if consumer != nil {
		consumer.Close()
	}
}