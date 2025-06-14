package models

import (
	"log"

	"gorm.io/gorm"
)

type Shipment struct {
	ID            string `json:"id"`
	SenderName    string `json:"sender_name" `
	ReceiverName  string `json:"receiver_name"`
	OriginAddress string `json:"origin_address"`
	DestAddress   string `json:"dest_address" `
	Status        string `json:"status"`
	UpdatedAt     string `json:"updated_at"`

	Locations []LocationUpdate `json:"locationsUpdates,omitempty" gorm:"foreignKey:ShipmentID;references:ID"`
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