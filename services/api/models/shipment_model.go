package models

import (
	"log"

	"gorm.io/gorm"
)

type Shipment struct {
	ID            int64  `json:"id"`
	SenderName    string `json:"sender_name" `
	ReceiverName  string `json:"receiver_name"`
	OriginAddress string `json:"origin_address"`
	DestAddress   string `json:"dest_address" `
	Status        string `json:"status"`
	UpdatedAt     string `json:"updated_at"`

	Locations []LocationUpdate `json:"locationsUpdates,omitempty" gorm:"foreignKey:Shipment_id;references:ID"`
}

type ShipmentCreate struct {
	SenderName    string `json:"sender_name" `
	ReceiverName  string `json:"receiver_name"`
	OriginAddress string `json:"origin_address"`
	DestAddress   string `json:"dest_address" `
	Status        string `json:"status"`
}

func (s *Shipment) TableName() string {
	return "shipments"
}

func ShipmentMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(&Shipment{})
	if err != nil {
		log.Printf("Error migrating Shipment model: %v", err)
	}
	return nil
}