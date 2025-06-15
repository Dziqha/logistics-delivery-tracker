package controllers

import (
	"encoding/json"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/res"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gofiber/fiber/v2"
)

type ShipmentController struct{}

func NewShipmentController() *ShipmentController {
	return &ShipmentController{}
}

func (s *ShipmentController) CreateShipment(c *fiber.Ctx) error {
	var req models.ShipmentCreate
	if err := c.BodyParser(&req); err != nil {
		res.BadRequestResponse("Invalid request body")
	}

	if req.Status != "created" {
		res.BadRequestResponse("Invalid status")
	}
	
	payload , err := json.Marshal(req)
	if err != nil {
		res.InternalServerErrorResponse("Failed to create shipment")
	}
	topic := "shipment"
	configs.KafkaProducer().Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:    &topic,
			Partition: kafka.PartitionAny,
			Offset: kafka.OffsetBeginning,
		},
		Key:   []byte(req.SenderName),
		Value: payload,
	}, nil)
	configs.KafkaProducer().Flush(5000)
	return c.JSON(res.SuccessResponse(fiber.StatusCreated, "Shipment send to kafka successfully", req))

}
		