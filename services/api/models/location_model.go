package models

import (
	"log"
	"time"

	"gorm.io/gorm"
)

type LocationUpdate struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"not null"`
	Latitude    float64   `json:"latitude" gorm:"not null"`
	Longitude   float64   `json:"longitude" gorm:"not null"`
	ShipmentID  int64     `json:"shipment_id" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:'in_transit'"` // picked_up, in_transit, out_for_delivery, delivered
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Shipment Shipment `json:"shipment,omitempty" gorm:"foreignKey:ShipmentID;references:ID"`
}

type LocationCreate struct {
	Name       string  `json:"name" validate:"required"`
	Latitude   float64 `json:"latitude" validate:"required"`
	Longitude  float64 `json:"longitude" validate:"required"`
	ShipmentID int64   `json:"shipment_id" validate:"required"`
	Status     string  `json:"status" validate:"required,oneof=picked_up in_transit out_for_delivery delivered"`
	Notes      string  `json:"notes"`
}

func (l *LocationUpdate) TableName() string {
	return "location_updates"
}

func LocationMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(&LocationUpdate{})
	if err != nil {
		log.Printf("Error migrating LocationUpdate model: %v", err)
		return err
	}
	log.Println("✅ LocationUpdate model migrated successfully")
	return nil
}