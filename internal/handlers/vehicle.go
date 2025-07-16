package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/services"
)

type VehicleHandler struct {
	locationService services.LocationService
	logger          *zap.Logger
}

func NewVehicleHandler(locationService services.LocationService, logger *zap.Logger) *VehicleHandler {
	return &VehicleHandler{
		locationService: locationService,
		logger:          logger,
	}
}

func (h *VehicleHandler) GetLatestLocation(c *fiber.Ctx) error {
	vehicleID := c.Params("vehicle_id")
	if vehicleID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "vehicle_id is required",
		})
	}

	ctx := c.Context()
	location, err := h.locationService.GetLatestLocation(ctx, vehicleID)
	if err != nil {
		h.logger.Error("Failed to get latest location", 
			zap.Error(err),
			zap.String("vehicle_id", vehicleID))
		return c.Status(404).JSON(fiber.Map{
			"error": "Vehicle location not found",
		})
	}

	return c.JSON(fiber.Map{
		"vehicle_id": location.VehicleID,
		"latitude":   location.Latitude,
		"longitude":  location.Longitude,
		"timestamp":  location.Timestamp,
	})
}

func (h *VehicleHandler) GetLocationHistory(c *fiber.Ctx) error {
	vehicleID := c.Params("vehicle_id")
	if vehicleID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "vehicle_id is required",
		})
	}

	// Parse query parameters
	startTimeStr := c.Query("start")
	endTimeStr := c.Query("end")

	if startTimeStr == "" || endTimeStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "start and end time parameters are required",
		})
	}

	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid start time format",
		})
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid end time format",
		})
	}

	ctx := c.Context()
	locations, err := h.locationService.GetLocationHistory(ctx, vehicleID, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get location history", 
			zap.Error(err),
			zap.String("vehicle_id", vehicleID))
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get location history",
		})
	}

	// Transform response
	result := make([]fiber.Map, len(locations))
	for i, location := range locations {
		result[i] = fiber.Map{
			"vehicle_id": location.VehicleID,
			"latitude":   location.Latitude,
			"longitude":  location.Longitude,
			"timestamp":  location.Timestamp,
		}
	}

	return c.JSON(fiber.Map{
		"vehicle_id": vehicleID,
		"count":      len(result),
		"locations":  result,
	})
}