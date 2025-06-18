package location

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"gorm.io/gorm"
)

func ProcessLocationUpdates() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ Location processor recovered from panic: %v", r)
			time.Sleep(5 * time.Second)
			go ProcessLocationUpdates()
		}
	}()

	log.Println("🔧 Initializing Kafka location consumer...")
	consumer := configs.KafkaLocationConsumer()
	defer configs.CloseKafkaConsumer(consumer)

	db := configs.DatabaseConnection()
	producer := configs.KafkaProducer()
	defer configs.CloseKafkaProducer(producer)

	log.Println("🚀 Location processor started...")
	log.Println("🎯 Listening for location updates on 'location-update' topic...")

	messageCount := 0

	for {
		log.Printf("🔍 Attempting to read location message... (count: %d)", messageCount)

		msg, err := consumer.ReadMessage(5 * time.Second)
		if err != nil {
			if err.Error() == "Local: Timed out" {
				log.Printf("⏱️ No location messages in 5s, continuing to listen...")
				continue
			}
			log.Printf("❌ Kafka read error: %v", err)
			continue
		}

		messageCount++
		log.Printf("📨 Received location message #%d from topic: %s", messageCount, *msg.TopicPartition.Topic)

		if *msg.TopicPartition.Topic != "location-update" {
			log.Printf("⚠️ Skipping message from wrong topic: %s", *msg.TopicPartition.Topic)
			if _, err := consumer.CommitMessage(msg); err != nil {
				log.Printf("❌ Failed to commit skipped message: %v", err)
			}
			continue
		}

		log.Printf("📄 Raw location message: %s", string(msg.Value))

		var locationReq models.LocationCreate
		err = json.Unmarshal(msg.Value, &locationReq)
		if err != nil {
			log.Printf("❌ Error parsing location JSON: %v\nPayload: %s", err, msg.Value)
			if _, err := consumer.CommitMessage(msg); err != nil {
				log.Printf("❌ Failed to commit error message: %v", err)
			}
			continue
		}

		log.Printf("📍 Processing location update: %+v", locationReq)

		var processedSuccessfully bool

		err = db.Transaction(func(tx *gorm.DB) error {
			var shipment models.Shipment
			if err := tx.First(&shipment, locationReq.ShipmentID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Printf("❌ Shipment not found: ID %d", locationReq.ShipmentID)
					return nil 
				}
				return fmt.Errorf("database error while finding shipment: %v", err)
			}

			log.Printf("📋 Found shipment: ID=%d, TrackingCode=%s, CurrentStatus=%s", 
				shipment.ID, shipment.TrackingCode, shipment.Status)

			location := models.LocationUpdate{
				Name:       locationReq.Name,
				Latitude:   locationReq.Latitude,
				Longitude:  locationReq.Longitude,
				ShipmentID: locationReq.ShipmentID,
				// Status:     locationReq.Status,
				Notes:      locationReq.Notes,
			}

			if err := tx.Create(&location).Error; err != nil {
				return fmt.Errorf("failed to create location update: %v", err)
			}

			log.Printf("✅ Location update saved: ID=%d, Name=%s", location.ID, location.Name)

			notification := models.NotificationCreate{
				ShipmentID: shipment.ID,
				Type:       "location_update",
				Title:      "Location Updated",
				Message:    fmt.Sprintf("Your shipment %s is now at %s with status: %s", 
					shipment.TrackingCode, locationReq.Name, shipment.Status),
			}

			notificationPayload, err := json.Marshal(notification)
			if err != nil {
				log.Printf("❌ Failed to marshal notification: %v", err)
				return nil 
			}

			topic := "notification"
			err = producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: kafka.PartitionAny,
				},
				Key:   []byte(shipment.TrackingCode),
				Value: notificationPayload,
			}, nil)

			if err != nil {
				log.Printf("❌ Failed to send notification to Kafka: %v", err)
				return nil 
			}

			log.Printf("📤 Notification sent to Kafka for shipment %s", shipment.TrackingCode)
			// if locationReq.Status != "" && shipment.Status != locationReq.Status {
			// 	oldStatus := shipment.Status
			// 	shipment.Status = locationReq.Status
			// 	if err := tx.Save(&shipment).Error; err != nil {
			// 		return fmt.Errorf("failed to update shipment status: %v", err)
			// 	}

			// 	log.Printf("📦 Shipment status updated: %s -> %s", oldStatus, locationReq.Status)

			// 	notification := models.NotificationCreate{
			// 		ShipmentID: shipment.ID,
			// 		Type:       "location_update",
			// 		Title:      "Location Updated",
			// 		Message:    fmt.Sprintf("Your shipment %s is now at %s with status: %s", 
			// 			shipment.TrackingCode, locationReq.Name, locationReq.Status),
			// 	}

			// 	notificationPayload, err := json.Marshal(notification)
			// 	if err != nil {
			// 		log.Printf("❌ Failed to marshal notification: %v", err)
			// 		return nil 
			// 	}

			// 	topic := "notification"
			// 	err = producer.Produce(&kafka.Message{
			// 		TopicPartition: kafka.TopicPartition{
			// 			Topic:     &topic,
			// 			Partition: kafka.PartitionAny,
			// 		},
			// 		Key:   []byte(shipment.TrackingCode),
			// 		Value: notificationPayload,
			// 	}, nil)

			// 	if err != nil {
			// 		log.Printf("❌ Failed to send notification to Kafka: %v", err)
			// 		return nil 
			// 	}

			// 	log.Printf("📤 Notification sent to Kafka for shipment %s", shipment.TrackingCode)
			// } else {
			// 	log.Printf("ℹ️ No status change needed (current: %s, requested: %s)", 
			// 		shipment.Status, locationReq.Status)
			// }

			processedSuccessfully = true
			return nil
		})

		if err != nil {
			log.Printf("❌ DB transaction error: %v", err)
			continue
		}

		if processedSuccessfully {
			if _, err := consumer.CommitMessage(msg); err != nil {
				log.Printf("❌ Failed to commit message: %v", err)
			} else {
				log.Printf("✅ Location update processed and committed for shipment ID: %d at %s", 
					locationReq.ShipmentID, locationReq.Name)
			}
		}

		log.Println("─────────────────────────────────────")
	}
}