package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Publisher struct {
	client *Client
	logger *zap.Logger
}

type GeofenceEventMessage struct {
	VehicleID string  `json:"vehicle_id"`
	Event     string  `json:"event"`
	Location  Location `json:"location"`
	Timestamp int64   `json:"timestamp"`
	GeofenceName string `json:"geofence_name,omitempty"`
	Distance     float64 `json:"distance,omitempty"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewPublisher(client *Client, logger *zap.Logger) *Publisher {
	return &Publisher{
		client: client,
		logger: logger,
	}
}

func (p *Publisher) PublishGeofenceEvent(ctx context.Context, message *GeofenceEventMessage) error {
	if !p.client.IsConnected() {
		return fmt.Errorf("RabbitMQ client not connected")
	}
	
	// Convert message to JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal geofence event: %w", err)
	}
	
	// Create routing key
	routingKey := fmt.Sprintf("geofence.%s", message.Event)
	
	// Create context with timeout
	pubCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	// Publish message
	err = p.client.channel.PublishWithContext(
		pubCtx,
		p.client.config.Exchange, // exchange
		routingKey,               // routing key
		false,                    // mandatory
		false,                    // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent, // Make message persistent
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	
	if err != nil {
		p.logger.Error("Failed to publish geofence event", 
			zap.Error(err),
			zap.String("vehicle_id", message.VehicleID),
			zap.String("event", message.Event))
		return fmt.Errorf("failed to publish geofence event: %w", err)
	}
	
	p.logger.Info("Geofence event published successfully", 
		zap.String("vehicle_id", message.VehicleID),
		zap.String("event", message.Event),
		zap.String("routing_key", routingKey))
	
	return nil
}