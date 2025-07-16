package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/models"
	"github.com/ivanadhi/transjakarta-fleet/internal/repositories"
)

// Enhanced location service with geofencing
type enhancedLocationService struct {
	vehicleLocationRepo repositories.VehicleLocationRepository
	geofenceService     GeofenceService
	logger              *zap.Logger
}

func NewEnhancedLocationService(
	vehicleLocationRepo repositories.VehicleLocationRepository,
	geofenceService GeofenceService,
	logger *zap.Logger,
) LocationService {
	return &enhancedLocationService{
		vehicleLocationRepo: vehicleLocationRepo,
		geofenceService:     geofenceService,
		logger:              logger,
	}
}

func (s *enhancedLocationService) SaveLocation(ctx context.Context, location *models.VehicleLocation) error {
	// Validate location data
	if location.VehicleID == "" {
		return fmt.Errorf("vehicle_id is required")
	}
	
	if location.Latitude < -90 || location.Latitude > 90 {
		return fmt.Errorf("invalid latitude: %f", location.Latitude)
	}
	
	if location.Longitude < -180 || location.Longitude > 180 {
		return fmt.Errorf("invalid longitude: %f", location.Longitude)
	}
	
	if location.Timestamp <= 0 {
		return fmt.Errorf("invalid timestamp: %d", location.Timestamp)
	}

	// Save to database
	if err := s.vehicleLocationRepo.Create(ctx, location); err != nil {
		s.logger.Error("Failed to save vehicle location", 
			zap.Error(err),
			zap.String("vehicle_id", location.VehicleID))
		return fmt.Errorf("failed to save location: %w", err)
	}

	// Process for geofencing (async to avoid blocking location save)
	go func() {
		// Create new context for async operation
		asyncCtx := context.Background()
		if err := s.geofenceService.ProcessLocationForGeofencing(asyncCtx, location); err != nil {
			s.logger.Error("Failed to process geofencing", 
				zap.Error(err),
				zap.String("vehicle_id", location.VehicleID))
		}
	}()

	s.logger.Info("Vehicle location saved successfully", 
		zap.String("vehicle_id", location.VehicleID),
		zap.Float64("latitude", location.Latitude),
		zap.Float64("longitude", location.Longitude))

	return nil
}

func (s *enhancedLocationService) GetLatestLocation(ctx context.Context, vehicleID string) (*models.VehicleLocation, error) {
	if vehicleID == "" {
		return nil, fmt.Errorf("vehicle_id is required")
	}

	location, err := s.vehicleLocationRepo.GetLatestByVehicleID(ctx, vehicleID)
	if err != nil {
		s.logger.Error("Failed to get latest vehicle location", 
			zap.Error(err),
			zap.String("vehicle_id", vehicleID))
		return nil, fmt.Errorf("failed to get latest location: %w", err)
	}

	return location, nil
}

func (s *enhancedLocationService) GetLocationHistory(ctx context.Context, vehicleID string, startTime, endTime int64) ([]*models.VehicleLocation, error) {
	if vehicleID == "" {
		return nil, fmt.Errorf("vehicle_id is required")
	}
	
	if startTime <= 0 || endTime <= 0 {
		return nil, fmt.Errorf("invalid time range")
	}
	
	if startTime >= endTime {
		return nil, fmt.Errorf("start_time must be less than end_time")
	}

	locations, err := s.vehicleLocationRepo.GetHistoryByVehicleID(ctx, vehicleID, startTime, endTime)
	if err != nil {
		s.logger.Error("Failed to get vehicle location history", 
			zap.Error(err),
			zap.String("vehicle_id", vehicleID))
		return nil, fmt.Errorf("failed to get location history: %w", err)
	}

	return locations, nil
}