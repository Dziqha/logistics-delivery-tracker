package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/res"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ShipmentController struct{}

func NewShipmentController() *ShipmentController {
	return &ShipmentController{}
}


func (s *ShipmentController) CreateShipment(c *fiber.Ctx) error {
	var req models.ShipmentCreate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Invalid request body"))
	}

	trackingCode := fmt.Sprintf("TRK-%s", uuid.New().String()[:8])
	
	shipment := models.Shipment{
		TrackingCode:  trackingCode,
		SenderName:    req.SenderName,
		ReceiverName:  req.ReceiverName,
		OriginAddress: req.OriginAddress,
		DestAddress:   req.DestAddress,
		Status:        "created",
		Priority:      req.Priority,
		Weight:        req.Weight,
		Dimensions:    req.Dimensions,
	}

	if shipment.Priority == "" {
		shipment.Priority = "normal"
	}

	payload, err := json.Marshal(shipment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to create shipment"))
	}

	producer := configs.KafkaProducer()
	defer configs.CloseKafkaProducer(producer)

	topic := "shipment"
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(trackingCode),
		Value: payload,
	}, nil)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to send shipment to queue"))
	}

	producer.Flush(5000)

	resShipment := res.ShipmentResponseData(shipment)
	
	return c.Status(fiber.StatusCreated).JSON(res.SuccessResponse(fiber.StatusCreated, "Shipment created successfully", map[string]interface{}{
		"tracking_code": trackingCode,
		"shipment":      resShipment,
	}))
}
