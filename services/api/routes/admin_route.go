package routes

import (
	"github.com/Dziqha/logistics-delivery-tracker/services/api/controllers"
	"github.com/gofiber/fiber/v2"
)

func NewAdminRoutes(route fiber.Router, controller *controllers.AdminController) {
	app := route.Group("/admin")
	app.Post("/register", controller.RegisterAdmin)
	app.Post("/login", controller.LoginAdmin)
}