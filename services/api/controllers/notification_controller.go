package controllers

import (
	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/res"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)


type NotificationController struct {}

func NewNotificationController() *NotificationController {
	return &NotificationController{}
}

func (n *NotificationController) GetNotificationsByTrackingCode(c *fiber.Ctx) error {
	var notifications []models.Notification
	var responseData []res.NotificationResponse

	trackingCode := c.Params("tracking_code")
	if trackingCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Tracking code is required"))
	}

	err := configs.DatabaseConnection().Transaction(func(tx *gorm.DB) error {
		// Ambil semua notifikasi yang memiliki shipment dengan tracking code yang diminta
		return tx.
			Joins("JOIN shipments ON shipments.id = notifications.shipment_id").
			Where("shipments.tracking_code = ?", trackingCode).
			Preload("Shipment").
			Find(&notifications).Error
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to fetch notifications from database"))
	}

	if len(notifications) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(res.NotFoundResponse("No notifications found for this tracking code"))
	}

	for _, notif := range notifications {
		responseData = append(responseData, res.NotificationResponse{
			ID:                  notif.ID,
			ShipmentID:          notif.ShipmentID,
			Type:                notif.Type,
			Title:               notif.Title,
			Message:             notif.Message,
			IsRead:              notif.IsRead,
			CreatedAt:           notif.CreatedAt,
			UpdatedAt:           notif.UpdatedAt,
			TrackingCodeShipment: notif.Shipment.TrackingCode,
		})
	}

	return c.JSON(res.SuccessResponse(fiber.StatusOK, "Successfully fetched notifications", responseData))
}
