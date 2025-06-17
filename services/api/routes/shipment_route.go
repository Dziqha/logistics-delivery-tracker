package routes

import (

	"github.com/Dziqha/logistics-delivery-tracker/services/api/controllers"
	"github.com/gofiber/fiber/v2"
)

func NewShipmentRoutes(route fiber.Router, controller *controllers.ShipmentController) {
	app := route.Group("/shipment")
	app.Post("/create", controller.CreateShipment)
	app.Put("/updateStatus/:tracking_code", controller.UpdateShipmentStatus)
}