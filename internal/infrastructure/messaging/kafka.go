package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"skoolz/config"

	"github.com/segmentio/kafka-go"
)

// KafkaClient represents a Kafka client
type KafkaClient struct {
	config  *config.KafkaConfig
	logger  *slog.Logger
	readers map[string]*kafka.Reader
	writer  *kafka.Writer
	mu      sync.RWMutex
}

// KafkaMessage represents a Kafka message
type KafkaMessage struct {
	Topic     string                 `json:"topic"`
	Key       string                 `json:"key"`
	Value     interface{}            `json:"value"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewKafkaClient creates a new Kafka client
func NewKafkaClient(cfg *config.KafkaConfig, log *slog.Logger) (*KafkaClient, error) {
	client := &KafkaClient{
		config:  cfg,
		logger:  log,
		readers: make(map[string]*kafka.Reader),
	}

	// Initialize Kafka writer
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		Logger:       kafka.LoggerFunc(func(msg string, args ...interface{}) {}), // No-op logger to suppress internal Kafka logs
	}

	client.writer = writer
	fmt.Println("Kafka client initialized successfully", "brokers", cfg.Brokers, "groupID", cfg.GroupID, "topics", cfg.Topics)

	return client, nil
}

// getOrCreateReader gets an existing reader for a topic or creates a new one
func (k *KafkaClient) getOrCreateReader(topic string) (*kafka.Reader, error) {
	k.mu.RLock()
	if reader, exists := k.readers[topic]; exists {
		defer k.mu.RUnlock()
		return reader, nil
	}
	k.mu.RUnlock()

	k.mu.Lock()
	defer k.mu.Unlock()

	// Double-check after acquiring write lock
	if reader, exists := k.readers[topic]; exists {
		return reader, nil
	}

	// Create new reader for this topic
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         k.config.Brokers,
		GroupID:         k.config.GroupID,
		Topic:           topic,
		MinBytes:        k.config.Consumer.MaxBytes,
		MaxBytes:        k.config.Consumer.MaxBytes,
		MaxWait:         time.Duration(k.config.Consumer.MaxWaitSeconds) * time.Second,
		ReadLagInterval: -1,
		Logger:          kafka.LoggerFunc(func(msg string, args ...interface{}) {}), // No-op logger to suppress internal Kafka logs
	})

	k.readers[topic] = reader
	fmt.Println("Created new Kafka reader", "topic", topic, "groupID", k.config.GroupID)

	return reader, nil
}

// PublishMessage publishes a message to Kafka
func (k *KafkaClient) PublishMessage(ctx context.Context, topic string, key string, value interface{}, headers map[string]string) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message value: %w", err)
	}

	// Create Kafka message
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: valueBytes,
	}

	// Publish message
	err = k.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}
	fmt.Println("Message published to Kafka", "topic", topic, "key", key, "size", len(valueBytes))
	// k.logger.Info("Message published to Kafka", "topic", topic, "key", key, "size", len(valueBytes))

	return nil
}

// Subscribe subscribes to a Kafka topic and processes messages
func (k *KafkaClient) Subscribe(ctx context.Context, topic string, handler func(KafkaMessage) error) error {
	reader, err := k.getOrCreateReader(topic)
	if err != nil {
		return fmt.Errorf("failed to get or create reader for topic %s: %w", topic, err)
	}
	fmt.Println("Starting Kafka subscription", "topic", topic, "groupID", k.config.GroupID)

	// Track consecutive timeouts for backoff
	consecutiveTimeouts := 0
	maxConsecutiveTimeouts := 5
	baseBackoff := 100 * time.Millisecond
	maxBackoff := 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Kafka subscription context cancelled", "topic", topic)
			// k.logger.Info("Kafka subscription context cancelled", "topic", topic)
			return ctx.Err()
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				// Check if it's a timeout error
				if errors.Is(err, context.DeadlineExceeded) ||
					(err.Error() != "" && (strings.Contains(err.Error(), "Request Timed Out") ||
						strings.Contains(err.Error(), "no messages received"))) {
					consecutiveTimeouts++

					// Apply exponential backoff if we have too many consecutive timeouts
					if consecutiveTimeouts >= maxConsecutiveTimeouts {
						backoff := time.Duration(consecutiveTimeouts-maxConsecutiveTimeouts+1) * baseBackoff
						if backoff > maxBackoff {
							backoff = maxBackoff
						}

						// k.logger.Debug("No messages available, applying backoff",
						// 	"topic", topic,
						// 	"consecutive_timeouts", consecutiveTimeouts,
						// 	"backoff_ms", backoff.Milliseconds())

						select {
						case <-ctx.Done():
							return ctx.Err()
						case <-time.After(backoff):
							continue
						}
					}

					// Log timeout only at debug level to reduce noise
					// k.logger.Debug("No messages received from Kafka",
					// 	"topic", topic,
					// 	"consecutive_timeouts", consecutiveTimeouts)
					continue
				}

				// Reset timeout counter for non-timeout errors
				consecutiveTimeouts = 0
				fmt.Println("Error reading Kafka message", "error", err.Error(), "topic", topic)
				// k.logger.Error("Error reading Kafka message", "error", err.Error(), "topic", topic)
				continue
			}

			// Reset timeout counter when we successfully read a message
			consecutiveTimeouts = 0

			// Parse headers
			headers := make(map[string]string)
			for _, header := range msg.Headers {
				headers[header.Key] = string(header.Value)
			}

			// Create KafkaMessage
			kafkaMsg := KafkaMessage{
				Topic:     msg.Topic,
				Key:       string(msg.Key),
				Value:     string(msg.Value),
				Headers:   headers,
				Timestamp: msg.Time,
			}

			// Process message
			if err := handler(kafkaMsg); err != nil {
				fmt.Println("Error processing Kafka message", "error", err.Error(), "topic", msg.Topic, "key", string(msg.Key))
				// k.logger.Error("Error processing Kafka message", "error", err.Error(), "topic", msg.Topic, "key", string(msg.Key))
			}
		}
	}
}

// Close closes the Kafka client
func (k *KafkaClient) Close() error {
	var errs []error

	// Close all readers
	k.mu.Lock()
	for topic, reader := range k.readers {
		if err := reader.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close reader for topic %s: %w", topic, err))
		}
	}
	k.mu.Unlock()

	if err := k.writer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close writer: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Kafka client: %v", errs)
	}

	fmt.Println("Kafka client closed successfully")
	// 		k.logger.Info("Kafka client closed successfully")
	return nil
}

// HealthCheck checks the health of the Kafka connection
func (k *KafkaClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to read from any available topic to check connectivity
	k.mu.RLock()
	if len(k.readers) == 0 {
		k.mu.RUnlock()
		// If no readers exist, try to create one for the default topic
		reader, err := k.getOrCreateReader(k.config.Topics.UserEvents)
		if err != nil {
			return fmt.Errorf("kafka health check failed: %w", err)
		}
		_, err = reader.ReadMessage(ctx)
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("kafka health check failed: %w", err)
		}
		return nil
	}

	// Try the first available reader
	var firstReader *kafka.Reader
	for _, reader := range k.readers {
		firstReader = reader
		break
	}
	k.mu.RUnlock()

	_, err := firstReader.ReadMessage(ctx)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("kafka health check failed: %w", err)
	}

	return nil
}
