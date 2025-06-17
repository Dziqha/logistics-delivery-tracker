package configs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

var conn *amqp.Connection
var channel *amqp.Channel

const queueName = "notifications"

func InitRabbitMQ() {
	var err error
	conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %v", err)
	}

	channel, err = conn.Channel()
	if err != nil {
		log.Fatalf("❌ Failed to open a channel: %v", err)
	}

	_, err = channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("❌ Failed to declare queue: %v", err)
	}

	log.Println("✅ Connected to RabbitMQ and queue declared")
}

func SendNotification(notification models.NotificationCreate) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("❌ Failed to marshal notification: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return fmt.Errorf("❌ Failed to publish message: %w", err)
	}

	return nil
}

func ConsumeNotifications(handler func(models.NotificationCreate)) error {
	msgs, err := channel.Consume(
		queueName,
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("❌ Failed to register consumer: %w", err)
	}

	log.Println("🔔 [RabbitMQ] Waiting for notifications...")

	go func() {
		for msg := range msgs {
			var n models.NotificationCreate
			if err := json.Unmarshal(msg.Body, &n); err != nil {
				log.Printf("❌ Failed to parse message: %v", err)
				continue
			}
			handler(n)
		}
	}()

	return nil
}

func CloseRabbitMQ() {
	if channel != nil {
		channel.Close()
	}
	if conn != nil {
		conn.Close()
	}
}
