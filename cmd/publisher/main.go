package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/config"
	"github.com/ivanadhi/transjakarta-fleet/internal/mqtt"
)

// Jakarta area bounds
const (
	JakartaCenterLat = -6.2088  // Monas
	JakartaCenterLng = 106.8456 // Monas
	MaxRadiusKm      = 25       // 25km radius from center
)

type Vehicle struct {
	ID        string
	Latitude  float64
	Longitude float64
	Speed     float64 // km/h
	Direction float64 // degrees
}

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize MQTT client
	mqttClient, err := mqtt.NewClient(&cfg.MQTT, logger)
	if err != nil {
		logger.Fatal("Failed to create MQTT client", zap.Error(err))
	}

	// Connect to MQTT broker
	if err := mqttClient.Connect(); err != nil {
		logger.Fatal("Failed to connect to MQTT broker", zap.Error(err))
	}
	defer mqttClient.Disconnect()

	// Initialize publisher
	publisher := mqtt.NewLocationPublisher(mqttClient, logger)

	// Create fleet of vehicles
	vehicles := createVehicleFleet(5) // 5 vehicles for simulation

	logger.Info("Starting vehicle location publisher", 
		zap.Int("vehicle_count", len(vehicles)))

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start publishing routine
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				publishVehicleLocations(publisher, vehicles, logger)
			case <-c:
				logger.Info("Shutting down publisher...")
				return
			}
		}
	}()

	// Wait for shutdown signal
	<-c
	logger.Info("Publisher stopped")
}

func createVehicleFleet(count int) []*Vehicle {
	vehicles := make([]*Vehicle, count)
	
	// Predefined vehicle IDs (TransJakarta style)
	vehicleIDs := []string{
		"B1234XYZ", "B5678ABC", "B9012DEF", 
		"B3456GHI", "B7890JKL", "B2468MNO",
		"B1357PQR", "B8642STU", "B9753VWX",
		"B4681YZA",
	}
	
	for i := 0; i < count; i++ {
		// Use predefined ID or generate one
		vehicleID := vehicleIDs[i%len(vehicleIDs)]
		if i >= len(vehicleIDs) {
			vehicleID = fmt.Sprintf("B%04d%s", i, "XYZ")
		}
		
		vehicles[i] = &Vehicle{
			ID:        vehicleID,
			Latitude:  generateRandomLatitude(),
			Longitude: generateRandomLongitude(),
			Speed:     15 + rand.Float64()*20, // 15-35 km/h
			Direction: rand.Float64() * 360,   // Random initial direction
		}
	}
	
	return vehicles
}

func publishVehicleLocations(publisher *mqtt.LocationPublisher, vehicles []*Vehicle, logger *zap.Logger) {
	for _, vehicle := range vehicles {
		// Update vehicle position (simulate movement)
		updateVehiclePosition(vehicle)
		
		// Get current timestamp
		timestamp := time.Now().Unix()
		
		// Publish location
		err := publisher.PublishLocation(
			vehicle.ID,
			vehicle.Latitude,
			vehicle.Longitude,
			timestamp,
		)
		
		if err != nil {
			logger.Error("Failed to publish vehicle location", 
				zap.Error(err),
				zap.String("vehicle_id", vehicle.ID))
		} else {
			logger.Info("Published vehicle location", 
				zap.String("vehicle_id", vehicle.ID),
				zap.Float64("latitude", vehicle.Latitude),
				zap.Float64("longitude", vehicle.Longitude))
		}
	}
}

func updateVehiclePosition(vehicle *Vehicle) {
	// Simulate realistic vehicle movement
	// Speed in km/h converted to degrees per 2 seconds
	
	// Random speed variation (±5 km/h)
	speedVariation := (rand.Float64() - 0.5) * 10
	currentSpeed := math.Max(5, math.Min(50, vehicle.Speed+speedVariation))
	
	// Random direction change (±30 degrees)
	directionChange := (rand.Float64() - 0.5) * 60
	vehicle.Direction += directionChange
	
	// Normalize direction to 0-360
	for vehicle.Direction < 0 {
		vehicle.Direction += 360
	}
	for vehicle.Direction >= 360 {
		vehicle.Direction -= 360
	}
	
	// Convert speed to distance in 2 seconds
	distanceKm := currentSpeed * (2.0 / 3600.0) // km in 2 seconds
	
	// Convert to degrees (approximate)
	latChange := distanceKm * math.Cos(vehicle.Direction*math.Pi/180) / 111.0  // 1 degree lat ≈ 111 km
	lngChange := distanceKm * math.Sin(vehicle.Direction*math.Pi/180) / (111.0 * math.Cos(vehicle.Latitude*math.Pi/180))
	
	// Update position
	newLat := vehicle.Latitude + latChange
	newLng := vehicle.Longitude + lngChange
	
	// Keep vehicle within Jakarta bounds
	if isWithinJakartaBounds(newLat, newLng) {
		vehicle.Latitude = newLat
		vehicle.Longitude = newLng
		vehicle.Speed = currentSpeed
	} else {
		// Reverse direction if hitting bounds
		vehicle.Direction += 180
		if vehicle.Direction >= 360 {
			vehicle.Direction -= 360
		}
	}
}

func generateRandomLatitude() float64 {
	// Jakarta latitude range: approximately -6.35 to -6.05
	return JakartaCenterLat + (rand.Float64()-0.5)*0.6
}

func generateRandomLongitude() float64 {
	// Jakarta longitude range: approximately 106.65 to 107.05
	return JakartaCenterLng + (rand.Float64()-0.5)*0.8
}

func isWithinJakartaBounds(lat, lng float64) bool {
	// Check if coordinates are within MaxRadiusKm from Jakarta center
	distance := haversineDistance(JakartaCenterLat, JakartaCenterLng, lat, lng)
	return distance <= MaxRadiusKm
}

func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0
	
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180
	
	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad
	
	a := math.Sin(dlat/2)*math.Sin(dlat/2) + 
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return earthRadiusKm * c
}