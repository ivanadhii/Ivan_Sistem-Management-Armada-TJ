package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Consumer struct {
	client *Client
	logger *zap.Logger
}

type MessageHandler func(ctx context.Context, message *GeofenceEventMessage) error

func NewConsumer(client *Client, logger *zap.Logger) *Consumer {
	return &Consumer{
		client: client,
		logger: logger,
	}
}

func (c *Consumer) StartConsuming(ctx context.Context, handler MessageHandler) error {
	if !c.client.IsConnected() {
		return fmt.Errorf("RabbitMQ client not connected")
	}
	
	// Start consuming messages
	messages, err := c.client.channel.Consume(
		c.client.config.Queue, // queue
		"",                    // consumer tag
		false,                 // auto-ack (we'll ack manually)
		false,                 // exclusive
		false,                 // no-local
		false,                 // no-wait
		nil,                   // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}
	
	c.logger.Info("Started consuming RabbitMQ messages", 
		zap.String("queue", c.client.config.Queue))
	
	// Process messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("RabbitMQ consumer stopping")
				return
			case delivery, ok := <-messages:
				if !ok {
					c.logger.Warn("RabbitMQ message channel closed")
					return
				}
				
				c.processMessage(ctx, delivery, handler)
			}
		}
	}()
	
	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

func (c *Consumer) processMessage(ctx context.Context, delivery amqp091.Delivery, handler MessageHandler) {
	// Parse message
	var message GeofenceEventMessage
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		c.logger.Error("Failed to parse RabbitMQ message", 
			zap.Error(err),
			zap.String("body", string(delivery.Body)))
		delivery.Nack(false, false) // Reject and don't requeue
		return
	}
	
	c.logger.Info("Processing geofence event", 
		zap.String("vehicle_id", message.VehicleID),
		zap.String("event", message.Event),
		zap.String("geofence_name", message.GeofenceName))
	
	// Handle message
	if err := handler(ctx, &message); err != nil {
		c.logger.Error("Failed to handle RabbitMQ message", 
			zap.Error(err),
			zap.String("vehicle_id", message.VehicleID))
		delivery.Nack(false, true) // Reject and requeue for retry
		return
	}
	
	// Acknowledge successful processing
	if err := delivery.Ack(false); err != nil {
		c.logger.Error("Failed to acknowledge RabbitMQ message", zap.Error(err))
	}
}