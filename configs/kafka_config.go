package configs

import (
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConfig struct {
	BootstrapServers string
	SecurityProtocol string
	GroupID         string
	AutoOffsetReset string
}

func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		BootstrapServers: os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		SecurityProtocol: os.Getenv("KAFKA_SECURITY_PROTOCOL"),
		GroupID:         os.Getenv("KAFKA_GROUP_ID"),
		AutoOffsetReset: os.Getenv("KAFKA_AUTO_OFFSET_RESET"),
	}
}

func (kc *KafkaConfig) GetConfigMap() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers": kc.BootstrapServers,
		"security.protocol": kc.SecurityProtocol,
		"group.id":          kc.GroupID,
		"auto.offset.reset": kc.AutoOffsetReset,
	}
}

func NewProducer() (*kafka.Producer, error) {
	config := NewKafkaConfig()
	
	// Additional producer-specific configuration
	producerConfig := config.GetConfigMap()
	producerConfig.SetKey("go.delivery.reports", true)
	producerConfig.SetKey("queue.buffering.max.messages", 100000)
	producerConfig.SetKey("message.send.max.retries", 3)

	producer, err := kafka.NewProducer(producerConfig)
	if err != nil {
		return nil, err
	}

	// Delivery report handler for produced messages
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition.Error)
				} else {
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	return producer, nil
}

func NewConsumer(topics []string) (*kafka.Consumer, error) {
	config := NewKafkaConfig()
	
	// Additional consumer-specific configuration
	consumerConfig := config.GetConfigMap()
	consumerConfig.SetKey("enable.auto.commit", false)
	consumerConfig.SetKey("max.poll.interval.ms", 300000)

	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		return nil, err
	}

	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}
