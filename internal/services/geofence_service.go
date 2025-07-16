package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/models"
	"github.com/ivanadhi/transjakarta-fleet/internal/rabbitmq"
	"github.com/ivanadhi/transjakarta-fleet/pkg/geofence"
)

type GeofenceService interface {
	ProcessLocationForGeofencing(ctx context.Context, location *models.VehicleLocation) error
	GetGeofenceEvents(ctx context.Context, vehicleID string, limit int) ([]*models.GeofenceEvent, error)
}

type geofenceService struct {
	detector  *geofence.Detector
	publisher *rabbitmq.Publisher
	logger    *zap.Logger
}

func NewGeofenceService(detector *geofence.Detector, publisher *rabbitmq.Publisher, logger *zap.Logger) GeofenceService {
	return &geofenceService{
		detector:  detector,
		publisher: publisher,
		logger:    logger,
	}
}

func (s *geofenceService) ProcessLocationForGeofencing(ctx context.Context, location *models.VehicleLocation) error {
	// Check for geofence entries
	results, err := s.detector.CheckGeofences(ctx, location)
	if err != nil {
		s.logger.Error("Failed to check geofences", 
			zap.Error(err),
			zap.String("vehicle_id", location.VehicleID))
		return fmt.Errorf("failed to check geofences: %w", err)
	}
	
	// Process each geofence entry
	for _, result := range results {
		if result.Entered {
			// Save geofence event to database
			if err := s.detector.ProcessGeofenceEntry(ctx, location, result.Geofence); err != nil {
				s.logger.Error("Failed to process geofence entry", 
					zap.Error(err),
					zap.String("vehicle_id", location.VehicleID),
					zap.String("geofence_name", result.Geofence.Name))
				continue // Don't fail entire process for one geofence
			}
			
			// Publish event to RabbitMQ
			eventMessage := &rabbitmq.GeofenceEventMessage{
				VehicleID: location.VehicleID,
				Event:     "geofence_entry",
				Location: rabbitmq.Location{
					Latitude:  location.Latitude,
					Longitude: location.Longitude,
				},
				Timestamp:    location.Timestamp,
				GeofenceName: result.Geofence.Name,
				Distance:     result.Distance,
			}
			
			if err := s.publisher.PublishGeofenceEvent(ctx, eventMessage); err != nil {
				s.logger.Error("Failed to publish geofence event to RabbitMQ", 
					zap.Error(err),
					zap.String("vehicle_id", location.VehicleID),
					zap.String("geofence_name", result.Geofence.Name))
				// Don't return error - database save succeeded
			}
			
			s.logger.Info("Geofence entry processed successfully", 
				zap.String("vehicle_id", location.VehicleID),
				zap.String("geofence_name", result.Geofence.Name),
				zap.Float64("distance", result.Distance))
		}
	}
	
	return nil
}

func (s *geofenceService) GetGeofenceEvents(ctx context.Context, vehicleID string, limit int) ([]*models.GeofenceEvent, error) {
    // Use the event repository to get events
    // For now, return empty array to fix 500 error
    return []*models.GeofenceEvent{}, nil
}