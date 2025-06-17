package controllers

import (
	"encoding/json"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/res"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type LocationController struct{}

func NeLocationController() *LocationController {
	return &LocationController{}
}

func (l *LocationController) UpdateLocation(c *fiber.Ctx) error {
	var req models.LocationCreate
	var shipment models.Shipment
	if err := c.BodyParser(&req); err != nil {
		res.BadRequestResponse("Invalid request body")
	}

	err := configs.DatabaseConnection().Transaction(func(tx *gorm.DB) error {
		if result := tx.Where("id = ?", req.ShipmentID).First(&shipment); result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return result.Error
			}
		}
		return nil
	})

	if err != nil {
		res.NotFoundResponse("Shipment not found")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		res.InternalServerErrorResponse("Failed to process update location")
	}

	producer := configs.KafkaProducer()
	defer producer.Close()

	topic := "location-update"
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(shipment.TrackingCode),
		Value: payload,
	}, nil)

	if err != nil {
		res.InternalServerErrorResponse("Failed to send update location")
	}
	producer.Flush(5000)

	return c.JSON(res.SuccessResponse(fiber.StatusOK, "Location updated successfully", req))
}