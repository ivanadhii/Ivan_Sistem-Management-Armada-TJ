package services

import (
	"context"

	"github.com/ivanadhi/transjakarta-fleet/internal/models"
)

type LocationService interface {
	SaveLocation(ctx context.Context, location *models.VehicleLocation) error
	GetLatestLocation(ctx context.Context, vehicleID string) (*models.VehicleLocation, error)
	GetLocationHistory(ctx context.Context, vehicleID string, startTime, endTime int64) ([]*models.VehicleLocation, error)
}