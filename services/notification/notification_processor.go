package notification

import (
	_"encoding/json"
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
		}
	}()

	db := configs.DatabaseConnection()
	log.Println("📨 Notification processor started...")

	err := configs.ConsumeNotifications(func(n models.NotificationCreate) {
		processNotification(db, n)
	})

	if err != nil {
		log.Fatalf("❌ Error starting notification consumer: %v", err)
	}
}

func processNotification(db *gorm.DB, n models.NotificationCreate) {
	// Validasi data notifikasi
	if n.ShipmentID == 0 {
		log.Printf("❌ Invalid notification: missing ShipmentID")
		return
	}

	if n.Title == "" || n.Message == "" {
		log.Printf("❌ Invalid notification: missing Title or Message")
		return
	}

	// Cek apakah shipment exists
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
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		// Cek duplikasi notifikasi (opsional: untuk mencegah spam)
		var existingCount int64
		tx.Model(&models.Notification{}).
			Where("shipment_id = ? AND type = ? AND title = ? AND created_at > ?", 
				n.ShipmentID, n.Type, n.Title, time.Now().Add(-5*time.Minute)).
			Count(&existingCount)

		if existingCount > 0 {
			log.Printf("⚠️ Duplicate notification detected, skipping: %s", n.Title)
			return nil
		}

		// Simpan notifikasi
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
