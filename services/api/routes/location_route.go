package routes

import (
	"github.com/Dziqha/logistics-delivery-tracker/services/api/controllers"
	"github.com/gofiber/fiber/v2"
)

func NewLocationRoutes(route fiber.Router, controller *controllers.LocationController) {
	app := route.Group("/location")
	app.Post("/update", controller.UpdateLocation)
}