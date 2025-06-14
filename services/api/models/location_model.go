package models

import (
	"log"

	"gorm.io/gorm"
)

type LocationUpdate struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Shipment_id int64   `json:"shipment_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`

	Shipment Shipment `json:"shipment" gorm:"foreignKey:ShipmentID;references:ID"`
}

func (l *LocationUpdate) TableName() string {
	return "locationsUpdates"
}

func LocationMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(&LocationUpdate{})
	if err != nil {
		log.Printf("Error migrating Location model: %v", err)
	}
	return nil
}