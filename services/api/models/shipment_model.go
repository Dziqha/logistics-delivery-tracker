package models

import (
	"log"
	"time"

	"gorm.io/gorm"
)

type Shipment struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TrackingCode  string    `json:"tracking_code" gorm:"unique;not null"`
	SenderName    string    `json:"sender_name" gorm:"not null"`
	ReceiverName  string    `json:"receiver_name" gorm:"not null"`
	OriginAddress string    `json:"origin_address" gorm:"not null"`
	DestAddress   string    `json:"dest_address" gorm:"not null"`
	Status        string    `json:"status" gorm:"default:'created'"` // created, picked_up, in_transit, out_for_delivery, delivered, cancelled
	Priority      string    `json:"priority" gorm:"default:'normal'"` // low, normal, high, urgent
	Weight        float64   `json:"weight"`
	Dimensions    string    `json:"dimensions"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Locations     []LocationUpdate `json:"locations,omitempty" gorm:"foreignKey:ShipmentID;references:ID"`
	Notifications []Notification   `json:"notifications,omitempty" gorm:"foreignKey:ShipmentID;references:ID"`
}

type ShipmentCreate struct {
	ReceiverName  string  `json:"receiver_name" validate:"required"`
	OriginAddress string  `json:"origin_address" validate:"required"`
	DestAddress   string  `json:"dest_address" validate:"required"`
	Priority      string  `json:"priority" validate:"oneof=low normal high urgent"`
	Weight        float64 `json:"weight"`
	Dimensions    string  `json:"dimensions"`
}

type ShipmentUpdate struct {
	Status string `json:"status" validate:"required,oneof=created picked_up in_transit out_for_delivery delivered cancelled"`
}

func (s *Shipment) TableName() string {
	return "shipments"
}

func ShipmentMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(&Shipment{})
	if err != nil {
		log.Printf("Error migrating Shipment model: %v", err)
		return err
	}
	log.Println("✅ Shipment model migrated successfully")
	return nil
}