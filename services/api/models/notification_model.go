package models

import (
	"log"
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ShipmentID  int64     `json:"shipment_id" gorm:"not null"`
	Type        string    `json:"type" gorm:"not null"` // status_update, location_update, delivery_alert
	Title       string    `json:"title" gorm:"not null"`
	Message     string    `json:"message" gorm:"not null"`
	IsRead      bool      `json:"is_read" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Shipment Shipment `json:"shipment,omitempty" gorm:"foreignKey:ShipmentID;references:ID"`
}

type NotificationCreate struct {
	ShipmentID int64  `json:"shipment_id" validate:"required"`
	Type       string `json:"type" validate:"required,oneof=status_update location_update delivery_alert"`
	Title      string `json:"title" validate:"required"`
	Message    string `json:"message" validate:"required"`
}

type NotificationUpdate struct {
	IsRead bool `json:"is_read"`
}

func (n *Notification) TableName() string {
	return "notifications"
}

func NotificationMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(&Notification{})
	if err != nil {
		log.Printf("Error migrating Notification model: %v", err)
		return err
	}
	log.Println("✅ Notification model migrated successfully")
	return nil
}