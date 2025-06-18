package routes

import (
	"github.com/Dziqha/logistics-delivery-tracker/services/api/controllers"
	"github.com/gofiber/fiber/v2"
)

func NewNotificationRoutes(route fiber.Router, controller *controllers.NotificationController) {
	app := route.Group("/notification")
	app.Get("/:tracking_code", controller.GetNotificationsByTrackingCode)
}