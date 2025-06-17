package main

import (
	"fmt"
	"log"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	_"github.com/Dziqha/logistics-delivery-tracker/helpers"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/controllers"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/routes"
	"github.com/Dziqha/logistics-delivery-tracker/services/location"
	"github.com/Dziqha/logistics-delivery-tracker/services/notification"
	"github.com/Dziqha/logistics-delivery-tracker/services/tracking"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env/api.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	app := fiber.New(
		fiber.Config{
			AppName: "Logistics Delivery Tracker",
		},
	)
	// app.Use("/shipment", helpers.AuthMiddleware())

	db := configs.DatabaseConnection()
	configs.InitRabbitMQ()
	defer configs.CloseRabbitMQ()
	models.AdminMigrate(db)
	models.ShipmentMigrate(db)
	models.LocationMigrate(db)
	models.NotificationMigrate(db)
	controllerAdmin := controllers.NewAdminController()
	controllerShipment := controllers.NewShipmentController()
	controllerLocation := controllers.NeLocationController()
	routes.NewAdminRoutes(app, controllerAdmin)
	routes.NewShipmentRoutes(app, controllerShipment)
	routes.NewLocationRoutes(app, controllerLocation)
	go tracking.ProcessTracking()
	go notification.NotificationProcessor()
	go location.ProcessLocationUpdates()
	fmt.Println("Server running on port 3000")
	app.Listen(":3000")

	
}