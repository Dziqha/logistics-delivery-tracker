package location

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"gorm.io/gorm"
)

func ProcessLocationUpdates() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ Location processor recovered from panic: %v", r)
		}
	}()

	consumer := configs.KafkaConsumer()
	defer configs.CloseKafkaConsumer(consumer)

	db := configs.DatabaseConnection()
	producer := configs.KafkaProducer()
	defer configs.CloseKafkaProducer(producer)

	log.Println("🚀 Location processor started...")

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("❌ Kafka read error: %v\n", err)
			continue
		}

		if *msg.TopicPartition.Topic != "location-update" {
			continue
		}

		var locationReq models.LocationCreate
		err = json.Unmarshal(msg.Value, &locationReq)
		if err != nil {
			log.Printf("❌ Error parsing location JSON: %v\nPayload: %s\n", err, msg.Value)
			continue
		}

		log.Printf("📍 Processing location update: %+v\n", locationReq)

		err = db.Transaction(func(tx *gorm.DB) error {
			// Validasi: shipment harus ada
			var shipment models.Shipment
			if err := tx.First(&shipment, locationReq.ShipmentID).Error; err != nil {
				return fmt.Errorf("❌ Invalid shipment ID %d: %v", locationReq.ShipmentID, err)
			}

			// Insert lokasi hanya jika shipment valid
			location := models.LocationUpdate{
				Name:       locationReq.Name,
				Latitude:   locationReq.Latitude,
				Longitude:  locationReq.Longitude,
				ShipmentID: locationReq.ShipmentID,
				Status:     locationReq.Status,
				Notes:      locationReq.Notes,
			}

			if err := tx.Create(&location).Error; err != nil {
				return err
			}

			// Update status shipment jika berubah
			if locationReq.Status != "" && shipment.Status != locationReq.Status {
				shipment.Status = locationReq.Status
				if err := tx.Save(&shipment).Error; err != nil {
					return err
				}

				notification := models.NotificationCreate{
					ShipmentID: shipment.ID,
					Type:       "location_update",
					Title:      "Location Updated",
					Message:    fmt.Sprintf("Your shipment %s is now at %s with status: %s", shipment.TrackingCode, locationReq.Name, locationReq.Status),
				}

				notificationPayload, _ := json.Marshal(notification)
				topic := "notification"
				producer.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     &topic,
						Partition: kafka.PartitionAny,
					},
					Key:   []byte(shipment.TrackingCode),
					Value: notificationPayload,
				}, nil)
			}

			return nil
		})

		if err != nil {
			log.Printf("❌ DB transaction error: %v\n", err)
			continue // ❌ Jangan commit kalau gagal
		}

		log.Printf("✅ Location update processed for shipment ID: %d at %s\n", locationReq.ShipmentID, locationReq.Name)
		consumer.CommitMessage(msg)
	}
}
