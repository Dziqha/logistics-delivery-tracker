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
		"queue.buffering.max.kbytes": 2 * 1024 * 1024,
	})

	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	log.Println("✅ Kafka producer connected successfully")
	return producer
}

// Consumer khusus untuk shipment tracking
func KafkaShipmentConsumer() *kafka.Consumer {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	groupID := os.Getenv("KAFKA_SHIPMENT_GROUP_ID")
	if groupID == "" {
		groupID = "shipment-tracker"
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
		"enable.auto.commit": false,
	})

	if err != nil {
		log.Fatalf("Failed to create Kafka shipment consumer: %v", err)
	}

	// Hanya subscribe ke topic shipment
	topics := []string{"shipment"}
	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to shipment topic: %v", err)
	}

	log.Println("✅ Kafka shipment consumer connected successfully")
	return consumer
}

// Consumer khusus untuk location updates
func KafkaLocationConsumer() *kafka.Consumer {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	groupID := os.Getenv("KAFKA_LOCATION_GROUP_ID")
	if groupID == "" {
		groupID = "location-tracker"
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
		"enable.auto.commit": false,
	})

	if err != nil {
		log.Fatalf("Failed to create Kafka location consumer: %v", err)
	}

	// Hanya subscribe ke topic location-update
	topics := []string{"location-update"}
	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to location-update topic: %v", err)
	}

	log.Println("✅ Kafka location consumer connected successfully")
	return consumer
}

// Consumer khusus untuk notifications
func KafkaNotificationConsumer() *kafka.Consumer {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	groupID := os.Getenv("KAFKA_NOTIFICATION_GROUP_ID")
	if groupID == "" {
		groupID = "notification-processor"
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
		"enable.auto.commit": false,
	})

	if err != nil {
		log.Fatalf("Failed to create Kafka notification consumer: %v", err)
	}

	// Hanya subscribe ke topic notification
	topics := []string{"notification"}
	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to notification topic: %v", err)
	}

	log.Println("✅ Kafka notification consumer connected successfully")
	return consumer
}

// Backward compatibility - gunakan shipment consumer sebagai default
func KafkaConsumer() *kafka.Consumer {
	return KafkaShipmentConsumer()
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