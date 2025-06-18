package main

import (
	_"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	_ "github.com/Dziqha/logistics-delivery-tracker/helpers"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/controllers"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/routes"
	"github.com/Dziqha/logistics-delivery-tracker/services/location"
	"github.com/Dziqha/logistics-delivery-tracker/services/notification"
	"github.com/Dziqha/logistics-delivery-tracker/services/tracking"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env/api.env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "Logistics Delivery Tracker",
		ErrorHandler: customErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	db := configs.DatabaseConnection()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	configs.InitRabbitMQ()
	defer configs.CloseRabbitMQ()

	log.Println("🔄 Running database migrations...")
	if err := models.AdminMigrate(db); err != nil {
		log.Printf("❌ Admin migration failed: %v", err)
	}
	if err := models.ShipmentMigrate(db); err != nil {
		log.Printf("❌ Shipment migration failed: %v", err)
	}
	if err := models.LocationMigrate(db); err != nil {
		log.Printf("❌ Location migration failed: %v", err)
	}
	if err := models.NotificationMigrate(db); err != nil {
		log.Printf("❌ Notification migration failed: %v", err)
	}
	log.Println("✅ Database migrations completed")

	controllerAdmin := controllers.NewAdminController()
	controllerShipment := controllers.NewShipmentController()
	controllerLocation := controllers.NeLocationController()
	controllerNotification := controllers.NewNotificationController()

	routes.NewAdminRoutes(app, controllerAdmin)
	routes.NewShipmentRoutes(app, controllerShipment)
	routes.NewLocationRoutes(app, controllerLocation)
	routes.NewNotificationRoutes(app, controllerNotification)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Logistics Delivery Tracker is running",
			"version": "1.0.0",
		})
	})

	log.Println("🚀 Starting background services...")
	go tracking.ProcessTracking()
	go notification.NotificationProcessor()
	go location.ProcessLocationUpdates()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("🛑 Gracefully shutting down...")
		configs.CloseRabbitMQ()
		app.Shutdown()
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🌟 Server starting on port %s", port)
	log.Printf("🔗 Health check: http://localhost:%s/health", port)
	log.Printf("📋 Admin endpoints: http://localhost:%s/admin/*", port)
	log.Printf("📦 Shipment endpoints: http://localhost:%s/shipment/*", port)
	log.Printf("📍 Location endpoints: http://localhost:%s/location/*", port)
	log.Printf("🔔 Notification endpoints: http://localhost:%s/notification/*", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	log.Printf("❌ Error: %v", err)

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"code":    code,
		"message": message,
	})
}