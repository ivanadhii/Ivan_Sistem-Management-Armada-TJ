package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/models"
	"github.com/ivanadhi/transjakarta-fleet/internal/services"
)

type LocationSubscriber struct {
	client          *Client
	locationService services.LocationService
	logger          *zap.Logger
	topicPattern    *regexp.Regexp
}

type LocationMessage struct {
	VehicleID string  `json:"vehicle_id" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	Timestamp int64   `json:"timestamp" validate:"required,min=1"`
}

func NewLocationSubscriber(client *Client, locationService services.LocationService, logger *zap.Logger) *LocationSubscriber {
	// Compile regex pattern for topic matching
	// Pattern: /fleet/vehicle/{vehicle_id}/location
	pattern := regexp.MustCompile(`^/fleet/vehicle/([^/]+)/location$`)
	
	return &LocationSubscriber{
		client:          client,
		locationService: locationService,
		logger:          logger,
		topicPattern:    pattern,
	}
}

func (s *LocationSubscriber) Start(ctx context.Context) error {
	if !s.client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}
	
	// Subscribe to wildcard topic pattern
	topic := "/fleet/vehicle/+/location"
	qos := byte(1) // At least once delivery
	
	err := s.client.Subscribe(topic, qos, s.handleLocationMessage)
	if err != nil {
		return fmt.Errorf("failed to subscribe to location topic: %w", err)
	}
	
	s.logger.Info("Location subscriber started", zap.String("topic_pattern", topic))
	
	// Keep subscriber running until context is cancelled
	<-ctx.Done()
	
	s.logger.Info("Location subscriber stopping")
	return nil
}

func (s *LocationSubscriber) handleLocationMessage(topic string, payload []byte) error {
	// Extract vehicle_id from topic
	vehicleID, err := s.extractVehicleIDFromTopic(topic)
	if err != nil {
		s.logger.Error("Failed to extract vehicle ID from topic", 
			zap.Error(err),
			zap.String("topic", topic))
		return err
	}
	
	// Parse JSON payload
	var locationMsg LocationMessage
	if err := json.Unmarshal(payload, &locationMsg); err != nil {
		s.logger.Error("Failed to parse location message", 
			zap.Error(err),
			zap.String("topic", topic),
			zap.String("payload", string(payload)))
		return fmt.Errorf("invalid JSON payload: %w", err)
	}
	
	// Validate that vehicle_id matches topic
	if locationMsg.VehicleID != vehicleID {
		s.logger.Warn("Vehicle ID mismatch between topic and payload", 
			zap.String("topic_vehicle_id", vehicleID),
			zap.String("payload_vehicle_id", locationMsg.VehicleID))
		// Use vehicle_id from topic as authoritative
		locationMsg.VehicleID = vehicleID
	}
	
	// Convert to domain model
	location := &models.VehicleLocation{
		VehicleID: locationMsg.VehicleID,
		Latitude:  locationMsg.Latitude,
		Longitude: locationMsg.Longitude,
		Timestamp: locationMsg.Timestamp,
	}
	
	// Save location via service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := s.locationService.SaveLocation(ctx, location); err != nil {
		s.logger.Error("Failed to save vehicle location", 
			zap.Error(err),
			zap.String("vehicle_id", location.VehicleID))
		return fmt.Errorf("failed to save location: %w", err)
	}
	
	s.logger.Info("Vehicle location saved successfully", 
		zap.String("vehicle_id", location.VehicleID),
		zap.Float64("latitude", location.Latitude),
		zap.Float64("longitude", location.Longitude),
		zap.Int64("timestamp", location.Timestamp))
	
	return nil
}

func (s *LocationSubscriber) extractVehicleIDFromTopic(topic string) (string, error) {
	matches := s.topicPattern.FindStringSubmatch(topic)
	if len(matches) != 2 {
		return "", fmt.Errorf("topic does not match expected pattern: %s", topic)
	}
	
	vehicleID := matches[1]
	if vehicleID == "" {
		return "", fmt.Errorf("empty vehicle_id in topic: %s", topic)
	}
	
	return vehicleID, nil
}