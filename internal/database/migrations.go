package database

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Migration struct {
	Version int
	Name    string
	SQL     string
}

var migrations = []Migration{
	{
		Version: 1,
		Name:    "create_vehicle_locations_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS vehicle_locations (
				id BIGSERIAL PRIMARY KEY,
				vehicle_id VARCHAR(50) NOT NULL,
				latitude DECIMAL(10, 8) NOT NULL,
				longitude DECIMAL(11, 8) NOT NULL,
				timestamp BIGINT NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
		`,
	},
	{
		Version: 2,
		Name:    "create_indexes_vehicle_locations",
		SQL: `
			CREATE INDEX IF NOT EXISTS idx_vehicle_locations_vehicle_id ON vehicle_locations(vehicle_id);
			CREATE INDEX IF NOT EXISTS idx_vehicle_locations_timestamp ON vehicle_locations(timestamp);
			CREATE INDEX IF NOT EXISTS idx_vehicle_locations_vehicle_timestamp ON vehicle_locations(vehicle_id, timestamp DESC);
		`,
	},
	{
		Version: 3,
		Name:    "create_geofences_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS geofences (
				id BIGSERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				latitude DECIMAL(10, 8) NOT NULL,
				longitude DECIMAL(11, 8) NOT NULL,
				radius INTEGER NOT NULL DEFAULT 50,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
		`,
	},
	{
		Version: 4,
		Name:    "insert_default_geofences",
		SQL: `
			INSERT INTO geofences (name, latitude, longitude, radius) VALUES
			('Monas', -6.1754, 106.8272, 50),
			('Bundaran HI', -6.1944, 106.8229, 50),
			('Grand Indonesia', -6.1944, 106.8229, 50),
			('Plaza Indonesia', -6.1928, 106.8218, 50),
			('Sarinah', -6.1922, 106.8219, 50)
			ON CONFLICT DO NOTHING;
		`,
	},
	{
		Version: 5,
		Name:    "create_geofence_events_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS geofence_events (
				id BIGSERIAL PRIMARY KEY,
				vehicle_id VARCHAR(50) NOT NULL,
				geofence_id BIGINT REFERENCES geofences(id),
				event_type VARCHAR(20) NOT NULL DEFAULT 'geofence_entry',
				latitude DECIMAL(10, 8) NOT NULL,
				longitude DECIMAL(11, 8) NOT NULL,
				timestamp BIGINT NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
		`,
	},
	{
		Version: 6,
		Name:    "create_schema_migrations_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
		`,
	},
}

func (db *DB) RunMigrations(ctx context.Context) error {
	db.Logger.Info("Starting database migrations")

	// Create migrations table first
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	rows, err := db.Pool.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	appliedMigrations := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedMigrations[version] = true
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if appliedMigrations[migration.Version] {
			db.Logger.Debug("Migration already applied", 
				zap.Int("version", migration.Version),
				zap.String("name", migration.Name))
			continue
		}

		db.Logger.Info("Applying migration", 
			zap.Int("version", migration.Version),
			zap.String("name", migration.Name))

		// Begin transaction
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", migration.Version, err)
		}

		// Execute migration
		_, err = tx.Exec(ctx, migration.SQL)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %d (%s): %w", migration.Version, migration.Name, err)
		}

		// Record migration
		_, err = tx.Exec(ctx, 
			"INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
			migration.Version, migration.Name)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		db.Logger.Info("Migration applied successfully", 
			zap.Int("version", migration.Version),
			zap.String("name", migration.Name))
	}

	db.Logger.Info("Database migrations completed successfully")
	return nil
}