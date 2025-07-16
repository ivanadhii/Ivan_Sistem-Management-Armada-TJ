package repositories

import (
	"context"

	"github.com/ivanadhi/transjakarta-fleet/internal/models"
)

type VehicleLocationRepository interface {
	Create(ctx context.Context, location *models.VehicleLocation) error
	GetLatestByVehicleID(ctx context.Context, vehicleID string) (*models.VehicleLocation, error)
	GetHistoryByVehicleID(ctx context.Context, vehicleID string, startTime, endTime int64) ([]*models.VehicleLocation, error)
}

type GeofenceRepository interface {
	GetAll(ctx context.Context) ([]*models.Geofence, error)
	GetByID(ctx context.Context, id int64) (*models.Geofence, error)
}

type GeofenceEventRepository interface {
	Create(ctx context.Context, event *models.GeofenceEvent) error
	GetByVehicleID(ctx context.Context, vehicleID string, limit int) ([]*models.GeofenceEvent, error)
}

