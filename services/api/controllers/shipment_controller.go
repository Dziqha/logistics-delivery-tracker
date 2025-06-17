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

// Event types untuk kafka
const (
	EventTypeCreated      = "created"
	EventTypeStatusUpdate = "status_updated"
	EventTypeInfoUpdate   = "info_updated"
)

// ShipmentEvent structure untuk kafka messages
type ShipmentKafkaEvent struct {
	EventType     string  `json:"event_type"`
	TrackingCode  string  `json:"tracking_code"`
	SenderName    string  `json:"sender_name,omitempty"`
	ReceiverName  string  `json:"receiver_name,omitempty"`
	OriginAddress string  `json:"origin_address,omitempty"`
	DestAddress   string  `json:"dest_address,omitempty"`
	Priority      string  `json:"priority,omitempty"`
	Weight        float64 `json:"weight,omitempty"`
	Dimensions    string  `json:"dimensions,omitempty"`
	OldStatus     string  `json:"old_status,omitempty"`
	NewStatus     string  `json:"new_status"`
}

func (s *ShipmentController) CreateShipment(c *fiber.Ctx) error {
	var req models.ShipmentCreate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Invalid request body"))
	}

	// Generate tracking code
	trackingCode := fmt.Sprintf("TRK-%s", uuid.New().String()[:8])
	
	// Set default priority
	if req.Priority == "" {
		req.Priority = "normal"
	}

	// Buat event untuk kafka
	event := ShipmentKafkaEvent{
		EventType:     EventTypeCreated,
		TrackingCode:  trackingCode,
		SenderName:    "budi", // hardcoded for now
		ReceiverName:  req.ReceiverName,
		OriginAddress: req.OriginAddress,
		DestAddress:   req.DestAddress,
		Priority:      req.Priority,
		Weight:        req.Weight,
		Dimensions:    req.Dimensions,
		NewStatus:     "created",
	}

	// Kirim ke Kafka
	if err := s.sendToKafka(event, trackingCode); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to create shipment"))
	}

	return c.Status(fiber.StatusCreated).JSON(res.SuccessResponse(fiber.StatusCreated, "Shipment creation request sent successfully", map[string]interface{}{
		"tracking_code": trackingCode,
		"status":        "processing",
		"message":       "Shipment is being created, you will receive notification shortly",
	}))
}

func (s *ShipmentController) UpdateShipmentStatus(c *fiber.Ctx) error {
	var req models.ShipmentUpdate
	trackingCode := c.Params("tracking_code")

	// Parse dan validasi request
	var raw map[string]any
	if err := c.BodyParser(&raw); err != nil {
		return c.JSON(res.BadRequestResponse("Invalid JSON"))
	}

	allowedFields := map[string]bool{
		"status": true,
	}

	for key := range raw {
		if !allowedFields[key] {
			return c.JSON(res.BadRequestResponse("Unknown field"))
		}
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(res.BadRequestResponse("Invalid request body"))
	}

	// Validasi status
	validStatuses := map[string]bool{
		"created":       true,
		"packaged":      true,
		"picked_up":     true,
		"shipped":       true,
		"transit_final": true,
		"delivered":     true,
		"done":          true,
		"cancelled":     true,
		"failed":        true,
	}

	if !validStatuses[req.Status] {
		return c.JSON(res.BadRequestResponse("Invalid status"))
	}

	// Buat event untuk kafka
	event := ShipmentKafkaEvent{
		EventType:    EventTypeStatusUpdate,
		TrackingCode: trackingCode,
		NewStatus:    req.Status,
	}

	// Kirim ke Kafka
	if err := s.sendToKafka(event, trackingCode); err != nil {
		return c.JSON(res.InternalServerErrorResponse("Failed to update shipment status"))
	}

	return c.JSON(res.SuccessResponse(fiber.StatusOK, "Status update request sent successfully", map[string]interface{}{
		"tracking_code": trackingCode,
		"new_status":    req.Status,
		"message":       "Status is being updated, you will receive notification shortly",
	}))
}

// Helper function untuk kirim ke Kafka
func (s *ShipmentController) sendToKafka(event ShipmentKafkaEvent, trackingCode string) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
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
		return fmt.Errorf("failed to send to kafka: %v", err)
	}

	producer.Flush(5000)
	return nil
}

