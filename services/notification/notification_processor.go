package notification

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"gorm.io/gorm"
)

func NotificationProcessor() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ Notification processor recovered from panic: %v", r)
			time.Sleep(5 * time.Second)
			go NotificationProcessor()
		}
	}()

	db := configs.DatabaseConnection()
	log.Println("📨 Notification processor started...")

	err := configs.ConsumeNotifications(func(n models.NotificationCreate) {
		processNotification(db, n)
	})

	if err != nil {
		log.Printf("❌ Error starting notification consumer: %v", err)
		time.Sleep(10 * time.Second)
		go NotificationProcessor()
	}

	// Also start Kafka notification consumer for notifications from location updates
	go processKafkaNotifications(db)
}

func processKafkaNotifications(db *gorm.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ Kafka notification processor recovered from panic: %v", r)
			time.Sleep(5 * time.Second)
			go processKafkaNotifications(db)
		}
	}()

	log.Println("🔧 Initializing Kafka notification consumer...")
	consumer := configs.KafkaNotificationConsumer()
	defer configs.CloseKafkaConsumer(consumer)

	log.Println("📨 Kafka notification processor started...")
	log.Println("🎯 Listening for notifications on 'notification' topic...")

	messageCount := 0

	for {
		log.Printf("🔍 Attempting to read notification message... (count: %d)", messageCount)

		msg, err := consumer.ReadMessage(5 * time.Second)
		if err != nil {
			if err.Error() == "Local: Timed out" {
				log.Printf("⏱️ No notification messages in 5s, continuing to listen...")
				continue
			}
			log.Printf("❌ Kafka read error: %v", err)
			continue
		}

		messageCount++
		log.Printf("📨 Received notification message #%d from topic: %s", messageCount, *msg.TopicPartition.Topic)

		if *msg.TopicPartition.Topic != "notification" {
			log.Printf("⚠️ Skipping message from wrong topic: %s", *msg.TopicPartition.Topic)
			if _, err := consumer.CommitMessage(msg); err != nil {
				log.Printf("❌ Failed to commit skipped message: %v", err)
			}
			continue
		}

		log.Printf("📄 Raw notification message: %s", string(msg.Value))

		var notification models.NotificationCreate
		if err := json.Unmarshal(msg.Value, &notification); err != nil {
			log.Printf("❌ Error parsing notification JSON: %v\nPayload: %s", err, msg.Value)
			if _, err := consumer.CommitMessage(msg); err != nil {
				log.Printf("❌ Failed to commit error message: %v", err)
			}
			continue
		}

		log.Printf("🔔 Processing Kafka notification: %+v", notification)

		processNotification(db, notification)

		if _, err := consumer.CommitMessage(msg); err != nil {
			log.Printf("❌ Failed to commit notification message: %v", err)
		} else {
			log.Printf("✅ Kafka notification message committed successfully")
		}

		log.Println("─────────────────────────────────────")
	}
}

func processNotification(db *gorm.DB, n models.NotificationCreate) {
	if n.ShipmentID == 0 {
		log.Printf("❌ Invalid notification: missing ShipmentID")
		return
	}

	if n.Title == "" || n.Message == "" {
		log.Printf("❌ Invalid notification: missing Title or Message")
		return
	}

	var shipment models.Shipment
	if err := db.First(&shipment, n.ShipmentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("❌ Shipment not found for notification: ShipmentID %d", n.ShipmentID)
			return
		}
		log.Printf("❌ Error checking shipment: %v", err)
		return
	}

	notification := models.Notification{
		ShipmentID: n.ShipmentID,
		Type:       n.Type,
		Title:      n.Title,
		Message:    n.Message,
		IsRead:     true,
		CreatedAt:  time.Now(),
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var existingCount int64
		tx.Model(&models.Notification{}).
			Where("shipment_id = ? AND type = ? AND title = ? AND created_at > ?", 
				n.ShipmentID, n.Type, n.Title, time.Now().Add(-5*time.Minute)).
			Count(&existingCount)

		if existingCount > 0 {
			log.Printf("⚠️ Duplicate notification detected, skipping: %s", n.Title)
			return nil
		}

		if err := tx.Create(&notification).Error; err != nil {
			return err
		}

		log.Printf("✅ Notification saved: ID=%d, ShipmentID=%d, Title=%s", 
			notification.ID, notification.ShipmentID, notification.Title)

		return nil
	})

	if err != nil {
		log.Printf("❌ Failed to save notification: %v", err)
		return
	}
}