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
	"gorm.io/gorm"
)

type ShipmentController struct{}

func NewShipmentController() *ShipmentController {
	return &ShipmentController{}
}

const (
	EventTypeCreated      = "created"
	EventTypeStatusUpdate = "status_updated"
	EventTypeInfoUpdate   = "info_updated"
)

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

	userID, ok := c.Locals("username").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(res.UnauthorizedResponse("unauthorized UserID"))
	}
	

	if req.ReceiverName == "" || req.OriginAddress == "" || req.DestAddress == "" {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("ReceiverName, OriginAddress, and DestAddress are required"))
	}

	trackingCode := fmt.Sprintf("TRK-%s", uuid.New().String()[:8])
	
	if req.Priority == "" {
		req.Priority = "normal"
	}

	validPriorities := map[string]bool{
		"low":    true,
		"normal": true,
		"high":   true,
		"urgent": true,
	}

	if !validPriorities[req.Priority] {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Invalid priority. Valid priorities: low, normal, high, urgent"))
	}

	shipment := models.Shipment{
		TrackingCode:  trackingCode,
		SenderName:    userID,
		ReceiverName:  req.ReceiverName,
		OriginAddress: req.OriginAddress,
		DestAddress:   req.DestAddress,
		Status:        "created",
		Priority:      req.Priority,
		Weight:        req.Weight,
		Dimensions:    req.Dimensions,
	}

	err := configs.DatabaseConnection().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&shipment).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to create shipment in database"))
	}

	event := ShipmentKafkaEvent{
		EventType:     EventTypeCreated,
		TrackingCode:  trackingCode,
		SenderName:    shipment.SenderName,
		ReceiverName:  req.ReceiverName,
		OriginAddress: req.OriginAddress,
		DestAddress:   req.DestAddress,
		Priority:      req.Priority,
		Weight:        req.Weight,
		Dimensions:    req.Dimensions,
		NewStatus:     "created",
	}

	if err := s.sendToKafka(event, trackingCode); err != nil {
		configs.DatabaseConnection().Delete(&shipment)
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to send shipment event"))
	}

	return c.Status(fiber.StatusCreated).JSON(res.SuccessResponse(fiber.StatusCreated, "Shipment created successfully", map[string]any{
		"shipment":      shipment,
		"tracking_code": trackingCode,
		"message":       "Shipment created successfully, notification will be sent shortly",
	}))
}

func (s *ShipmentController) UpdateShipmentStatus(c *fiber.Ctx) error {
	var req models.ShipmentUpdate
	trackingCode := c.Params("tracking_code")

	if trackingCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Tracking code is required"))
	}

	var raw map[string]any
	if err := c.BodyParser(&raw); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Invalid JSON"))
	}

	allowedFields := map[string]bool{"status": true}
	for key := range raw {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Only 'status' field is allowed"))
		}
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Invalid request body"))
	}

	validStatuses := map[string]bool{
		"created": true, "packaged": true, "picked_up": true, "shipped": true,
		"transit_final": true, "delivered": true, "done": true, "cancelled": true, "failed": true,
	}

	if !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Invalid status"))
	}

	var shipment models.Shipment
	var oldStatus string

	err := configs.DatabaseConnection().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tracking_code = ?", trackingCode).First(&shipment).Error; err != nil {
			return err
		}
		oldStatus = shipment.Status
		shipment.Status = req.Status
		return tx.Save(&shipment).Error
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(res.NotFoundResponse("Shipment not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to update shipment status"))
	}

	event := ShipmentKafkaEvent{
		EventType:    EventTypeStatusUpdate,
		TrackingCode: trackingCode,
		OldStatus:    oldStatus,
		NewStatus:    req.Status,
	}

	if err := s.sendToKafka(event, trackingCode); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Status updated, but failed to send event"))
	}

	return c.Status(fiber.StatusOK).JSON(res.SuccessResponse(fiber.StatusOK, "Shipment status updated successfully", map[string]interface{}{
		"tracking_code": trackingCode,
		"old_status":    oldStatus,
		"new_status":    req.Status,
		"message":       "Status updated successfully, notification will be sent shortly",
	}))
}


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