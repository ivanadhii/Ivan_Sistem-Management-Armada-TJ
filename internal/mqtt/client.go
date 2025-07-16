package mqtt

import (
	"crypto/tls"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"

	"github.com/ivanadhi/transjakarta-fleet/internal/config"
)

type Client struct {
	client mqtt.Client
	logger *zap.Logger
	config *config.MQTTConfig
}

type MessageHandler func(topic string, payload []byte) error

func NewClient(cfg *config.MQTTConfig, logger *zap.Logger) (*Client, error) {
	// Create MQTT client options
	opts := mqtt.NewClientOptions()
	
	// Broker URL
	brokerURL := fmt.Sprintf("tcp://%s:%s", cfg.Broker, cfg.Port)
	opts.AddBroker(brokerURL)
	
	// Client ID (unique)
	clientID := fmt.Sprintf("transjakarta-fleet-%d", time.Now().Unix())
	opts.SetClientID(clientID)
	
	// Credentials (if provided)
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}
	
	// Connection options
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetCleanSession(true)
	
	// TLS Configuration (for production)
	if cfg.Username != "" && cfg.Password != "" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // For development only
		}
		opts.SetTLSConfig(tlsConfig)
	}
	
	// Connection callbacks
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logger.Error("MQTT connection lost", zap.Error(err))
	})
	
	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		logger.Info("MQTT reconnecting to broker")
	})
	
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		logger.Info("MQTT connected to broker", zap.String("broker", brokerURL))
	})
	
	// Create client
	client := mqtt.NewClient(opts)
	
	return &Client{
		client: client,
		logger: logger,
		config: cfg,
	}, nil
}

func (c *Client) Connect() error {
	c.logger.Info("Connecting to MQTT broker", 
		zap.String("broker", c.config.Broker),
		zap.String("port", c.config.Port))
	
	token := c.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	
	c.logger.Info("Successfully connected to MQTT broker")
	return nil
}

func (c *Client) Disconnect() {
	if c.client.IsConnected() {
		c.client.Disconnect(250)
		c.logger.Info("Disconnected from MQTT broker")
	}
}

func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

func (c *Client) Subscribe(topic string, qos byte, handler MessageHandler) error {
	if !c.client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}
	
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		c.logger.Debug("Received MQTT message", 
			zap.String("topic", msg.Topic()),
			zap.Int("payload_size", len(msg.Payload())))
		
		if err := handler(msg.Topic(), msg.Payload()); err != nil {
			c.logger.Error("Failed to handle MQTT message", 
				zap.Error(err),
				zap.String("topic", msg.Topic()))
		}
	}
	
	token := c.client.Subscribe(topic, qos, messageHandler)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
	}
	
	c.logger.Info("Subscribed to MQTT topic", 
		zap.String("topic", topic),
		zap.Uint8("qos", qos))
	
	return nil
}

func (c *Client) Publish(topic string, qos byte, retained bool, payload []byte) error {
	if !c.client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}
	
	token := c.client.Publish(topic, qos, retained, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, token.Error())
	}
	
	c.logger.Debug("Published MQTT message", 
		zap.String("topic", topic),
		zap.Int("payload_size", len(payload)))
	
	return nil
}