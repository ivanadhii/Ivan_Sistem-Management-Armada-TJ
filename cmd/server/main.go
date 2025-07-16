package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/config"
	"github.com/ivanadhi/transjakarta-fleet/internal/database"
	"github.com/ivanadhi/transjakarta-fleet/internal/handlers"
	"github.com/ivanadhi/transjakarta-fleet/internal/mqtt"
	"github.com/ivanadhi/transjakarta-fleet/internal/rabbitmq"
	"github.com/ivanadhi/transjakarta-fleet/internal/repositories"
	"github.com/ivanadhi/transjakarta-fleet/internal/services"
	"github.com/ivanadhi/transjakarta-fleet/pkg/geofence"
)

func main() {
	// Initialize logger
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer zapLogger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		zapLogger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize database
	db, err := database.NewConnection(&cfg.Database, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	ctx := context.Background()
	if err := db.RunMigrations(ctx); err != nil {
		zapLogger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize repositories
	vehicleLocationRepo := repositories.NewVehicleLocationRepository(db.Pool, zapLogger)
	geofenceRepo := repositories.NewGeofenceRepository(db.Pool, zapLogger)
	geofenceEventRepo := repositories.NewGeofenceEventRepository(db.Pool, zapLogger)

	// Initialize RabbitMQ client
	rabbitClient, err := rabbitmq.NewClient(&cfg.RabbitMQ, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to create RabbitMQ client", zap.Error(err))
	}
	defer rabbitClient.Close()

	// Initialize RabbitMQ publisher
	rabbitPublisher := rabbitmq.NewPublisher(rabbitClient, zapLogger)

	// Initialize geofence detector
	geofenceDetector := geofence.NewDetector(geofenceRepo, geofenceEventRepo, zapLogger)

	// Initialize services
	geofenceService := services.NewGeofenceService(geofenceDetector, rabbitPublisher, zapLogger)
	locationService := services.NewEnhancedLocationService(vehicleLocationRepo, geofenceService, zapLogger)

	// Initialize handlers
	vehicleHandler := handlers.NewVehicleHandler(locationService, zapLogger)
	geofenceHandler := handlers.NewGeofenceHandler(geofenceService, zapLogger)

	// Initialize MQTT client
	mqttClient, err := mqtt.NewClient(&cfg.MQTT, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to create MQTT client", zap.Error(err))
	}

	// Connect to MQTT broker
	if err := mqttClient.Connect(); err != nil {
		zapLogger.Fatal("Failed to connect to MQTT broker", zap.Error(err))
	}
	defer mqttClient.Disconnect()

	// Initialize MQTT subscriber
	locationSubscriber := mqtt.NewLocationSubscriber(mqttClient, locationService, zapLogger)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "TransJakarta Fleet Management API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorHandler: errorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Routes
	setupRoutes(app, vehicleHandler, geofenceHandler, db, mqttClient, rabbitClient)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WaitGroup for goroutines
	var wg sync.WaitGroup

	// Start MQTT subscriber
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := locationSubscriber.Start(ctx); err != nil {
			zapLogger.Error("MQTT subscriber error", zap.Error(err))
		}
	}()

	// Graceful shutdown handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		zapLogger.Info("Shutting down server...")
		
		// Cancel context to stop MQTT subscriber
		cancel()
		
		// Shutdown Fiber app
		if err := app.Shutdown(); err != nil {
			zapLogger.Error("Error during server shutdown", zap.Error(err))
		}
	}()

	// Start server
	address := cfg.Server.Host + ":" + cfg.Server.Port
	zapLogger.Info("Server starting with full geofencing integration", 
		zap.String("address", address),
		zap.String("mqtt_broker", cfg.MQTT.Broker+":"+cfg.MQTT.Port),
		zap.String("rabbitmq_url", cfg.RabbitMQ.URL))
	
	// Start HTTP server in goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.Listen(address); err != nil {
			zapLogger.Error("Server error", zap.Error(err))
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	zapLogger.Info("Server stopped gracefully")
}

func setupRoutes(app *fiber.App, vehicleHandler *handlers.VehicleHandler, geofenceHandler *handlers.GeofenceHandler, db *database.DB, mqttClient *mqtt.Client, rabbitClient *rabbitmq.Client) {
	// Health check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "TransJakarta Fleet Management API",
			"status":  "running",
			"version": "1.0.0",
			"features": []string{
				"Real-time Vehicle Tracking",
				"MQTT Integration",
				"Geofencing Detection",
				"RabbitMQ Event Processing",
			},
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		health := fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		}

		// Check database health
		if err := db.HealthCheck(ctx); err != nil {
			health["database"] = "disconnected"
			health["status"] = "unhealthy"
		} else {
			health["database"] = "connected"
		}

		// Check MQTT health
		if mqttClient.IsConnected() {
			health["mqtt"] = "connected"
		} else {
			health["mqtt"] = "disconnected"
			health["status"] = "degraded"
		}

		// Check RabbitMQ health
		if rabbitClient.IsConnected() {
			health["rabbitmq"] = "connected"
		} else {
			health["rabbitmq"] = "disconnected"
			health["status"] = "degraded"
		}

		statusCode := 200
		if health["status"] == "unhealthy" {
			statusCode = 503
		} else if health["status"] == "degraded" {
			statusCode = 206
		}

		return c.Status(statusCode).JSON(health)
	})

	// API routes
	api := app.Group("/api/v1")
	
	// Vehicle routes
	vehicles := api.Group("/vehicles")
	vehicles.Get("/:vehicle_id/location", vehicleHandler.GetLatestLocation)
	vehicles.Get("/:vehicle_id/history", vehicleHandler.GetLocationHistory)
	vehicles.Get("/:vehicle_id/geofence-events", geofenceHandler.GetGeofenceEvents)

	// System status routes
	api.Get("/mqtt/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"connected": mqttClient.IsConnected(),
			"timestamp": time.Now().UTC(),
		})
	})

	api.Get("/rabbitmq/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"connected": rabbitClient.IsConnected(),
			"timestamp": time.Now().UTC(),
		})
	})

	// Statistics endpoint
	api.Get("/stats", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		// Get database statistics
		var locationCount, eventCount int
		
		// Query total locations
		db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM vehicle_locations").Scan(&locationCount)
		
		// Query total geofence events
		db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM geofence_events").Scan(&eventCount)

		return c.JSON(fiber.Map{
			"total_locations":      locationCount,
			"total_geofence_events": eventCount,
			"timestamp":            time.Now().UTC(),
		})
	})
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
		"code":  code,
	})
}