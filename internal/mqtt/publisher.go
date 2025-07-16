package mqtt

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

type LocationPublisher struct {
	client *Client
	logger *zap.Logger
}

func NewLocationPublisher(client *Client, logger *zap.Logger) *LocationPublisher {
	return &LocationPublisher{
		client: client,
		logger: logger,
	}
}

func (p *LocationPublisher) PublishLocation(vehicleID string, latitude, longitude float64, timestamp int64) error {
	if !p.client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}
	
	// Create location message
	message := LocationMessage{
		VehicleID: vehicleID,
		Latitude:  latitude,
		Longitude: longitude,
		Timestamp: timestamp,
	}
	
	// Convert to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal location message: %w", err)
	}
	
	// Create topic
	topic := fmt.Sprintf("/fleet/vehicle/%s/location", vehicleID)
	
	// Publish message
	qos := byte(1) // At least once delivery
	retained := false
	
	if err := p.client.Publish(topic, qos, retained, payload); err != nil {
		p.logger.Error("Failed to publish location", 
			zap.Error(err),
			zap.String("vehicle_id", vehicleID))
		return err
	}
	
	p.logger.Debug("Location published successfully", 
		zap.String("vehicle_id", vehicleID),
		zap.String("topic", topic))
	
	return nil
}