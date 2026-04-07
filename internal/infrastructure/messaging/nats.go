package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"skoolz/config"
	"skoolz/internal/logger"

	"github.com/nats-io/nats.go"
)

// NatsClient represents a NATS client
type NatsClient struct {
	config *config.NatsConfig
	logger *slog.Logger
	conn   *nats.Conn
	js     nats.JetStreamContext
}

// NatsMessage represents a NATS message
type NatsMessage struct {
	Subject   string                 `json:"subject"`
	Data      interface{}            `json:"data"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewNatsClient creates a new NATS client
func NewNatsClient(cfg *config.NatsConfig, log *slog.Logger) (*NatsClient, error) {
	// Connect to NATS
	conn, err := nats.Connect(cfg.URL, nats.Name(cfg.ClientID))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	startupLog := logger.NewStartupLogger(log)

	client := &NatsClient{
		config: cfg,
		logger: log,
		conn:   conn,
		js:     js,
	}

	startupLog.Messaging("NATS client ready", "url", cfg.URL, "clientID", cfg.ClientID)
	return client, nil
}

// PublishMessage publishes a message to NATS
func (n *NatsClient) PublishMessage(ctx context.Context, subject string, data interface{}, headers map[string]string) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message data: %w", err)
	}

	// Create NATS message
	msg := nats.NewMsg(subject)
	msg.Data = dataBytes

	// Publish message
	_, err = n.js.PublishMsg(msg)
	if err != nil {
		return fmt.Errorf("failed to publish message to NATS: %w", err)
	}

	n.logger.Info("Message published to NATS", "subject", subject, "size", len(dataBytes))

	return nil
}

// Subscribe subscribes to a NATS subject and processes messages
func (n *NatsClient) Subscribe(ctx context.Context, subject string, queueGroup string, handler func(NatsMessage) error) error {
	// Create subscription
	sub, err := n.js.QueueSubscribe(subject, queueGroup, func(msg *nats.Msg) {
		// Parse headers
		headers := make(map[string]string)
		if msg.Header != nil {
			for key, values := range msg.Header {
				if len(values) > 0 {
					headers[key] = values[0]
				}
			}
		}

		// Create NatsMessage
		natsMsg := NatsMessage{
			Subject:   msg.Subject,
			Data:      string(msg.Data),
			Headers:   headers,
			Timestamp: time.Now(),
		}

		// Process message
		if err := handler(natsMsg); err != nil {
			n.logger.Error("Error processing NATS message", "error", err.Error(), "subject", msg.Subject)
		}

		// Acknowledge message
		msg.Ack()
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to NATS subject %s: %w", subject, err)
	}

	startupLog := logger.NewStartupLogger(n.logger)
	startupLog.Subscription("NATS subscription active", "subject", subject)

	// Wait for context cancellation
	<-ctx.Done()

	// Unsubscribe
	sub.Unsubscribe()

	n.logger.Info("NATS subscription stopped", "subject", subject)

	return ctx.Err()
}

// Request sends a request and waits for a response
func (n *NatsClient) Request(ctx context.Context, subject string, data interface{}, timeout time.Duration) (*NatsMessage, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Send request
	msg, err := n.conn.RequestWithContext(ctx, subject, dataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to send NATS request: %w", err)
	}

	// Parse response headers
	headers := make(map[string]string)
	if msg.Header != nil {
		for key, values := range msg.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
	}

	// Create response message
	response := &NatsMessage{
		Subject:   msg.Subject,
		Data:      string(msg.Data),
		Headers:   headers,
		Timestamp: time.Now(),
	}

	return response, nil
}

// CreateStream creates a JetStream stream
func (n *NatsClient) CreateStream(name string, subjects []string) error {
	// Check if stream already exists
	stream, err := n.js.StreamInfo(name)
	if err == nil && stream != nil {
		startupLog := logger.NewStartupLogger(n.logger)
		startupLog.Stream("Stream exists", "name", name)
		return nil
	}

	// Create stream configuration
	streamConfig := &nats.StreamConfig{
		Name:     name,
		Subjects: subjects,
		Storage:  nats.FileStorage,
		MaxAge:   24 * time.Hour, // Keep messages for 24 hours
		MaxMsgs:  10000,          // Keep max 10k messages
	}

	// Create the stream
	_, err = n.js.AddStream(streamConfig)
	if err != nil {
		return fmt.Errorf("failed to create JetStream stream %s: %w", name, err)
	}

	startupLog := logger.NewStartupLogger(n.logger)
	startupLog.Stream("Stream created", "name", name)

	return nil
}

// Close closes the NATS client
func (n *NatsClient) Close() error {
	if n.conn != nil {
		n.conn.Close()
	}

	n.logger.Info("NATS client closed successfully")
	return nil
}

// HealthCheck checks the health of the NATS connection
func (n *NatsClient) HealthCheck() error {
	if n.conn == nil {
		return fmt.Errorf("NATS connection is nil")
	}

	if !n.conn.IsConnected() {
		return fmt.Errorf("NATS connection is not connected")
	}

	return nil
}
