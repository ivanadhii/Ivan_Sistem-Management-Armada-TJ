package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/models"
)

type geofenceRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

type geofenceEventRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewGeofenceRepository(db *pgxpool.Pool, logger *zap.Logger) GeofenceRepository {
	return &geofenceRepository{
		db:     db,
		logger: logger,
	}
}

func NewGeofenceEventRepository(db *pgxpool.Pool, logger *zap.Logger) GeofenceEventRepository {
	return &geofenceEventRepository{
		db:     db,
		logger: logger,
	}
}

func (r *geofenceRepository) GetAll(ctx context.Context) ([]*models.Geofence, error) {
	query := `
		SELECT id, name, latitude, longitude, radius, created_at, updated_at
		FROM geofences
		ORDER BY name
	`
	
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get all geofences", zap.Error(err))
		return nil, fmt.Errorf("failed to get geofences: %w", err)
	}
	defer rows.Close()
	
	var geofences []*models.Geofence
	for rows.Next() {
		geofence := &models.Geofence{}
		err := rows.Scan(
			&geofence.ID,
			&geofence.Name,
			&geofence.Latitude,
			&geofence.Longitude,
			&geofence.Radius,
			&geofence.CreatedAt,
			&geofence.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan geofence", zap.Error(err))
			return nil, fmt.Errorf("failed to scan geofence: %w", err)
		}
		geofences = append(geofences, geofence)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating geofence rows: %w", err)
	}
	
	return geofences, nil
}

func (r *geofenceRepository) GetByID(ctx context.Context, id int64) (*models.Geofence, error) {
	query := `
		SELECT id, name, latitude, longitude, radius, created_at, updated_at
		FROM geofences
		WHERE id = $1
	`
	
	geofence := &models.Geofence{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&geofence.ID,
		&geofence.Name,
		&geofence.Latitude,
		&geofence.Longitude,
		&geofence.Radius,
		&geofence.CreatedAt,
		&geofence.UpdatedAt,
	)
	
	if err != nil {
		r.logger.Error("Failed to get geofence by ID", 
			zap.Error(err),
			zap.Int64("geofence_id", id))
		return nil, fmt.Errorf("failed to get geofence %d: %w", id, err)
	}
	
	return geofence, nil
}

func (r *geofenceEventRepository) Create(ctx context.Context, event *models.GeofenceEvent) error {
	query := `
		INSERT INTO geofence_events (vehicle_id, geofence_id, event_type, latitude, longitude, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	
	err := r.db.QueryRow(ctx, query,
		event.VehicleID,
		event.GeofenceID,
		event.EventType,
		event.Latitude,
		event.Longitude,
		event.Timestamp,
	).Scan(&event.ID, &event.CreatedAt)
	
	if err != nil {
		r.logger.Error("Failed to create geofence event", 
			zap.Error(err),
			zap.String("vehicle_id", event.VehicleID))
		return fmt.Errorf("failed to create geofence event: %w", err)
	}
	
	r.logger.Debug("Geofence event created successfully", 
		zap.Int64("event_id", event.ID),
		zap.String("vehicle_id", event.VehicleID))
	
	return nil
}

func (r *geofenceEventRepository) GetByVehicleID(ctx context.Context, vehicleID string, limit int) ([]*models.GeofenceEvent, error) {
	query := `
		SELECT ge.id, ge.vehicle_id, ge.geofence_id, ge.event_type, 
		       ge.latitude, ge.longitude, ge.timestamp, ge.created_at
		FROM geofence_events ge
		WHERE ge.vehicle_id = $1
		ORDER BY ge.timestamp DESC
		LIMIT $2
	`
	
	rows, err := r.db.Query(ctx, query, vehicleID, limit)
	if err != nil {
		r.logger.Error("Failed to get geofence events by vehicle ID", 
			zap.Error(err),
			zap.String("vehicle_id", vehicleID))
		return nil, fmt.Errorf("failed to get geofence events for vehicle %s: %w", vehicleID, err)
	}
	defer rows.Close()
	
	var events []*models.GeofenceEvent
	for rows.Next() {
		event := &models.GeofenceEvent{}
		err := rows.Scan(
			&event.ID,
			&event.VehicleID,
			&event.GeofenceID,
			&event.EventType,
			&event.Latitude,
			&event.Longitude,
			&event.Timestamp,
			&event.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan geofence event", zap.Error(err))
			return nil, fmt.Errorf("failed to scan geofence event: %w", err)
		}
		events = append(events, event)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating geofence event rows: %w", err)
	}
	
	return events, nil
}