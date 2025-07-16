package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/services"
)

type GeofenceHandler struct {
	geofenceService services.GeofenceService
	logger          *zap.Logger
}

func NewGeofenceHandler(geofenceService services.GeofenceService, logger *zap.Logger) *GeofenceHandler {
	return &GeofenceHandler{
		geofenceService: geofenceService,
		logger:          logger,
	}
}

func (h *GeofenceHandler) GetGeofenceEvents(c *fiber.Ctx) error {
	vehicleID := c.Params("vehicle_id")
	if vehicleID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "vehicle_id is required",
		})
	}

	// Parse limit parameter
	limitStr := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Cap at 100 for performance
	}

	ctx := c.Context()
	events, err := h.geofenceService.GetGeofenceEvents(ctx, vehicleID, limit)
	if err != nil {
		h.logger.Error("Failed to get geofence events", 
			zap.Error(err),
			zap.String("vehicle_id", vehicleID))
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get geofence events",
		})
	}

	return c.JSON(fiber.Map{
		"vehicle_id": vehicleID,
		"count":      len(events),
		"events":     events,
	})
}