package controllers

import (
	"encoding/json"
	"fmt"

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
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Invalid request body"))
	}

	if req.ShipmentID == 0 || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("ShipmentID and Name are required"))
	}

	err := configs.DatabaseConnection().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", req.ShipmentID).First(&shipment).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(res.NotFoundResponse("Shipment with given ID not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to fetch shipment from database"))
	}

	validStatus := map[string]bool{
		"shipped":   true,
		"delivered": true,
	}

	if !validStatus[shipment.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse(fmt.Sprintf(
			"Shipment found, but status '%s' is not allowed for this operation", shipment.Status)))
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to process update location"))
	}

	producer := configs.KafkaProducer()
	defer configs.CloseKafkaProducer(producer)

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
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to send update location"))
	}
	producer.Flush(5000)

	return c.Status(fiber.StatusOK).JSON(res.SuccessResponse(fiber.StatusOK, "Location update request sent successfully", map[string]interface{}{
		"shipment_id":    req.ShipmentID,
		"tracking_code":  shipment.TrackingCode,
		"location":       req.Name,
		"message":        "Location is being updated, you will receive notification shortly",
	}))
}