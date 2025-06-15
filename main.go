package main

import (
	"fmt"
	"log"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/controllers"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/routes"
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


	db := configs.DatabaseConnection()
	models.AdminMigrate(db)
	models.ShipmentMigrate(db)
	models.LocationMigrate(db)
	controllerAdmin := controllers.NewAdminController()
	controllerShipment := controllers.NewShipmentController()
	routes.NewAdminRoutes(app, controllerAdmin)
	routes.NewShipmentRoutes(app, controllerShipment)
	app.Listen(":3000")
	fmt.Println("Server running on port 3000")
	
}