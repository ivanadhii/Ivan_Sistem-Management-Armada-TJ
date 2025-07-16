package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/config"
)

type Client struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  *config.RabbitMQConfig
	logger  *zap.Logger
}

func NewClient(cfg *config.RabbitMQConfig, logger *zap.Logger) (*Client, error) {
	// Connect to RabbitMQ
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	
	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create RabbitMQ channel: %w", err)
	}
	
	client := &Client{
		conn:    conn,
		channel: channel,
		config:  cfg,
		logger:  logger,
	}
	
	// Setup exchange and queue
	if err := client.setupInfrastructure(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to setup RabbitMQ infrastructure: %w", err)
	}
	
	logger.Info("RabbitMQ client connected successfully",
		zap.String("exchange", cfg.Exchange),
		zap.String("queue", cfg.Queue))
	
	return client, nil
}

func (c *Client) setupInfrastructure() error {
	// Declare exchange
	err := c.channel.ExchangeDeclare(
		c.config.Exchange, // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}
	
	// Declare queue
	_, err = c.channel.QueueDeclare(
		c.config.Queue, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	
	// Bind queue to exchange
	err = c.channel.QueueBind(
		c.config.Queue,    // queue name
		"geofence.*",      // routing key pattern
		c.config.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}
	
	c.logger.Info("RabbitMQ infrastructure setup complete")
	return nil
}

func (c *Client) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	c.logger.Info("RabbitMQ client closed")
}

func (c *Client) IsConnected() bool {
	return c.conn != nil && !c.conn.IsClosed()
}