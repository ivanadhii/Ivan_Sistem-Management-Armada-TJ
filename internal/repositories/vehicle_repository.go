package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/models"
)

type vehicleLocationRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewVehicleLocationRepository(db *pgxpool.Pool, logger *zap.Logger) VehicleLocationRepository {
	return &vehicleLocationRepository{
		db:     db,
		logger: logger,
	}
}

func (r *vehicleLocationRepository) Create(ctx context.Context, location *models.VehicleLocation) error {
	query := `
		INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, timestamp)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query, 
		location.VehicleID, 
		location.Latitude, 
		location.Longitude, 
		location.Timestamp,
	).Scan(&location.ID, &location.CreatedAt)

	if err != nil {
		r.logger.Error("Failed to create vehicle location", 
			zap.Error(err),
			zap.String("vehicle_id", location.VehicleID))
		return fmt.Errorf("failed to create vehicle location: %w", err)
	}

	r.logger.Debug("Vehicle location created successfully", 
		zap.Int64("id", location.ID),
		zap.String("vehicle_id", location.VehicleID))

	return nil
}

func (r *vehicleLocationRepository) GetLatestByVehicleID(ctx context.Context, vehicleID string) (*models.VehicleLocation, error) {
	query := `
		SELECT id, vehicle_id, latitude, longitude, timestamp, created_at
		FROM vehicle_locations
		WHERE vehicle_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	location := &models.VehicleLocation{}
	err := r.db.QueryRow(ctx, query, vehicleID).Scan(
		&location.ID,
		&location.VehicleID,
		&location.Latitude,
		&location.Longitude,
		&location.Timestamp,
		&location.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to get latest vehicle location", 
			zap.Error(err),
			zap.String("vehicle_id", vehicleID))
		return nil, fmt.Errorf("failed to get latest location for vehicle %s: %w", vehicleID, err)
	}

	return location, nil
}

func (r *vehicleLocationRepository) GetHistoryByVehicleID(ctx context.Context, vehicleID string, startTime, endTime int64) ([]*models.VehicleLocation, error) {
	query := `
		SELECT id, vehicle_id, latitude, longitude, timestamp, created_at
		FROM vehicle_locations
		WHERE vehicle_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
	`

	rows, err := r.db.Query(ctx, query, vehicleID, startTime, endTime)
	if err != nil {
		r.logger.Error("Failed to get vehicle location history", 
			zap.Error(err),
			zap.String("vehicle_id", vehicleID))
		return nil, fmt.Errorf("failed to get location history for vehicle %s: %w", vehicleID, err)
	}
	defer rows.Close()

	var locations []*models.VehicleLocation
	for rows.Next() {
		location := &models.VehicleLocation{}
		err := rows.Scan(
			&location.ID,
			&location.VehicleID,
			&location.Latitude,
			&location.Longitude,
			&location.Timestamp,
			&location.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan vehicle location", zap.Error(err))
			return nil, fmt.Errorf("failed to scan location: %w", err)
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating location rows: %w", err)
	}

	return locations, nil
}