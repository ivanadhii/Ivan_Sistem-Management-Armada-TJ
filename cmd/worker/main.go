package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/config"
	"github.com/ivanadhi/transjakarta-fleet/internal/database"
	"github.com/ivanadhi/transjakarta-fleet/internal/rabbitmq"
	"github.com/ivanadhi/transjakarta-fleet/internal/repositories"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize database connection
	db, err := database.NewConnection(&cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories (for potential future use)
	geofenceEventRepo := repositories.NewGeofenceEventRepository(db.Pool, logger)
	_ = geofenceEventRepo // Placeholder for now

	// Initialize RabbitMQ client
	rabbitClient, err := rabbitmq.NewClient(&cfg.RabbitMQ, logger)
	if err != nil {
		logger.Fatal("Failed to create RabbitMQ client", zap.Error(err))
	}
	defer rabbitClient.Close()

	// Initialize RabbitMQ consumer
	consumer := rabbitmq.NewConsumer(rabbitClient, logger)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WaitGroup for goroutines
	var wg sync.WaitGroup

	// Start RabbitMQ consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := consumer.StartConsuming(ctx, handleGeofenceEvent); err != nil {
			logger.Error("RabbitMQ consumer error", zap.Error(err))
		}
	}()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutting down worker...")
		cancel()
	}()

	logger.Info("Geofence worker started successfully")

	// Wait for all goroutines to finish
	wg.Wait()
	logger.Info("Worker stopped gracefully")
}

func handleGeofenceEvent(ctx context.Context, message *rabbitmq.GeofenceEventMessage) error {
	// Get logger from context or create a new one
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Log the geofence event
	logger.Info("🚌 GEOFENCE ALERT! Vehicle entered landmark area!", 
		zap.String("vehicle_id", message.VehicleID),
		zap.String("landmark", message.GeofenceName),
		zap.String("event_type", message.Event),
		zap.Float64("latitude", message.Location.Latitude),
		zap.Float64("longitude", message.Location.Longitude),
		zap.Float64("distance_meters", message.Distance),
		zap.Time("timestamp", time.Unix(message.Timestamp, 0)))

	// Here you can add more processing logic:
	// - Send notifications to monitoring systems
	// - Update real-time dashboards
	// - Trigger alerts for specific vehicles/areas
	// - Send push notifications to mobile apps
	// - Update passenger information systems
	
	// Example: Process different event types
	switch message.Event {
	case "geofence_entry":
		processGeofenceEntry(message, logger)
	case "geofence_exit":
		processGeofenceExit(message, logger)
	default:
		logger.Warn("Unknown geofence event type", zap.String("event", message.Event))
	}

	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	return nil
}

func processGeofenceEntry(message *rabbitmq.GeofenceEventMessage, logger *zap.Logger) {
	// Business logic for geofence entry
	logger.Info("Processing geofence entry", 
		zap.String("vehicle_id", message.VehicleID),
		zap.String("geofence_name", message.GeofenceName))

	// Examples of what you might do:
	// 1. Update passenger info systems
	// 2. Send arrival notifications
	// 3. Update ETA calculations
	// 4. Trigger automatic announcements
	// 5. Update real-time maps

	// For demo purposes, we'll just log important landmarks
	switch message.GeofenceName {
	case "Monas":
		logger.Info("🏛️ Vehicle is now at the National Monument area!")
	case "Bundaran HI":
		logger.Info("🎡 Vehicle is now at Hotel Indonesia Roundabout!")
	case "Grand Indonesia":
		logger.Info("🏢 Vehicle is now at Grand Indonesia shopping area!")
	case "Plaza Indonesia":
		logger.Info("🏬 Vehicle is now at Plaza Indonesia area!")
	case "Sarinah":
		logger.Info("🏪 Vehicle is now at Sarinah area!")
	}
}

func processGeofenceExit(message *rabbitmq.GeofenceEventMessage, logger *zap.Logger) {
	// Business logic for geofence exit
	logger.Info("Processing geofence exit", 
		zap.String("vehicle_id", message.VehicleID),
		zap.String("geofence_name", message.GeofenceName))

	// Examples:
	// 1. Update departure times
	// 2. Calculate time spent in area
	// 3. Update route progress
}