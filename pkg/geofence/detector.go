package geofence

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/models"
	"github.com/ivanadhi/transjakarta-fleet/internal/repositories"
)

type Detector struct {
	geofenceRepo repositories.GeofenceRepository
	eventRepo    repositories.GeofenceEventRepository
	logger       *zap.Logger
	cache        map[int64]*models.Geofence // Cache for geofences
	lastUpdated  time.Time
}

type GeofenceResult struct {
	Geofence *models.Geofence
	Distance float64
	Entered  bool
}

func NewDetector(geofenceRepo repositories.GeofenceRepository, eventRepo repositories.GeofenceEventRepository, logger *zap.Logger) *Detector {
	return &Detector{
		geofenceRepo: geofenceRepo,
		eventRepo:    eventRepo,
		logger:       logger,
		cache:        make(map[int64]*models.Geofence),
	}
}

func (d *Detector) CheckGeofences(ctx context.Context, location *models.VehicleLocation) ([]*GeofenceResult, error) {
	// Refresh cache if needed (every 5 minutes)
	if time.Since(d.lastUpdated) > 5*time.Minute {
		if err := d.refreshGeofenceCache(ctx); err != nil {
			d.logger.Error("Failed to refresh geofence cache", zap.Error(err))
			return nil, err
		}
	}
	
	var results []*GeofenceResult
	
	// Check each geofence
	for _, geofence := range d.cache {
		distance := HaversineDistance(
			location.Latitude, location.Longitude,
			geofence.Latitude, geofence.Longitude,
		)
		
		entered := distance <= float64(geofence.Radius)
		
		if entered {
			result := &GeofenceResult{
				Geofence: geofence,
				Distance: distance,
				Entered:  true,
			}
			results = append(results, result)
			
			d.logger.Info("Vehicle entered geofence", 
				zap.String("vehicle_id", location.VehicleID),
				zap.String("geofence_name", geofence.Name),
				zap.Float64("distance", distance),
				zap.Int("radius", geofence.Radius))
		}
	}
	
	return results, nil
}

func (d *Detector) ProcessGeofenceEntry(ctx context.Context, location *models.VehicleLocation, geofence *models.Geofence) error {
	// Create geofence event
	event := &models.GeofenceEvent{
		VehicleID:   location.VehicleID,
		GeofenceID:  &geofence.ID,
		EventType:   "geofence_entry",
		Latitude:    location.Latitude,
		Longitude:   location.Longitude,
		Timestamp:   location.Timestamp,
	}
	
	// Save event to database
	if err := d.eventRepo.Create(ctx, event); err != nil {
		d.logger.Error("Failed to create geofence event", 
			zap.Error(err),
			zap.String("vehicle_id", location.VehicleID),
			zap.String("geofence_name", geofence.Name))
		return err
	}
	
	d.logger.Info("Geofence event created", 
		zap.String("vehicle_id", location.VehicleID),
		zap.String("geofence_name", geofence.Name),
		zap.Int64("event_id", event.ID))
	
	return nil
}

func (d *Detector) refreshGeofenceCache(ctx context.Context) error {
	geofences, err := d.geofenceRepo.GetAll(ctx)
	if err != nil {
		return err
	}
	
	// Update cache
	newCache := make(map[int64]*models.Geofence)
	for _, geofence := range geofences {
		newCache[geofence.ID] = geofence
	}
	
	d.cache = newCache
	d.lastUpdated = time.Now()
	
	d.logger.Info("Geofence cache refreshed", 
		zap.Int("geofence_count", len(geofences)))
	
	return nil
}
