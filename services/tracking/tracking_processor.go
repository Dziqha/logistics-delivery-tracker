package tracking

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"gorm.io/gorm"
)

func ProcessTracking() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ Tracking processor recovered from panic: %v", r)
		}
	}()

	consumer := configs.KafkaConsumer()
	defer configs.CloseKafkaConsumer(consumer)
	
	db := configs.DatabaseConnection()
	producer := configs.KafkaProducer()
	defer configs.CloseKafkaProducer(producer)

	log.Println("📦 Shipment processor started...")

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("❌ Kafka read error: %v\n", err)
			continue
		}

		// Only process shipment topic
		if *msg.TopicPartition.Topic != "shipment" {
			continue
		}

		var shipment models.Shipment
		err = json.Unmarshal(msg.Value, &shipment)
		if err != nil {
			log.Printf("❌ Error parsing shipment JSON: %v\nPayload: %s\n", err, msg.Value)
			continue
		}

		log.Printf("📦 Processing shipment: %+v\n", shipment)

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&shipment).Error; err != nil {
				return err
			}

			// Send welcome notification
			notification := models.NotificationCreate{
				ShipmentID: shipment.ID,
				Type:       "status_update",
				Title:      "Shipment Created",
				Message:    fmt.Sprintf("Your shipment %s has been created and is being processed", shipment.TrackingCode),
			}

			notificationPayload, _ := json.Marshal(notification)
			topic := "notification"
			producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: kafka.PartitionAny,
				},
				Key:   []byte(shipment.SenderName),
				Value: notificationPayload,
			}, nil)

			return nil
		})

		if err != nil {
			log.Printf("❌ DB insert error: %v\n", err)
		} else {
			log.Printf("✅ Shipment %s from %s to %s inserted successfully\n", shipment.TrackingCode, shipment.SenderName, shipment.ReceiverName)
		}

		// Commit the message
		consumer.CommitMessage(msg)
	}
}