package controllers

import (
	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/res"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
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

	newShpment := models.Shipment{
		SenderName:    req.SenderName,
		ReceiverName:  req.ReceiverName,
		OriginAddress: req.OriginAddress,
		DestAddress:   req.DestAddress,
		Status:        req.Status,
	}

	err := configs.DatabaseConnection().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&newShpment).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		res.InternalServerErrorResponse("Failed to create shipment")
	}

	response := res.ShipmentResponseData(newShpment)
	return c.JSON(res.SuccessResponse(fiber.StatusCreated, "Shipment created successfully", response))

}
		