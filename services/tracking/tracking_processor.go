package tracking

import (
	"encoding/json"
	"log"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"gorm.io/gorm"
)

func ProcessTracking() {
	defer func() {
        if r := recover(); r != nil {
            log.Printf("⚠️ Recovered from panic: %v", r)
        }
    }()
	consumer := configs.KafkaConsumer()
	db := configs.DatabaseConnection()

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("❌ Kafka read error: %v\n", err)
			continue
		}

		var shipment models.ShipmentCreate
		err = json.Unmarshal(msg.Value, &shipment)
		if err != nil {
			log.Printf("❌ Error parsing JSON: %v\nPayload: %s\n", err, msg.Value)
			continue
		}

		log.Printf("📦 Received shipment: %+v\n", shipment)

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&shipment).Error; err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			log.Printf("❌ DB insert error: %v\n", err)
		} else {
			log.Printf("✅ Shipment from %s to %s inserted.\n", shipment.SenderName, shipment.ReceiverName)
		}
	}
}
